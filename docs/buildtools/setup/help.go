package setup

import (
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/setup"
	"strings"
)

var Usage = []string{"setup [command options]",
	"setup <package manager> [command options]"}

func GetDescription() string {
	return "An interactive command to configure your local package manager (e.g., npm, pip) itself to work with JFrog Artifactory, so its own commands resolve through it. " +
		"By default, settings persist in your user-level package manager configuration until you change them, and apply to every project you build as this user."
}

func GetArguments() string {
	return `	package manager
		The package manager to configure. Supported package managers are: ` + strings.Join(setup.GetSupportedPackageManagersList(), ", ") + "."
}

func GetAIDescription() string {
	return `Interactively (or non-interactively, via flags) configure a local package manager (` + strings.Join(setup.GetSupportedPackageManagersList(), ", ") + `) to resolve and/or publish through JFrog Artifactory. This is the fastest way to point an existing project at Artifactory without hand-editing the tool's native config files.

When to use:
- One-time setup of a package manager on a developer machine or CI runner to work against Artifactory.
- Switching an already-configured package manager to a different repository or server.

Not the same command as jf npm-config / jf mvn-config / jf pip-config, which look similar and do something else entirely:
- jf setup writes the package manager's OWN configuration (~/.npmrc, pip.conf, settings.xml, GOPROXY, the docker credential store, ...), so the native client resolves through Artifactory from then on. A plain 'npm install' or 'mvn install' is affected. Credentials are written into that file and stay there.
- jf <pm>-config writes .jfrog/projects/<pm>.yaml inside the project, holding a server ID and the resolve/deploy repositories. Only 'jf <pm>' commands read it - a plain 'npm install' is not affected at all - and the routing exists only while that command runs: jf npm swaps in a temporary .npmrc and restores the original, jf pip appends --index-url to the single invocation, jf go sets GOPROXY for the child process. No credentials are stored in the yaml. This is also what enables build-info collection (--build-name / --build-number, published with jf rt build-publish), which jf setup does not do.
- They are independent. Running one does not configure the other, and each can point at a different repository. Use jf setup to make the client itself work against Artifactory; use jf <pm>-config when you build through 'jf <pm>' and want build-info.
- Neither covers the same package managers: docker, podman, helm, twine and uv have setup but no -config command; conan, ruby and terraform have a -config command but no setup.

Prerequisites:
- A configured server (jf c add or jf login), or pass --url/--user/--password/--access-token directly.
- The Artifactory repository name for the package manager (a virtual repo where supported).

Common patterns:
  $ jf setup npm
  $ jf setup go --repo=go-remote
  $ jf setup docker --server-id=my-server

Gotchas:
- Without --repo, the command prompts for repository selection, so it is interactive by default; pass --repo for non-interactive/CI use.
- Configuration is applied globally to the package manager's native config (e.g., ~/.npmrc, NuGet.Config), affecting all projects on the machine, not just the current one.
- The change persists until it is changed again, and re-running for the same package manager replaces the previous repository and server rather than adding to them. A machine already pointed at one Artifactory will silently start resolving from the new one, so if the user works across several instances or projects, confirm which repository they want before running this.
- Some package managers let an environment variable move that configuration off its user-level default: PIP_CONFIG_FILE (pip, pipenv), NPM_CONFIG_USERCONFIG (npm), POETRY_CONFIG_DIR (poetry), UV_CONFIG_FILE (uv), GOENV (go) and GRADLE_USER_HOME (gradle). When one of these is set the settings are written there instead, so their scope follows that path rather than the whole machine, and the command reports the path it used. pnpm is not in that list: "pnpm config set" writes to pnpm's own config directory and ignores NPM_CONFIG_USERCONFIG.
- Some package managers share one configuration file, so setting up the second one overwrites the first even though they are different package managers: pip and pipenv both write pip.conf, and nuget and dotnet both write NuGet.Config. "jf setup pip --repo a" followed by "jf setup pipenv --repo b" leaves pip resolving from b. For the same reason, do not run those two setups concurrently. npm, pnpm and Yarn each write their own file, as do docker and podman.
- pnpm and npm can end up on different repositories without any warning. pnpm reads its own configuration first and ~/.npmrc only as a fallback, so a machine with no pnpm setup follows "jf setup npm", but once "jf setup pnpm" has run, a later "jf setup npm --repo b" moves npm alone and pnpm keeps resolving from the repository it was given. If both are in use, run "jf setup" for both.
- maven and gradle do not need their client installed: their setup writes settings.xml and a Gradle init script directly, so it works on a machine that only has ./mvnw or ./gradlew. Every other package manager's setup runs its client. helm additionally needs 3.8.0 or newer, because its login targets an OCI registry.
- docker/podman authenticate directly against the registry and skip the repository prompt entirely (no --repo needed); helm still goes through repository selection like the other package managers even though its login step doesn't end up using the repo name.

Related: jf npm-config, jf go-config, jf pip-config, jf c add`
}
