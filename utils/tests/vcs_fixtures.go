package tests

import (
	"path/filepath"
	"testing"

	biutils "github.com/jfrog/build-info-go/utils"
	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/stretchr/testify/assert"
)

const (
	VcsFixtureMainURL      = "https://github.com/jfrog/jfrog-cli.git"
	VcsFixtureMainRevision = "d63c5957ad6819f4c02a817abe757f210d35ff92"
	VcsFixtureMainBranch   = "master"

	VcsFixtureOtherURL      = "https://github.com/jfrog/jfrog-client-go.git"
	VcsFixtureOtherRevision = "ad99b6c068283878fde4d49423728f0bdc00544a"
	VcsFixtureOtherBranch   = "InnerGit"
)

// CopyVcsGitFixture copies testdata/vcs into destDir and renames gitdata -> .git.
// Returns the absolute path to destDir.
func CopyVcsGitFixture(t *testing.T, destDir string) string {
	t.Helper()
	src := filepath.Join(filepath.FromSlash(GetTestResourcesPath()), "vcs")
	assert.NoError(t, biutils.CopyDir(src, destDir, true, nil))
	if found, err := fileutils.IsDirExists(filepath.Join(destDir, "gitdata"), false); found {
		assert.NoError(t, err)
		coretests.RenamePath(filepath.Join(destDir, "gitdata"), filepath.Join(destDir, ".git"), t)
	}
	if found, err := fileutils.IsDirExists(filepath.Join(destDir, "OtherGit", "gitdata"), false); found {
		assert.NoError(t, err)
		coretests.RenamePath(
			filepath.Join(destDir, "OtherGit", "gitdata"),
			filepath.Join(destDir, "OtherGit", ".git"),
			t,
		)
	}
	abs, err := filepath.Abs(destDir)
	assert.NoError(t, err)
	return abs
}

// CopyGitFixtureIntoProject copies only gitdata from testdata/vcs/gitdata into projectDir/.git.
func CopyGitFixtureIntoProject(t *testing.T, projectDir string) {
	t.Helper()
	src := filepath.Join(filepath.FromSlash(GetTestResourcesPath()), "vcs", "gitdata")
	dest := filepath.Join(projectDir, ".git")
	assert.NoError(t, biutils.CopyDir(src, dest, true, nil))
}
