package tests

import (
	"strings"

	buildinfo "github.com/jfrog/build-info-go/entities"
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
