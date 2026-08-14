package npminstall

import "github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"

var Usage = []string{"npm install [npm install args] [command options]"}

func GetDescription() string {
	return `Run npm install, using the npm repository, configured by the '` + coreutils.GetCliExecutableName() + ` npmc' command.`
}

func GetArguments() string {
	return `	npm install args
		The npm install args to run npm install. For example, --global.`
}

func GetAIDescription() string {
	return `Run 'npm install' through JFrog: dependencies resolve via the Artifactory npm virtual repo configured by 'jf npm-config'. Captures build-info when --build-name and --build-number are passed.

When to use:
- Installing npm dependencies through a private Artifactory npm repo.
- Producing build-info from an npm install step.

Prerequisites:
- A local npm binary.
- 'jf npm-config' run once in the project directory.
- A configured server.

Common patterns:
  $ jf npm install
  $ jf npm install --build-name=my-app --build-number=42
  $ jf npm install lodash --save

Gotchas:
- 'jf npm-config' must be run first.
- --global behaves like npm's --global flag; the install goes outside the project.

Related: jf npm-config, jf npm ci, jf npm publish`
}
