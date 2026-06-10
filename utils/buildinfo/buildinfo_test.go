package buildinfo

import (
	"reflect"
	"testing"
)

// TestPoetryShowGroupArgs verifies that only the dependency-group filter flags
// (--only/--with/--without) are forwarded to `poetry show`, that both the
// "--flag value" and "--flag=value" forms are handled, and that --all-groups
// short-circuits to include-all (legacy behaviour).
func TestPoetryShowGroupArgs(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		wantGroupArgs  []string
		wantIncludeAll bool
	}{
		{
			name:          "only main, space form",
			args:          []string{"--only", "main", "--no-root"},
			wantGroupArgs: []string{"--only", "main"},
		},
		{
			name:          "only main, equals form",
			args:          []string{"--only=main", "--no-root"},
			wantGroupArgs: []string{"--only=main"},
		},
		{
			name:          "without dev",
			args:          []string{"--without", "dev"},
			wantGroupArgs: []string{"--without", "dev"},
		},
		{
			name:          "with optional, multiple group flags",
			args:          []string{"--with", "docs", "--without", "dev"},
			wantGroupArgs: []string{"--with", "docs", "--without", "dev"},
		},
		{
			name:           "all-groups short-circuits to include-all",
			args:           []string{"--all-groups"},
			wantGroupArgs:  nil,
			wantIncludeAll: true,
		},
		{
			name:          "no group flags drops everything else",
			args:          []string{"--no-root", "--compile", "-E", "extra"},
			wantGroupArgs: nil,
		},
		{
			name:          "trailing group flag without value is forwarded as-is",
			args:          []string{"--only"},
			wantGroupArgs: []string{"--only"},
		},
		{
			name:          "empty args",
			args:          nil,
			wantGroupArgs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotGroupArgs, gotIncludeAll := poetryShowGroupArgs(tt.args)
			if gotIncludeAll != tt.wantIncludeAll {
				t.Errorf("includeAll = %v, want %v", gotIncludeAll, tt.wantIncludeAll)
			}
			if !reflect.DeepEqual(gotGroupArgs, tt.wantGroupArgs) {
				t.Errorf("groupArgs = %#v, want %#v", gotGroupArgs, tt.wantGroupArgs)
			}
		})
	}
}

// TestNormalizePoetryPipName verifies PEP 503 normalisation: lowercase with runs
// of [-_.] collapsed to a single "-" and no leading/trailing separator.
func TestNormalizePoetryPipName(t *testing.T) {
	tests := map[string]string{
		"Requests":           "requests",
		"ruamel.yaml":        "ruamel-yaml",
		"my__weird..name":    "my-weird-name",
		"Already-Normalised": "already-normalised",
		"typing_extensions":  "typing-extensions",
	}
	for in, want := range tests {
		if got := normalizePoetryPipName(in); got != want {
			t.Errorf("normalizePoetryPipName(%q) = %q, want %q", in, got, want)
		}
	}
}
