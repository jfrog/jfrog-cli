package pip

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"

	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

func GetPipConfiguration() (*utils.RepositoryConfig, error) {
	// Get configuration file path
	confFilePath, exists, err := utils.GetProjectConfFilePath(utils.Pip)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errorutils.CheckError(fmt.Errorf("Pip Project configuration does not exist."))
	}
	// Read config file
	log.Debug("Preparing to read the config file", confFilePath)
	vConfig, err := utils.ReadConfigFile(confFilePath, utils.YAML)
	if err != nil {
		return nil, err
	}
	return utils.GetRepoConfigByPrefix(confFilePath, utils.ProjectConfigResolverPrefix, vConfig)
}
