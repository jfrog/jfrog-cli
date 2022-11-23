package pipelines

import (
	utilsconfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	clientConfig "github.com/jfrog/jfrog-client-go/config"
	"github.com/jfrog/jfrog-client-go/pipelines"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
)

/*
getServiceDetails returns server details based on serverID
if serverID is empty returns default config otherwise error
*/
func getServiceDetails(serverID string) (*utilsconfig.ServerDetails, error) {
	if serverID == "" {
		conf, err := utilsconfig.GetDefaultServerConf()
		if err != nil {
			clientlog.Error("unable to find server configuration exiting")
			return nil, err
		}
		serverID = conf.ServerId
	}
	serviceDetails, err := utilsconfig.GetSpecificConfig(serverID, false, false)
	if err != nil {
		clientlog.Error(err)
		return nil, err
	}
	return serviceDetails, err
}

/*
getPipelinesManager creates pipelines manager from jfrog-go-client
*/
func getPipelinesManager(serviceDetails *utilsconfig.ServerDetails) (*pipelines.PipelinesServicesManager, error) {
	pipelinesDetails := *serviceDetails
	pAuth, authErr := pipelinesDetails.CreatePipelinesAuthConfig()
	if authErr != nil {
		return nil, authErr
	}
	serviceConfig, err := clientConfig.NewConfigBuilder().
		SetServiceDetails(pAuth).
		SetDryRun(false).
		Build()
	if err != nil {
		clientlog.Error(err)
		return nil, err
	}
	pipelinesMgr, err := pipelines.New(serviceConfig)
	if err != nil {
		clientlog.Error(err)
		return nil, err
	}
	return pipelinesMgr, nil
}
