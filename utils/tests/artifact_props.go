package tests

import (
	"strings"
	"testing"

	buildinfo "github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ArtifactFullPath builds the Artifactory item path for GetItemProps.
// When OriginalDeploymentRepo is empty (common with Gradle extractor build-info),
// defaultRepo is used as the repository prefix.
func ArtifactFullPath(a buildinfo.Artifact, defaultRepo string) string {
	path := strings.TrimPrefix(a.Path, "/")
	repo := a.OriginalDeploymentRepo
	if repo == "" {
		repo = defaultRepo
	}
	if repo != "" {
		return repo + "/" + path
	}
	return path
}

// ArtifactItemPath returns the Artifactory item path for GetItemProps.
// When Name is set and not already part of Path (e.g. UV stores Path as a directory),
// Name is appended as the filename segment.
func ArtifactItemPath(a buildinfo.Artifact, defaultRepo string) string {
	fullPath := ArtifactFullPath(a, defaultRepo)
	if a.Name == "" {
		return fullPath
	}
	if strings.HasSuffix(fullPath, "/"+a.Name) || strings.HasSuffix(fullPath, a.Name) {
		return fullPath
	}
	return fullPath + "/" + a.Name
}

// ValidateLocalGitVcsPropsOnBuildInfoArtifacts fetches props for each build-info artifact
// and asserts local-git VCS fields. Returns the number of artifacts validated.
func ValidateLocalGitVcsPropsOnBuildInfoArtifacts(
	t *testing.T,
	serviceManager artifactory.ArtifactoryServicesManager,
	publishedBuildInfo *buildinfo.PublishedBuildInfo,
	defaultRepo string,
	expectedURL, expectedRevision, expectedBranch string,
) int {
	t.Helper()
	require.NotNil(t, publishedBuildInfo)

	count := 0
	for _, module := range publishedBuildInfo.BuildInfo.Modules {
		for _, artifact := range module.Artifacts {
			fullPath := ArtifactItemPath(artifact, defaultRepo)
			if fullPath == "" {
				continue
			}

			props, err := serviceManager.GetItemProps(fullPath)
			require.NoError(t, err, "GetItemProps failed for %s", fullPath)
			if props == nil {
				assert.Fail(t, "Properties are nil for artifact: %s", fullPath)
				continue
			}

			assert.Contains(t, props.Properties, "vcs.url", "Missing vcs.url on %s", artifact.Name)
			assert.Contains(t, props.Properties["vcs.url"], expectedURL, "Wrong vcs.url on %s", artifact.Name)

			assert.Contains(t, props.Properties, "vcs.revision", "Missing vcs.revision on %s", artifact.Name)
			assert.Contains(t, props.Properties["vcs.revision"], expectedRevision, "Wrong vcs.revision on %s", artifact.Name)

			if expectedBranch != "" {
				assert.Contains(t, props.Properties, "vcs.branch", "Missing vcs.branch on %s", artifact.Name)
				assert.Contains(t, props.Properties["vcs.branch"], expectedBranch, "Wrong vcs.branch on %s", artifact.Name)
			}

			// Local-git-only: provider/org/repo must NOT appear when CI is cleared
			_, hasProvider := props.Properties["vcs.provider"]
			assert.False(t, hasProvider, "vcs.provider should not be set on %s in local-git-only mode", artifact.Name)

			count++
		}
	}
	return count
}
