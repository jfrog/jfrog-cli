package services

import (
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

type statsTestFunc func(string) ([]byte, error)

func TestAllProductAPIs(t *testing.T) {
	ss := generic.NewStatsCommand()
	serverDetails, err := config.GetDefaultServerConf()
	if err != nil {
		assert.NoError(t, err)
	}
	servicesManager, err := utils.CreateServiceManager(serverDetails, -1, 0, false)
	if err != nil {
		assert.NoError(t, err)
	}
	ss.ServicesManager = servicesManager
	ss.AccessToken = serverDetails.AccessToken

	productFunctions := map[string]statsTestFunc{
		"Repositories":   ss.ServicesManager.GetRepositoriesStats,
		"ReleaseBundles": ss.ServicesManager.GetReleaseBundlesStats,
	}
	for product, getFunc := range productFunctions {
		t.Run(product, func(t *testing.T) {
			t.Run(product, func(t *testing.T) {
				_, err := getFunc(serverDetails.Url)
				if err != nil {
					assert.NoError(t, err)
				}
			})
		})
	}
}
