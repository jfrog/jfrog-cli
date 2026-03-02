package packagealias

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsJFrogCLINameReturnsTrueForCLIBinaries(t *testing.T) {
	require.True(t, isJFrogCLIName("jf"))
	require.True(t, isJFrogCLIName("jfrog"))
}

func TestIsJFrogCLINameReturnsFalseForSupportedTools(t *testing.T) {
	for _, tool := range SupportedTools {
		require.False(t, isJFrogCLIName(tool), "SupportedTools entry %q must not be a JFrog CLI name", tool)
	}
}

func TestIsJFrogCLINameReturnsFalseForArbitraryNames(t *testing.T) {
	require.False(t, isJFrogCLIName(""))
	require.False(t, isJFrogCLIName("JF"))
	require.False(t, isJFrogCLIName("Jfrog"))
	require.False(t, isJFrogCLIName("random-binary"))
}

func TestSupportedToolsNeverContainsCLINames(t *testing.T) {
	for _, tool := range SupportedTools {
		require.NotEqual(t, "jf", tool, "SupportedTools must never contain 'jf'")
		require.NotEqual(t, "jfrog", tool, "SupportedTools must never contain 'jfrog'")
	}
}
