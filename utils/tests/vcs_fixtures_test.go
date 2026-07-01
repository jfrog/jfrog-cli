package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCopyGitFixtureIntoProject_WorksAfterChdir(t *testing.T) {
	repoRoot, err := os.Getwd()
	require.NoError(t, err)

	projectDir := t.TempDir()
	subDir := filepath.Join(projectDir, "nested")
	require.NoError(t, os.MkdirAll(subDir, 0o755))

	// Simulate prepareGoProject
	// Leaving cwd inside the project tree.
	require.NoError(t, os.Chdir(subDir))
	t.Cleanup(func() { _ = os.Chdir(repoRoot) })

	CopyGitFixtureIntoProject(t, projectDir)

	require.FileExists(t, filepath.Join(projectDir, ".git", "HEAD"))
	require.FileExists(t, filepath.Join(projectDir, ".git", "config"))
}
