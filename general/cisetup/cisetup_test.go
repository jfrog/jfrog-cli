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
	}{
		{"simpleURL", "https://github.com/jfrog/jfrog-cli", "jfrog-cli", "jfrog"},
		{"URLWithTrailingDash", "https://github.com/jfrog/jfrog-cli/", "jfrog-cli", "jfrog"},
		{"URLWithGitExtension", "https://github.com/jfrog/jfrog-cli.git", "jfrog-cli", "jfrog"},
		{"emptyURL", "", "", ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cc := &CiSetupCommand{}
			err := cc.prepareConfigurationData()
			cc.data.VcsCredentials.Url = test.repoUrl
			assert.NoError(t, err)
			cc.extractRepositoryName()
			assert.Equal(t, test.expectedProjectName, cc.data.RepositoryName)
			assert.Equal(t, test.expectedProjectDomain, cc.data.ProjectDomain)

		})
	}
}
