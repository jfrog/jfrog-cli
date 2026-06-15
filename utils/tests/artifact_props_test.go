package tests

import (
	"testing"

	buildinfo "github.com/jfrog/build-info-go/entities"
	"github.com/stretchr/testify/assert"
)

func TestArtifactFullPath(t *testing.T) {
	t.Run("uses OriginalDeploymentRepo when set", func(t *testing.T) {
		a := buildinfo.Artifact{OriginalDeploymentRepo: "cli-gradle-123", Path: "com/foo/1.0/foo.jar"}
		assert.Equal(t, "cli-gradle-123/com/foo/1.0/foo.jar", ArtifactFullPath(a))
	})

	t.Run("falls back to Path when repo empty", func(t *testing.T) {
		a := buildinfo.Artifact{Path: "com/foo/1.0/foo.jar"}
		assert.Equal(t, "com/foo/1.0/foo.jar", ArtifactFullPath(a))
	})

	t.Run("strips leading slash from Path", func(t *testing.T) {
		a := buildinfo.Artifact{Path: "/minimal-example/1.0/minimal-example-1.0.jar"}
		assert.Equal(t, "minimal-example/1.0/minimal-example-1.0.jar", ArtifactFullPath(a))
	})
}
