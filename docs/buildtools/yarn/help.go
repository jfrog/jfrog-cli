package yarn

var Usage = []string{"yarn [yarn command] [command options]"}

func GetDescription() string {
	return "Run Yarn commands."
}

func GetAIDescription() string {
	return `Run a Yarn command through JFrog: resolves dependencies via an Artifactory npm/yarn virtual repo, optionally collecting build-info. Wraps the local yarn binary.

When to use:
- Installing Yarn dependencies through an Artifactory virtual repo.
- Producing JFrog build-info for Yarn projects.

Prerequisites:
- A local yarn binary on PATH (Yarn 1.x and Yarn berry both supported).
- 'jf yarn-config' run once in the project directory.
- A configured server.

Common patterns:
  $ jf yarn install
  $ jf yarn install --build-name=my-app --build-number=7
  $ jf yarn add lodash

Gotchas:
- 'jf yarn-config' must be run first; the command fails without it.
- --build-name and --build-number are needed together for build-info.
- Yarn berry has different config semantics; verify the .yarnrc.yml after config.

Related: jf yarn-config, jf npm, jf pnpm`
}
