package tests

import (
	"testing"

	"github.com/jfrog/build-info-go/utils/cienv"
	"github.com/stretchr/testify/assert"
)

func TestSetupGitHubActionsEnvForLocalGitMerge_ClearsUrlRevisionBranch(t *testing.T) {
	t.Setenv("CI", "true")
	t.Setenv("GITHUB_ACTIONS", "true")
	t.Setenv("GITHUB_WORKFLOW", "wf")
	t.Setenv("GITHUB_RUN_ID", "99")
	t.Setenv("GITHUB_REPOSITORY_OWNER", "jfrog")
	t.Setenv("GITHUB_REPOSITORY", "jfrog/jfrog-cli")
	t.Setenv("GITHUB_SERVER_URL", "https://github.com")
	t.Setenv("GITHUB_SHA", "abc123")
	t.Setenv("GITHUB_REF", "refs/heads/feature")

	cleanup, _, _ := SetupGitHubActionsEnvForLocalGitMerge(t)
	defer cleanup()

	info := cienv.GetCIVcsInfo()
	assert.Equal(t, "github", info.Provider)
	assert.Equal(t, "jfrog", info.Org)
	assert.Equal(t, "jfrog-cli", info.Repo)
	assert.Empty(t, info.Url)
	assert.Empty(t, info.Revision)
	assert.Empty(t, info.Branch)
}
