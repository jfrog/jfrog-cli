package npmcommand

var Usage = []string{"npm <npm arguments> [command options]"}

func GetDescription() string {
	return "Run npm command."
}

func GetArguments() string {
	return `	ci                        Run npm ci.
	publish, p                Packs and deploys the npm package to the designated npm repository.
	install, i, isntall, add  Run npm install.
	help, h`
}

func GetAIDescription() string {
	return `Run an npm command through JFrog: install/ci routes through an Artifactory virtual repo, publish deploys to a target repo, and build-info can be collected. Wraps the local npm binary; arguments after the command are passed through.

When to use:
- Installing dependencies via an Artifactory npm virtual repo (caches public registries, enforces curation).
- Publishing an npm package to an Artifactory npm-local repo.
- Producing JFrog build-info for npm projects.

Prerequisites:
- A local npm binary on PATH.
- 'jf npm-config' run once in the project directory.
- A configured server.

Common patterns:
  $ jf npm install
  $ jf npm ci --build-name=my-app --build-number=42
  $ jf npm publish --build-name=my-app --build-number=42
  $ jf npm install --omit=dev

Gotchas:
- 'jf npm-config' must be run first; without it, npm commands resolve through the public registry, defeating the point.
- --build-name and --build-number are needed together for build-info.
- Some flags conflict between the npm version and the JFrog wrapper; in case of doubt, use --command-name to be explicit.
- Curation (if enabled) can block install on policy-failed packages; surface those errors with --verbose.

Related: jf npm-config, jf rt build-publish, jf audit`
}
