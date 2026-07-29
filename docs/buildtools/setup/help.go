package setup

import (
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/setup"
	"strings"
)

var Usage = []string{"setup [command options]",
	"setup <package manager> [command options]"}

func GetDescription() string {
	// Kept to roughly the length of the longest existing command description: this string
	// is printed unwrapped on a single line in the `jf --help` command list. The
	// consequences of re-running the command live in GetAIDescription, which is rendered
	// as a multi-line block where length is not a constraint.
	return "An interactive command to configure your local package manager (e.g., npm, pip) to work with JFrog Artifactory. " +
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
- Some package managers let an environment variable move that configuration off its user-level default: PIP_CONFIG_FILE (pip, pipenv), NPM_CONFIG_USERCONFIG (npm, pnpm), POETRY_CONFIG_DIR and UV_CONFIG_FILE. When one of these is set the settings are written there instead, so their scope follows that file rather than the whole machine, and the command reports the path it used.
- docker/podman authenticate directly against the registry and skip the repository prompt entirely (no --repo needed); helm still goes through repository selection like the other package managers even though its login step doesn't end up using the repo name.

Related: jf npm-config, jf go-config, jf pip-config, jf c add`
}
