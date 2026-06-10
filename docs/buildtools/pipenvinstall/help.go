package pipenvinstall

var Usage = []string{"pipenv <pipenv arguments> [command options]"}

func GetDescription() string {
	return "Run pipenv install."
}

func GetArguments() string {
	return `	pipenv sub-command
		Arguments and options for the pipenv command.`
}

func GetAIDescription() string {
	return `Run pipenv install through JFrog, routing dependency resolution via an Artifactory PyPI repository.

When to use:
- Installing Python dependencies in a pipenv-managed project that should resolve through Artifactory.

Prerequisites:
- A local pipenv binary.
- 'jf pipenv-config' run once in the project directory.
- A configured server.

Common patterns:
  $ jf pipenv install
  $ jf pipenv install --dev --build-name=my-svc --build-number=8

Gotchas:
- 'jf pipenv-config' must be run first.
- pipenv writes its own Pipfile.lock; ensure the lockfile is committed and updated through 'jf pipenv'.

Related: jf pipenv-config, jf pip, jf poetry`
}
