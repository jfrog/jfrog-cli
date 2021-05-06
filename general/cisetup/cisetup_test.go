package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractRepositoryName(t *testing.T) {
	tests := []struct {
		name                  string
		repoUrl               string
		expectedProjectName   string
		expectedProjectDomain string
		expectedBaseUrl       string
		errorExpected         bool
	}{
		{"simpleURL", "https://github.com/jfrog/jfrog-cli", "jfrog-cli", "jfrog", "https://github.com", false},
		{"URLWithTrailingDash", "https://github.com/jfrog/jfrog-cli/", "jfrog-cli", "jfrog", "https://github.com", false},
		{"URLWithGitExtension", "https://github.com/jfrog/jfrog-cli.git", "jfrog-cli", "jfrog", "https://github.com", false},
		{"noProtocol", "localhost:1234/jfrog/jfrog-cli.git", "jfrog-cli", "jfrog", "localhost:1234", false},
		{"emptyURL", "", "", "", "", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cc := &CiSetupCommand{}
			err := cc.prepareConfigurationData()
			if err != nil {
				assert.NoError(t, err)
				return
			}
			cc.data.VcsCredentials.Url = test.repoUrl

			err = cc.extractRepositoryName()
			if test.errorExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedProjectName, cc.data.RepositoryName)
				assert.Equal(t, test.expectedProjectDomain, cc.data.ProjectDomain)
				assert.Equal(t, test.expectedBaseUrl, cc.data.VcsBaseUrl)
			}
		})
	}
}

func TestGetExplicitTechsListByNumber(t *testing.T) {
	tests := []struct {
		name                  string
		techs               []string
		expected				string
	}{
		{"one tech", []string{"maven"}, "maven"},
		{"two techs", []string{"maven", "gradle"}, "maven and gradle"},
		{"three techs", []string{"maven", "gradle", "npm"}, "maven, gradle and npm"},
		{"four techs", []string{"maven", "gradle", "npm", "something"}, "maven, gradle, npm and something"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := getExplicitTechsListByNumber(test.techs)
			assert.Equal(t, test.expected, output)
		})
	}
}
