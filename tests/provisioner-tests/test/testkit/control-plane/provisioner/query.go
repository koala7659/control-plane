package provisioner

import "fmt"

type queryProvider struct{}

func (qp queryProvider) provisionRuntime(config string) string {
	return fmt.Sprintf(`mutation {
	result: provisionRuntime(config: %s) { 
		%s 
	}
}`, config, operationStatusData())
}

func (qp queryProvider) upgradeKymaRuntime(runtimeID string, config string) string {
	return fmt.Sprintf(`mutation {
	result: upgradeRuntime(id: "%s", config: %s) {
    	%s
  	}
}`, runtimeID, config, operationStatusData())
}

func (qp queryProvider) upgradeGardenerCluster(runtimeID string, config string) string {
	return fmt.Sprintf(`mutation {
	result: upgradeShoot(id: "%s", config: %s) {
    	%s
  	}
}`, runtimeID, config, operationStatusData())
}

func (qp queryProvider) deprovisionRuntime(runtimeID string) string {
	return fmt.Sprintf(`mutation {
	result: deprovisionRuntime(id: "%s")
}`, runtimeID)
}

func (qp queryProvider) reconnectRuntimeAgent(runtimeID string) string {
	return fmt.Sprintf(`mutation {
	result: reconnectRuntimeAgent(id: "%s")
}`, runtimeID)
}

func (qp queryProvider) runtimeStatus(operationID string) string {
	return fmt.Sprintf(`query {
	result: runtimeStatus(id: "%s") {
		%s
	}
}`, operationID, runtimeStatusData())
}

func (qp queryProvider) runtimeOperationStatus(operationID string) string {
	return fmt.Sprintf(`query {
	result: runtimeOperationStatus(id: "%s") {
		%s
	}
}`, operationID, operationStatusData())
}

func runtimeStatusData() string {
	return fmt.Sprintf(`lastOperationStatus { operation state message }
			runtimeConnectionStatus { status }
			runtimeConfiguration { 
				kubeconfig
				clusterConfig { 
					%s
				} 
				kymaConfig { 
					version
				} 
			}`, clusterConfig())
}

func clusterConfig() string {
	return fmt.Sprintf(`
		name
		kubernetesVersion
		volumeSizeGB
		diskType
		machineType
		region
		purpose
		provider
		seed
		targetSecret
		diskType
		workerCidr
		autoScalerMin
		autoScalerMax
		maxSurge
		maxUnavailable
		enableKubernetesVersionAutoUpdate
		enableMachineImageVersionAutoUpdate
		allowPrivilegedContainers
		providerSpecificConfig {
			%s
		}
`, providerSpecificConfig())
}

func providerSpecificConfig() string {
	return fmt.Sprint(`
		... on GCPProviderConfig { 
			zones 
		} 
		... on AzureProviderConfig {
			vnetCidr
			zones
		}
		... on AWSProviderConfig {
			zone 
			internalCidr 
			vpcCidr 
			publicCidr
		}  
	`)
}

func operationStatusData() string {
	return `id
			operation 
			state
			message
			runtimeID`
}
