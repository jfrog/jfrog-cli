package services

//
//import (
//	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
//	"github.com/jfrog/jfrog-cli-core/v2/general"
//	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
//	"github.com/stretchr/testify/assert"
//	"testing"
//)
//
//type statsTestFunc func(string) ([]byte, error)
//
//func TestAllProductAPIs(t *testing.T) {
//	ss := general.NewStatsCommand()
//	serverDetails, err := config.GetDefaultServerConf()
//	if err != nil {
//		assert.NoError(t, err)
//	}
//	ss.AccessToken = serverDetails.AccessToken
//
//	productFunctions := map[string]statsTestFunc{
//		"Repositories":   ss.ServicesManager.GetRepositoriesStats,
//		"ReleaseBundles": ss.ServicesManager.GetReleaseBundlesStats,
//	}
//	for product, getFunc := range productFunctions {
//		t.Run(product, func(t *testing.T) {
//			t.Run(product, func(t *testing.T) {
//				_, err := getFunc(serverDetails.Url)
//				if err != nil {
//					assert.NoError(t, err)
//				}
//			})
//		})
//	}
//}
