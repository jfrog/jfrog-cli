package apkcommand

var Usage = []string{
	"apk add <packages...> [command options]",
	"apk upgrade [packages...] [command options]",
	"apk upload <file.apk> --repo <repo-key> --alpine-version <vX.Y> [command options]",
	"apk <native-subcommand> [args...] [command options]",
}

// GetDescription returns the short command description shown in jf --help.
func GetDescription() string {
	return "Run Alpine apk commands through Artifactory, with Build Info collection for add, upgrade, and upload."
}

// GetAIDescription returns the extended description used by AI-assisted help.
func GetAIDescription() string {
	return `jf apk wraps the native apk binary with Artifactory credential injection.
Build Info is collected only for 'add', 'upgrade', and 'upload'. Every other
apk subcommand is a passthrough: flags are accepted (including Build Info
flags), credentials are injected when configured, and the native command runs
as-is. If Build Info flags are passed on a passthrough command, a warning is
printed after execution.

To point apk at Artifactory, run: jf setup apk`
}

// GetArguments returns the argument reference shown in jf apk --help.
func GetArguments() string {
	return `	apk subcommand

		Build Info is collected only for:
		  add      — install packages
		  upgrade  — upgrade packages
		  upload   — publish a local .apk file to Artifactory

		All other apk subcommands are passthroughs. They accept the same
		JFrog flags (including Build Info flags), inject Artifactory
		credentials when configured, and forward the rest to the native
		apk binary. Build Info flags on a passthrough command produce a
		warning after the command finishes; no Build Info is recorded.

		Passthrough examples: update, fetch, search, del, info, fix,
		audit, version, stats, and any other native apk subcommand.

		Flags for add / upgrade:
		  --server-id        JFrog server config ID. Default: active server.
		  --repo             Artifactory Alpine repository key.
		  --alpine-version   Alpine release, e.g. v3.20.
		  --user             Override Artifactory username.
		  --password         Override Artifactory password or token.
		  --build-name       Build Info name. Env: JFROG_CLI_BUILD_NAME.
		  --build-number     Build Info number. Env: JFROG_CLI_BUILD_NUMBER.
		  --module           Build Info module name. Default: <repo>:<arch>:<alpine-version>.
		  --project          JFrog Project key. Env: JFROG_CLI_BUILD_PROJECT.

		Flags for upload:
		  --repo             Artifactory Alpine repository key (required).
		  --alpine-version   Alpine release, e.g. v3.20 (required).
		  --server-id        JFrog server config ID. Default: active server.
		  --branch           Alpine repo branch. Default: main.
		  --arch             CPU architecture. Default: inferred from filename.
		  --user             Override Artifactory username.
		  --password         Override Artifactory password or token.
		  --build-name       Build Info name. Env: JFROG_CLI_BUILD_NAME.
		  --build-number     Build Info number. Env: JFROG_CLI_BUILD_NUMBER.
		  --module           Build Info module name. Default: <repo>:<arch>:<alpine-version>.
		  --project          JFrog Project key. Env: JFROG_CLI_BUILD_PROJECT.

		Flags for passthrough commands:
		  --server-id        JFrog server config ID. Default: active server.
		  --repo             Artifactory Alpine repository key.
		  --alpine-version   Alpine release, e.g. v3.20.
		  --user             Override Artifactory username.
		  --password         Override Artifactory password or token.
		  --build-name       Accepted, but Build Info is not collected.
		  --build-number     Accepted, but Build Info is not collected.
		  --module           Accepted, but Build Info is not collected.
		  --project          Accepted, but Build Info is not collected.

		Examples:
		  jf setup apk --server-id my-server --repo my-alpine-repo
		  jf apk add curl bash --build-name ci-image --build-number 42
		  jf apk upgrade --server-id my-server --repo my-alpine-repo
		  jf apk upload ./myapp-1.0.0-r0.x86_64.apk --repo my-alpine-repo \
		             --alpine-version v3.20 --build-name ci-image --build-number 42
		  jf apk update
		  jf apk search curl
		  jf apk info musl`
}
