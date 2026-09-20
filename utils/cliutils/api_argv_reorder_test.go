package cliutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

func TestNormalizeApiTrailingFlags(t *testing.T) {
	tests := []struct {
		name string
		argv []string
		want []string
	}{
		{
			name: "JGC-539 repro: single flag after path",
			argv: []string{"jf", "api", "/xray/api/v1/system/version", "--server-id=repo21"},
			want: []string{"jf", "api", "--server-id=repo21", "/xray/api/v1/system/version"},
		},
		{
			// Flags are extracted in the command's own flag-definition order (alphabetical
			// by flag name, see buildAndSortFlags), not their original relative order in
			// argv -- that's fine, since flag.Parse doesn't care about order between
			// distinct flags, only that each flag stays adjacent to its own value.
			name: "JGC-539 repro: multiple flags after path",
			argv: []string{"jf", "api", "/xray/api/v1/ignore_rules", "-X", "POST", "-H", "Content-Type: application/json", "--input", "./body.json"},
			want: []string{"jf", "api", "-H", "Content-Type: application/json", "--input", "./body.json", "-X", "POST", "/xray/api/v1/ignore_rules"},
		},
		{
			name: "flags already before path: no-op",
			argv: []string{"jf", "api", "--server-id=repo21", "/xray/api/v1/system/version"},
			want: []string{"jf", "api", "--server-id=repo21", "/xray/api/v1/system/version"},
		},
		{
			name: "boolean flag after path",
			argv: []string{"jf", "api", "/artifactory/api/repositories", "--verbose"},
			want: []string{"jf", "api", "--verbose", "/artifactory/api/repositories"},
		},
		{
			name: "repeated header flag after path",
			argv: []string{"jf", "api", "/artifactory/api/repositories", "-H", "A: 1", "-H", "B: 2"},
			want: []string{"jf", "api", "-H", "A: 1", "-H", "B: 2", "/artifactory/api/repositories"},
		},
		{
			name: "too many positionals: order preserved, arity error left for the real parser",
			argv: []string{"jf", "api", "/a", "/b", "-X", "GET"},
			want: []string{"jf", "api", "-X", "GET", "/a", "/b"},
		},
		{
			name: "unrecognized flag after path is left untouched",
			argv: []string{"jf", "api", "/artifactory/api/repositories", "--bogus", "x"},
			want: []string{"jf", "api", "/artifactory/api/repositories", "--bogus", "x"},
		},
		{
			name: "docs search: untouched",
			argv: []string{"jf", "api", "docs", "search", "user", "--tag=access"},
			want: []string{"jf", "api", "docs", "search", "user", "--tag=access"},
		},
		{
			name: "docs describe: untouched",
			argv: []string{"jf", "api", "docs", "describe", "GET", "/access/api/v2/users", "--format=json"},
			want: []string{"jf", "api", "docs", "describe", "GET", "/access/api/v2/users", "--format=json"},
		},
		{
			name: "bare docs: untouched",
			argv: []string{"jf", "api", "docs"},
			want: []string{"jf", "api", "docs"},
		},
		{
			name: "unrelated command: untouched",
			argv: []string{"jf", "rt", "upload", "a", "b", "--flat=true"},
			want: []string{"jf", "rt", "upload", "a", "b", "--flat=true"},
		},
		{
			name: "argv too short: untouched",
			argv: []string{"jf", "api"},
			want: []string{"jf", "api"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NormalizeApiTrailingFlags(tt.argv))
		})
	}
}

// TestNormalizeApiTrailingFlags_FixesRealParsing proves the reordering actually
// resolves urfave/cli's arity/flag-parsing failure for "api", using the same
// Flags + Subcommands shape production registers in main.go.
func TestNormalizeApiTrailingFlags_FixesRealParsing(t *testing.T) {
	var gotArgs cli.Args
	var gotMethod string
	var gotHeaders []string
	var gotInput string

	app := cli.NewApp()
	app.Commands = []cli.Command{
		{
			Name:  "api",
			Flags: GetCommandFlags(Api),
			Action: func(c *cli.Context) error {
				gotArgs = c.Args()
				gotMethod = c.String("method")
				gotHeaders = c.StringSlice("header")
				gotInput = c.String("input")
				return nil
			},
			Subcommands: []cli.Command{{Name: "docs"}},
		},
	}

	rawArgs := []string{"jf", "api", "/xray/api/v1/ignore_rules", "-X", "POST", "-H", "Content-Type: application/json", "--input", "./body.json"}

	// Sanity check: without the fix, urfave/cli's flag parsing breaks on this
	// exact input (path before flags), matching the ticket's reported bug.
	require.NoError(t, app.Run(rawArgs))
	assert.NotEqual(t, 1, len(gotArgs), "expected the unfixed argv to already be broken (more than 1 positional arg)")

	gotArgs, gotMethod, gotHeaders, gotInput = nil, "", nil, ""
	require.NoError(t, app.Run(NormalizeApiTrailingFlags(rawArgs)))
	require.Equal(t, 1, len(gotArgs))
	assert.Equal(t, "/xray/api/v1/ignore_rules", gotArgs.First())
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, []string{"Content-Type: application/json"}, gotHeaders)
	assert.Equal(t, "./body.json", gotInput)
}
