package cliutils

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestGetStringsArrFlagValue(t *testing.T) {
	tests := []struct {
		name     string
		flagValue string
		expected []string
	}{
		{
			name:     "Single value",
			flagValue: "repo1",
			expected: []string{"repo1"},
		},
		{
			name:     "Multiple values",
			flagValue: "repo1;repo2;repo3",
			expected: []string{"repo1", "repo2", "repo3"},
		},
		{
			name:     "Values with spaces",
			flagValue: "repo1; repo2 ; repo3",
			expected: []string{"repo1", "repo2", "repo3"},
		},
		{
			name:     "Empty value",
			flagValue: "",
			expected: []string{},
		},
		{
			name:     "Values with empty parts",
			flagValue: "repo1;;repo2;",
			expected: []string{"repo1", "repo2"},
		},
		{
			name:     "Complex repository names",
			flagValue: "my-org/repo1;another-org/repo-with-dashes;third_org/repo_with_underscores",
			expected: []string{"my-org/repo1", "another-org/repo-with-dashes", "third_org/repo_with_underscores"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
			flagSet.String("test-flag", tt.flagValue, "")
			err := flagSet.Parse([]string{"--test-flag", tt.flagValue})
			assert.NoError(t, err)

			c := cli.NewContext(nil, flagSet, nil)

			result := GetStringsArrFlagValue(c, "test-flag")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOverrideArrayIfSet(t *testing.T) {
	tests := []struct {
		name     string
		flagValue string
		expected []string
	}{
		{
			name:     "Single repository",
			flagValue: "my-repo",
			expected: []string{"my-repo"},
		},
		{
			name:     "Multiple repositories",
			flagValue: "repo1;repo2;repo3",
			expected: []string{"repo1", "repo2", "repo3"},
		},
		{
			name:     "Repositories with spaces",
			flagValue: " repo1 ; repo2 ; repo3 ",
			expected: []string{"repo1", "repo2", "repo3"},
		},
		{
			name:     "Empty flag value",
			flagValue: "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
			flagSet.String("test-flag", tt.flagValue, "")
			err := flagSet.Parse([]string{"--test-flag", tt.flagValue})
			assert.NoError(t, err)

			c := cli.NewContext(nil, flagSet, nil)

			var result []string
			overrideArrayIfSet(&result, c, "test-flag")
			assert.Equal(t, tt.expected, result)
		})
	}
}