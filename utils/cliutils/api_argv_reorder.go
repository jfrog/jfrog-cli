package cliutils

import (
	"strings"

	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/urfave/cli"
)

// NormalizeApiTrailingFlags works around a quirk of urfave/cli v1: a command with
// Subcommands (like "api", which has "docs") skips the library's own flag
// reordering and hands args straight to the stdlib flag package, which stops
// parsing at the first non-flag token. That makes "jf api <path> --flag" fail
// with a bogus argument-count error, even though it's the documented usage.
// "jf api docs search/describe" are unaffected (they're leaf commands, so
// urfave/cli reorders their flags on its own) and must be left untouched here.
func NormalizeApiTrailingFlags(argv []string) []string {
	if len(argv) < 3 || argv[1] != "api" {
		return argv
	}
	if len(argv) >= 3 && argv[2] == "docs" {
		return argv
	}

	prefix := argv[:2]
	tail := append([]string(nil), argv[2:]...)
	flagTokens, positionals := extractKnownFlags(GetCommandFlags(Api), tail)

	reordered := make([]string, 0, len(argv)-2)
	reordered = append(reordered, flagTokens...)
	reordered = append(reordered, positionals...)
	return append(prefix, reordered...)
}

// extractKnownFlags pulls every occurrence of the given flags out of args, wherever
// they appear, and returns them separately from the remaining positional arguments.
// Relative order is preserved within each group.
func extractKnownFlags(flags []cli.Flag, args []string) (flagTokens, positionals []string) {
	remaining := append([]string(nil), args...)
	for _, f := range flags {
		aliases := flagAliases(f)
		if _, isBool := f.(cli.BoolFlag); isBool {
			for {
				flagIndex, _, err := findFirstBooleanFlag(aliases, remaining)
				if err != nil || flagIndex == -1 {
					break
				}
				flagTokens = append(flagTokens, remaining[flagIndex])
				coreutils.RemoveFlagFromCommand(&remaining, flagIndex, flagIndex)
			}
			continue
		}
		for {
			flagIndex, flagValueIndex, _, err := coreutils.FindFlagFirstMatch(aliases, remaining)
			if err != nil || flagIndex == -1 {
				break
			}
			flagTokens = append(flagTokens, remaining[flagIndex:flagValueIndex+1]...)
			coreutils.RemoveFlagFromCommand(&remaining, flagIndex, flagValueIndex)
		}
	}
	return flagTokens, remaining
}

func findFirstBooleanFlag(aliases, args []string) (flagIndex int, flagValue bool, err error) {
	for _, alias := range aliases {
		flagIndex, flagValue, err = coreutils.FindBooleanFlag(alias, args)
		if err != nil || flagIndex != -1 {
			return
		}
	}
	return -1, false, nil
}

// flagAliases returns every spelling of a flag, dash-prefixed as urfave/cli expects
// it on the command line, e.g. "header, H" -> ["--header", "-H"].
func flagAliases(f cli.Flag) []string {
	names := strings.Split(f.GetName(), ",")
	aliases := make([]string, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if len(name) == 1 {
			aliases = append(aliases, "-"+name)
		} else {
			aliases = append(aliases, "--"+name)
		}
	}
	return aliases
}
