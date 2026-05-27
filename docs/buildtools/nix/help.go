package nix

var Usage = []string{"nix <native-command> <args> [command options]"}

func GetDescription() string {
	return "Run native nix commands with build-info support."
}

func GetArguments() string {
	return `	nix native-command
		Wraps native Nix commands (nix-channel, nix-env, nix-build, nix copy)
		with build-info collection. All args are passed through to the native tool.

		Examples:
		- jf nix nix-channel --add https://server/artifactory/api/nix/repo/channels/nixos-25.11 nixpkgs
		- jf nix nix-channel --update
		- jf nix nix-env -iA nixpkgs.hello --build-name=my-build --build-number=1
		- jf nix nix-build '<nixpkgs>' -A hello --build-name=my-build --build-number=1
		- jf nix copy --to "http://user:pass@server/artifactory/api/nix/repo/" ./result --build-name=my-build --build-number=1`
}
