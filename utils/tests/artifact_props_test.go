package tests

import (
	"testing"

	buildinfo "github.com/jfrog/build-info-go/entities"
	"github.com/stretchr/testify/assert"
)

func TestArtifactFullPath(t *testing.T) {
	t.Run("uses OriginalDeploymentRepo when set", func(t *testing.T) {
		a := buildinfo.Artifact{OriginalDeploymentRepo: "cli-gradle-123", Path: "com/foo/1.0/foo.jar"}
		assert.Equal(t, "cli-gradle-123/com/foo/1.0/foo.jar", ArtifactFullPath(a, "fallback-repo"))
	})

	t.Run("falls back to defaultRepo when OriginalDeploymentRepo empty", func(t *testing.T) {
		a := buildinfo.Artifact{Path: "com/foo/1.0/foo.jar"}
		assert.Equal(t, "cli-gradle-123/com/foo/1.0/foo.jar", ArtifactFullPath(a, "cli-gradle-123"))
	})

	t.Run("falls back to Path when repo empty and no default", func(t *testing.T) {
		a := buildinfo.Artifact{Path: "com/foo/1.0/foo.jar"}
		assert.Equal(t, "com/foo/1.0/foo.jar", ArtifactFullPath(a, ""))
	})

	t.Run("strips leading slash from Path", func(t *testing.T) {
		a := buildinfo.Artifact{Path: "/minimal-example/1.0/minimal-example-1.0.jar"}
		assert.Equal(t, "cli-gradle-123/minimal-example/1.0/minimal-example-1.0.jar", ArtifactFullPath(a, "cli-gradle-123"))
	})
}

func TestValidateLocalGitVcsPropsOnBuildInfoArtifacts_UsesArtifactFullPath(t *testing.T) {
	// Smoke-test ArtifactFullPath integration used by the helper (no Artifactory call).
	a := buildinfo.Artifact{
		OriginalDeploymentRepo: "",
		Path:                   "/com/foo/1.0/foo.jar",
	}
	assert.Equal(t, "my-repo/com/foo/1.0/foo.jar", ArtifactFullPath(a, "my-repo"))
}

func TestArtifactItemPath_AppendsNameForDirectoryPath(t *testing.T) {
	a := buildinfo.Artifact{
		OriginalDeploymentRepo: "uv-local",
		Path:                   "my-pkg/0.1.0",
		Name:                   "my_pkg-0.1.0-py3-none-any.whl",
	}
	assert.Equal(t, "uv-local/my-pkg/0.1.0/my_pkg-0.1.0-py3-none-any.whl", ArtifactItemPath(a, ""))
}

func TestArtifactItemPath_DoesNotDoubleAppendName(t *testing.T) {
	a := buildinfo.Artifact{
		OriginalDeploymentRepo: "mvn-local",
		Path:                   "com/foo/1.0/foo.jar",
		Name:                   "foo.jar",
	}
	assert.Equal(t, "mvn-local/com/foo/1.0/foo.jar", ArtifactItemPath(a, ""))
}

func TestArtifactItemPath_NpmTarballPathDoesNotAppendName(t *testing.T) {
	a := buildinfo.Artifact{
		OriginalDeploymentRepo: "cli-npm-123",
		Path:                   "jfrog-cli-tests/-/jfrog-cli-tests-1.0.0.tgz",
		Name:                   "jfrog-cli-tests-v1.0.0.tgz",
	}
	assert.Equal(t, "cli-npm-123/jfrog-cli-tests/-/jfrog-cli-tests-1.0.0.tgz", ArtifactItemPath(a, ""))
}
