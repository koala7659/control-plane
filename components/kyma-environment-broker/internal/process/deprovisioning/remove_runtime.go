package deprovisioning

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/process"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/storage/dberr"
)

const (
	// the time after which the operation is marked as expired
	RemoveRuntimeTimeout = 1 * time.Hour
)

type RemoveRuntimeStep struct {
	operationManager  *process.DeprovisionOperationManager
	instanceStorage   storage.Instances
	provisionerClient provisioner.Client
}

func NewRemoveRuntimeStep(os storage.Operations, is storage.Instances, cli provisioner.Client) *RemoveRuntimeStep {
	return &RemoveRuntimeStep{
		operationManager:  process.NewDeprovisionOperationManager(os),
		instanceStorage:   is,
		provisionerClient: cli,
	}
}

func (s *RemoveRuntimeStep) Name() string {
	return "Remove_Runtime"
}

func (s *RemoveRuntimeStep) Run(operation internal.DeprovisioningOperation, log logrus.FieldLogger) (internal.DeprovisioningOperation, time.Duration, error) {
	if time.Since(operation.UpdatedAt) > RemoveRuntimeTimeout {
		log.Infof("operation has reached the time limit: updated operation time: %s", operation.UpdatedAt)
		return s.operationManager.OperationFailed(operation, fmt.Sprintf("operation has reached the time limit: %s", RemoveRuntimeTimeout))
	}

	instance, err := s.instanceStorage.GetByID(operation.InstanceID)
	switch {
	case err == nil:
	case dberr.IsNotFound(err):
		return s.operationManager.OperationSucceeded(operation, "instance already deprovisioned")
	default:
		log.Errorf("unable to get instance from storage: %s", err)
		return operation, 1 * time.Second, nil
	}

	if instance.RuntimeID == "" {
		// happens when provisioning process failed and Create_Runtime step was never reached
		log.Warnf("Runtime does not exist for instance id %q", instance.InstanceID)

		operation, _, _ := s.operationManager.OperationSucceeded(operation, "Runtime was never provisioned")
		// return repeat mode (1 sec) to start the initialization step which will finish process and remove instance
		return operation, 1 * time.Second, nil
	}
	log = log.WithField("runtimeID", instance.RuntimeID)

	var provisionerResponse string
	if operation.ProvisionerOperationID == "" {

		provisionerResponse, err = s.provisionerClient.DeprovisionRuntime(instance.GlobalAccountID, instance.RuntimeID)
		if err != nil {
			log.Errorf("unable to deprovision runtime: %s", err)
			return operation, 10 * time.Second, nil
		}
		operation.ProvisionerOperationID = provisionerResponse
		log.Infof("fetched ProvisionerOperationID=%s", provisionerResponse)

		operation, repeat, err := s.operationManager.UpdateOperation(operation)
		if repeat != 0 {
			log.Errorf("cannot save operation ID from provisioner: %s", err)
			return operation, 5 * time.Second, nil
		}
	}

	log.Infof("runtime deletion process initiated successfully")
	// return repeat mode (1 sec) to start the initialization step which will now check the runtime status
	return operation, 1 * time.Second, nil
}
