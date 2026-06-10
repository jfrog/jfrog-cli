package gopublish

var Usage = []string{"gp [command options] <project version>"}

func GetDescription() string {
	return "Publish go package and/or its dependencies to Artifactory."
}

func GetArguments() string {
	return `	project version
		Package version to be published.`
}

func GetAIDescription() string {
	return `Publish a Go module (and optionally its dependencies) to an Artifactory Go repository. The version argument is the semver tag attached to the module in the repository.

When to use:
- Releasing a new version of a Go module to a private Artifactory Go repo.

Prerequisites:
- A configured server.
- 'jf go-config' must be run with --repo-deploy set.
- A go.mod in the project root.

Common patterns:
  $ jf gp v1.2.3
  $ jf gp v1.2.3 --build-name=my-svc --build-number=12

Gotchas:
- The version must follow Go module versioning (vMAJOR.MINOR.PATCH); without 'v' prefix go tooling will reject it.
- Republishing the same version is rejected by Artifactory's Go repository.

Related: jf go, jf go-config, jf rt build-publish`
}
