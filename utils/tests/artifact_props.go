package tests

import (
	"strings"

	buildinfo "github.com/jfrog/build-info-go/entities"
)

func ArtifactFullPath(a buildinfo.Artifact) string {
	path := strings.TrimPrefix(a.Path, "/")
	if a.OriginalDeploymentRepo != "" {
		return a.OriginalDeploymentRepo + "/" + path
	}
	return path
}
