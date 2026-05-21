package pnpmcommand

var Usage = []string{"pnpm <pnpm arguments> [command options]"}

func GetDescription() string {
	return "Run pnpm command."
}

func GetArguments() string {
	return `	install, i                Run pnpm install.
	publish                   Packs and deploys the pnpm package to the designated npm repository.
	help, h`
}

func GetAIDescription() string {
	return `Run pnpm install/publish through JFrog: resolves through an Artifactory npm virtual repo and publishes to a local repo. Wraps the local pnpm binary.

When to use:
- Installing or publishing pnpm-managed packages via Artifactory.

Prerequisites:
- A local pnpm binary on PATH.
- 'jf pnpm-config' run once in the project directory.
- A configured server.

Common patterns:
  $ jf pnpm install
  $ jf pnpm install --build-name=my-app --build-number=3
  $ jf pnpm publish

Gotchas:
- 'jf pnpm-config' must be run first.
- pnpm's workspace mode (monorepo) requires careful config; verify pnpm-workspace.yaml is respected.

Related: jf pnpm-config, jf npm, jf yarn`
}
