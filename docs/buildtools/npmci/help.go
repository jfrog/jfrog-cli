package npmci

import "github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"

var Usage = []string{"npm ci [npm ci args] [command options]"}

func GetDescription() string {
	return `Run npm ci, using the npm repository, configured by the '` + coreutils.GetCliExecutableName() + ` npmc' command.`
}

func GetArguments() string {
	return `	npm ci args
		The npm ci args to run npm ci.`
}

func GetAIDescription() string {
	return `Run 'npm ci' through JFrog: dependencies resolve via the Artifactory npm virtual repo configured by 'jf npm-config'. Strictly uses package-lock.json (will fail if it does not match package.json). Captures build-info when --build-name and --build-number are passed.

When to use:
- Reproducible installs in CI (no dependency hoisting, no version drift).
- Producing build-info for an immutable install step.

Prerequisites:
- A local npm binary.
- 'jf npm-config' run once.
- A configured server.
- A valid package-lock.json checked into the repo.

Common patterns:
  $ jf npm ci
  $ jf npm ci --build-name=my-app --build-number=42

Gotchas:
- Fails if package-lock.json is missing or stale relative to package.json.
- Deletes node_modules before installing; slower than 'npm install' but deterministic.

Related: jf npm install, jf npm publish, jf npm-config`
}
