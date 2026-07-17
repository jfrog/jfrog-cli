package poetry

var Usage = []string{"poetry <poetry args> [command options]"}

func GetDescription() string {
	return "Run poetry command"
}

func GetArguments() string {
	return `	poetry sub-command
		Arguments and options for the poetry command.`
}

func GetAIDescription() string {
	return `Run a Poetry command (install, build, publish) through JFrog, routing dependency resolution via an Artifactory PyPI repository, with optional build-info.

When to use:
- Installing dependencies in a Poetry project against a private PyPI repo.
- Publishing a Poetry-built wheel to Artifactory.

Prerequisites:
- A local poetry binary.
- 'jf poetry-config' run once in the project directory.
- A configured server.

Common patterns:
  $ jf poetry install
  $ jf poetry build
  $ jf poetry publish --build-name=my-pkg --build-number=1

Gotchas:
- 'jf poetry-config' must be run first.
- Poetry's lockfile must be regenerated through 'jf poetry lock' to match the routed source.

Related: jf poetry-config, jf pip, jf twine`
}
