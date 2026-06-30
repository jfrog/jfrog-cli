package cliutils

import "testing"

func TestCargoFlagGroupExists(t *testing.T) {
	flags := GetCommandFlags(Cargo)
	if len(flags) == 0 {
		t.Fatal("expected cargo flag group to be registered")
	}
}
