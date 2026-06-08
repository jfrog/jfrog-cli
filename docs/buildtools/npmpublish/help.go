package npmpublish

import "github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"

var Usage = []string{"npm publish [command options]"}

func GetDescription() string {
	return `Packs and deploys the npm package to the Artifactory npm repository, configured by the '` + coreutils.GetCliExecutableName() + ` npmc' command.`
}

func GetAIDescription() string {
	return `Build and publish the current npm package to the Artifactory npm repository configured by 'jf npm-config' (--repo-deploy). Captures build-info when --build-name and --build-number are passed.

When to use:
- Releasing a new version of an internal npm package to a private repo.

Prerequisites:
- A local npm binary.
- 'jf npm-config' must have --repo-deploy set.
- A configured server.
- A package.json with a unique version (republish of same version may fail by repo policy).

Common patterns:
  $ jf npm publish
  $ jf npm publish --build-name=my-app --build-number=42

Gotchas:
- The version in package.json must be incremented; otherwise the publish is rejected.
- 'npm pack' is implicit; control output with the standard npmignore / files fields in package.json.

Related: jf npm install, jf npm-config, jf rt build-publish`
}
