package apkcommand

var Usage = []string{
	"apk upload <file.apk> --repo <repo-key> --alpine-version <vX.Y> [--branch <name>] [--arch <arch>] [jf-flags]",
	"apk add <packages...> [jf-flags]",
	"apk upgrade [packages...] [jf-flags]",
	"apk update [jf-flags]",
	"apk fetch <packages...> [jf-flags]",
	"apk search <pattern> [jf-flags]",
	"apk del <packages...>",
	"apk info [packages...]",
}

// GetDescription returns the short command description shown in jf --help.
func GetDescription() string {
	return "Manage Alpine packages via Artifactory: upload packages and run native apk commands with credential injection and Build Info capture."
}

// GetAIDescription returns the extended description used by AI-assisted help.
func GetAIDescription() string {
	return `jf apk provides two modes of operation for working with Artifactory Alpine repositories:

1. jf apk upload — Publish a local .apk file to Artifactory.
   Performs a direct REST PUT (no native apk binary required). Infers arch
   and package metadata from the filename. Sets Artifactory properties and
   records a Build Info artifact.

2. jf apk <native-subcommand> — Wrap native apk commands with Artifactory
   credentials and Build Info capture.
   Injects HTTP_AUTH into the apk subprocess environment.
   For 'add' and 'upgrade', diffs the installed package list before/after
   to record dependencies into a Build Info alpine module.

To configure APK to use Artifactory as its repository, use: jf setup apk`
}

// GetArguments returns the argument reference shown in jf apk --help.
func GetArguments() string {
	return `	apk subcommand
		Subcommands:
		  upload  — publish a local .apk file to Artifactory
		  add     — install packages (Build Info + HTTP_AUTH)
		  upgrade — upgrade packages (Build Info + HTTP_AUTH)
		  update  — refresh index (HTTP_AUTH only)
		  fetch   — download .apk files (HTTP_AUTH only)
		  search  — search index (HTTP_AUTH only)
		  del     — remove packages (passthrough)
		  info    — query package info (passthrough)

		Common flags (all subcommands):
		  --server-id      JFrog server config ID (from jf c add). Default: active server.
		  --repo           Artifactory Alpine repository key.
		  --alpine-version Alpine release, e.g. v3.20.
		  --user           Override Artifactory username.
		  --password       Override Artifactory password or token.

		upload-specific flags:
		  --branch         Alpine repo branch. Default: main.
		  --arch           CPU architecture. Default: inferred from filename.
		  --build-name     Build Info name. Env fallback: JFROG_CLI_BUILD_NAME.
		  --build-number   Build Info number. Env fallback: JFROG_CLI_BUILD_NUMBER.
		  --project        JFrog Projects key. Env fallback: JFROG_CLI_BUILD_PROJECT.

		add/upgrade flags:
		  --build-name     Build Info name.
		  --build-number   Build Info number.
		  --project        JFrog Projects key.

		Examples:
		  jf setup apk --server-id my-server --repo my-alpine-repo
		  jf apk upload ./myapp-1.0.0-r0.x86_64.apk --repo my-alpine-repo --alpine-version v3.20
		  jf apk add curl bash --server-id my-server --repo my-alpine-repo \
		             --build-name ci-image --build-number 42
		  jf apk upgrade --server-id my-server --repo my-alpine-repo`
}
