package provisioning

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/broker"

	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/avs"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/logger"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/process/provisioning/automock"
	provisionerAutomock "github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/provisioner/automock"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-project/control-plane/components/provisioner/pkg/gqlschema"

	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/stretchr/testify/assert"
)

const (
	statusOperationID            = "17f3ddba-1132-466d-a3c5-920f544d7ea6"
	statusInstanceID             = "9d75a545-2e1e-4786-abd8-a37b14e185b9"
	statusRuntimeID              = "ef4e3210-652c-453e-8015-bba1c1cd1e1c"
	statusGlobalAccountID        = "abf73c71-a653-4951-b9c2-a26d6c2cccbd"
	statusProvisionerOperationID = "e04de524-53b3-4890-b05a-296be393e4ba"

	dashboardURL = "http://runtime.com"
)

func TestInitialisationStep_RunInitialized(t *testing.T) {
	// given
	memoryStorage := storage.NewMemoryStorage()

	operation := fixOperationRuntimeStatus(t, broker.GCPPlanID)
	err := memoryStorage.Operations().InsertProvisioningOperation(operation)
	assert.NoError(t, err)

	instance := fixInstanceRuntimeStatus()
	err = memoryStorage.Instances().Insert(instance)
	assert.NoError(t, err)

	provisionerClient := &provisionerAutomock.Client{}
	provisionerClient.On("RuntimeOperationStatus", statusGlobalAccountID, statusProvisionerOperationID).Return(gqlschema.OperationStatus{
		ID:        ptr.String(statusProvisionerOperationID),
		Operation: "",
		State:     gqlschema.OperationStateSucceeded,
		Message:   nil,
		RuntimeID: nil,
	}, nil)

	directorClient := &automock.DirectorClient{}
	directorClient.On("GetConsoleURL", statusGlobalAccountID, statusRuntimeID).Return(dashboardURL, nil)

	idh := &idHolder{}
	mockOauthServer := newMockAvsOauthServer()
	defer mockOauthServer.Close()
	mockAvsServer := newMockAvsServer(t, idh, false)
	defer mockAvsServer.Close()
	avsConfig := avsConfig(mockOauthServer, mockAvsServer)
	avsClient, err := avs.NewClient(context.TODO(), avsConfig, logrus.New())
	assert.NoError(t, err)
	avsDel := avs.NewDelegator(avsClient, avsConfig, memoryStorage.Operations())
	externalEvalAssistant := avs.NewExternalEvalAssistant(avsConfig)
	externalEvalCreator := NewExternalEvalCreator(avsDel, false, externalEvalAssistant)
	iasType := NewIASType(nil, true)

	rvc := &automock.RuntimeVersionConfiguratorForProvisioning{}
	defer rvc.AssertExpectations(t)

	step := NewInitialisationStep(memoryStorage.Operations(), memoryStorage.Instances(), provisionerClient,
		directorClient, nil, externalEvalCreator, iasType, time.Hour, rvc, nil)

	// when
	operation, repeat, err := step.Run(operation, logger.NewLogDummy())

	// then
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), repeat)
	assert.Equal(t, domain.Succeeded, operation.State)

	updatedInstance, err := memoryStorage.Instances().GetByID(statusInstanceID)
	assert.NoError(t, err)
	assert.Equal(t, dashboardURL, updatedInstance.DashboardURL)

	assert.Equal(t, idh.id, operation.Avs.AVSEvaluationExternalId)
	inDB, err := memoryStorage.Operations().GetProvisioningOperationByID(operation.ID)
	assert.NoError(t, err)
	assert.Equal(t, inDB.Avs.AVSEvaluationExternalId, idh.id)
}

func TestInitialisationStep_RunUninitialized(t *testing.T) {
	// given
	memoryStorage := storage.NewMemoryStorage()

	operation := fixOperationRuntimeStatus(t, broker.GCPPlanID)
	err := memoryStorage.Operations().InsertProvisioningOperation(operation)
	assert.NoError(t, err)

	instance := fixInstanceRuntimeStatus()
	err = memoryStorage.Instances().Insert(instance)
	assert.NoError(t, err)

	provisionerClient := &provisionerAutomock.Client{}
	provisionerClient.On("RuntimeOperationStatus", statusGlobalAccountID, statusProvisionerOperationID).Return(gqlschema.OperationStatus{
		ID:        ptr.String(statusProvisionerOperationID),
		Operation: "",
		State:     gqlschema.OperationStateSucceeded,
		Message:   nil,
		RuntimeID: nil,
	}, nil)

	directorClient := &automock.DirectorClient{}
	directorClient.On("GetConsoleURL", statusGlobalAccountID, statusRuntimeID).Return(dashboardURL, nil)

	idh := &idHolder{}
	mockOauthServer := newMockAvsOauthServer()
	defer mockOauthServer.Close()
	mockAvsServer := newMockAvsServer(t, idh, false)
	defer mockAvsServer.Close()
	avsConfig := avsConfig(mockOauthServer, mockAvsServer)
	avsClient, err := avs.NewClient(context.TODO(), avsConfig, logger.NewLogDummy())
	assert.NoError(t, err)
	avsDel := avs.NewDelegator(avsClient, avsConfig, memoryStorage.Operations())
	externalEvalAssistant := avs.NewExternalEvalAssistant(avsConfig)
	externalEvalCreator := NewExternalEvalCreator(avsDel, false, externalEvalAssistant)
	iasType := NewIASType(nil, true)

	rvc := &automock.RuntimeVersionConfiguratorForProvisioning{}
	defer rvc.AssertExpectations(t)

	step := NewInitialisationStep(memoryStorage.Operations(), memoryStorage.Instances(), provisionerClient,
		directorClient, nil, externalEvalCreator, iasType, time.Hour, rvc, nil)

	// when
	operation, repeat, err := step.Run(operation, logger.NewLogDummy())

	// then
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), repeat)
	assert.Equal(t, domain.Succeeded, operation.State)

	updatedInstance, err := memoryStorage.Instances().GetByID(statusInstanceID)
	assert.NoError(t, err)
	assert.Equal(t, dashboardURL, updatedInstance.DashboardURL)

	assert.Equal(t, idh.id, operation.Avs.AVSEvaluationExternalId)
	inDB, err := memoryStorage.Operations().GetProvisioningOperationByID(operation.ID)
	assert.NoError(t, err)
	assert.Equal(t, inDB.Avs.AVSEvaluationExternalId, idh.id)
}

func fixOperationRuntimeStatus(t *testing.T, planId string) internal.ProvisioningOperation {
	return internal.ProvisioningOperation{
		Operation: internal.Operation{
			ID:                     statusOperationID,
			InstanceID:             statusInstanceID,
			ProvisionerOperationID: statusProvisionerOperationID,
			Description:            "",
			UpdatedAt:              time.Now(),
		},
		ProvisioningParameters: fixProvisioningParametersRuntimeStatus(t, planId),
	}
}

func fixOperationRuntimeStatusWithProvider(t *testing.T, planId string, provider internal.TrialCloudProvider) internal.ProvisioningOperation {
	return internal.ProvisioningOperation{
		Operation: internal.Operation{
			ID:                     statusOperationID,
			InstanceID:             statusInstanceID,
			ProvisionerOperationID: statusProvisionerOperationID,
			Description:            "",
			UpdatedAt:              time.Now(),
		},
		ProvisioningParameters: fixProvisioningParametersRuntimeStatusWithProvider(t, planId, &provider),
	}
}

func fixProvisioningParametersRuntimeStatus(t *testing.T, planId string) string {
	return fixProvisioningParametersRuntimeStatusWithProvider(t, planId, nil)
}

func fixProvisioningParametersRuntimeStatusWithProvider(t *testing.T, planId string, provider *internal.TrialCloudProvider) string {
	parameters := internal.ProvisioningParameters{
		PlanID:    planId,
		ServiceID: "",
		ErsContext: internal.ERSContext{
			GlobalAccountID: statusGlobalAccountID,
		},
		Parameters: internal.ProvisioningParametersDTO{
			Provider: provider,
		},
	}

	rawParameters, err := json.Marshal(parameters)
	if err != nil {
		t.Errorf("cannot marshal provisioning parameters: %s", err)
	}

	return string(rawParameters)
}

func fixInstanceRuntimeStatus() internal.Instance {
	return internal.Instance{
		InstanceID:      statusInstanceID,
		RuntimeID:       statusRuntimeID,
		DashboardURL:    "",
		GlobalAccountID: statusGlobalAccountID,
		CreatedAt:       time.Time{},
		UpdatedAt:       time.Time{},
		DeletedAt:       time.Time{},
	}
}

func newInMemoryKymaVersionConfigurator(versions map[string]string) *inMemoryKymaVersionConfigurator {
	return &inMemoryKymaVersionConfigurator{
		perGAID: versions,
	}
}

type inMemoryKymaVersionConfigurator struct {
	perGAID map[string]string
}

func (c *inMemoryKymaVersionConfigurator) ForGlobalAccount(string) (string, bool, error) {
	return "", true, nil
}
