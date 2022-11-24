package pipelines

import (
	utilsconfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
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
