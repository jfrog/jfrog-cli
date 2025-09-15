package services

import (
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"testing"

	"github.com/jfrog/jfrog-client-go/http/httpclient"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	clientStats "github.com/jfrog/jfrog-client-go/utils/stats"
	"github.com/stretchr/testify/assert"
)

const adminTokenValue = "applied-permissions/admin"

type TokenInfo struct {
	TokenID string `json:"token_id"`
	Subject string `json:"subject"`
	Scope   string `json:"scope"`
}

type statsTestFunc func(client *httpclient.HttpClient, artifactoryUrl string, hd httputils.HttpClientDetails) ([]byte, error)

func GetTokenID(tokenString string) (string, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("failed to read token claims")
	}

	tokenID, ok := claims["jti"].(string)
	if !ok {
		return "", fmt.Errorf("token does not contain a valid 'jti' (Token ID) claim")
	}

	return tokenID, nil
}

func IsAdminToken(client *httpclient.HttpClient, baseUrl string, tokenString string, httpClientDetails httputils.HttpClientDetails) bool {
	tokenID, err := GetTokenID(tokenString)
	if err != nil {
		return false
	}

	body, err := clientStats.GetTokenDetails(client, baseUrl, tokenID, httpClientDetails)
	if err != nil {
		return false
	}
	var tokenInfo TokenInfo
	err = json.Unmarshal(body, &tokenInfo)
	if err != nil {
		log.Error("Error parsing JSON:", err)
		return false
	}
	return tokenInfo.Scope == adminTokenValue
}

func setupTestClient(t *testing.T) (*httpclient.HttpClient, httputils.HttpClientDetails, *config.ServerDetails, bool) {
	serverDetails, err := config.GetDefaultServerConf()
	if err != nil {
		assert.NoError(t, err)
	}

	httpClientDetails := httputils.HttpClientDetails{AccessToken: serverDetails.AccessToken}

	httpClientDetails.SetContentTypeApplicationJson()
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		assert.NoError(t, err)
	}

	serverUrl := serverDetails.GetUrl()

	isAdminToken := IsAdminToken(client, serverUrl, serverDetails.AccessToken, httpClientDetails)

	assert.NoError(t, err)
	return client, httpClientDetails, serverDetails, isAdminToken
}

func TestAllProductAPIs(t *testing.T) {
	productFunctions := map[string]statsTestFunc{
		"Artifactory":    clientStats.GetArtifactoryStats,
		"Repositories":   clientStats.GetRepositoriesStats,
		"Projects":       clientStats.GetProjectsStats,
		"JPDs":           clientStats.GetJPDsStats,
		"ReleaseBundles": clientStats.GetReleaseBundlesStats,
	}
	testCasesMap := map[string]bool{
		"Artifactory":    false,
		"Repositories":   false,
		"Projects":       true,
		"JPDs":           true,
		"ReleaseBundles": false,
	}
	for product, getFunc := range productFunctions {
		t.Run(product, func(t *testing.T) {
			t.Run(product, func(t *testing.T) {
				client, httpClientDetails, server, isAdmin := setupTestClient(t)
				if isAdmin {
					_, err := getFunc(client, server.GetUrl(), httpClientDetails)
					if err != nil {
						assert.NoError(t, err)
					}
				} else {
					_, err := getFunc(client, server.GetUrl(), httpClientDetails)
					if err != nil {
						if !testCasesMap[product] {
							assert.NoError(t, err)
						}
					}
				}
			})
		})
	}
}
