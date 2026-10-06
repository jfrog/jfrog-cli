package psresource

import (
	"strings"
	"testing"
)

// TestGetAIDescriptionKnownCmdlets closes a real test-coverage gap: cmdletMetaByName and
// GetAIDescription's map lookup into it had no test at all across any of the four cmdlets, so a
// typo'd key or a renamed SubCommand* constant elsewhere would silently render an empty/broken
// description with nothing to catch it.
func TestGetAIDescriptionKnownCmdlets(t *testing.T) {
	for _, cmdletName := range []string{"Install-PSResource", "Save-PSResource", "Update-PSResource", "Publish-PSResource"} {
		t.Run(cmdletName, func(t *testing.T) {
			desc := GetAIDescription(cmdletName)
			if desc == "" {
				t.Fatalf("GetAIDescription(%q) returned an empty string", cmdletName)
			}
			for _, marker := range []string{cmdletName, "Prerequisites:", "Examples:", "Gotchas:", "$ jf " + cmdletName} {
				if !strings.Contains(desc, marker) {
					t.Errorf("GetAIDescription(%q) is missing expected content %q\nfull output:\n%s", cmdletName, marker, desc)
				}
			}
		})
	}
}

// TestGetAIDescriptionUnknownCmdletDoesNotPanic documents the current fallback behavior for a
// cmdletName with no matching cmdletMetaByName entry (e.g. a future typo or renamed constant):
// it must not panic, even though the rendered text is necessarily incomplete.
func TestGetAIDescriptionUnknownCmdletDoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("GetAIDescription panicked on an unknown cmdlet name: %v", r)
		}
	}()
	desc := GetAIDescription("Not-A-Real-Cmdlet")
	if !strings.Contains(desc, "Prerequisites:") {
		t.Errorf("expected the shared prose to still render even with no metadata match, got: %q", desc)
	}
}

func TestUsageGetDescriptionGetArguments(t *testing.T) {
	for _, cmdletName := range []string{"Install-PSResource", "Save-PSResource", "Update-PSResource", "Publish-PSResource"} {
		if got := Usage(cmdletName); len(got) != 1 || !strings.Contains(got[0], cmdletName) {
			t.Errorf("Usage(%q) = %v, want a single line containing the cmdlet name", cmdletName, got)
		}
		if got := GetDescription(cmdletName); !strings.Contains(got, cmdletName) {
			t.Errorf("GetDescription(%q) = %q, want it to contain the cmdlet name", cmdletName, got)
		}
		if got := GetArguments(cmdletName); !strings.Contains(got, cmdletName) {
			t.Errorf("GetArguments(%q) = %q, want it to contain the cmdlet name", cmdletName, got)
		}
	}
}
