package pipenvconfig

var Usage = []string{"pipenv-config"}

func GetDescription() string {
	return "Generate pipenv build configuration."
}

func GetAIDescription() string {
	return `Write a per-project pipenv configuration (.jfrog/projects/pipenv.yaml) so 'jf pipenv install' resolves through an Artifactory PyPI repository.

When to use:
- Initial setup of a pipenv project to use a private PyPI index.

Prerequisites:
- A configured server.
- The Artifactory PyPI repository key.

Common patterns:
  $ jf pipenv-config --server-id-resolve=my-server --repo-resolve=pypi-virtual

Gotchas:
- Interactive prompts trigger when required flags are missing.
- This does not configure the pipenv client itself. It is read only by 'jf pipenv' commands; a plain 'pipenv install' keeps resolving from its own configuration, which this command never touches. To point the client itself at Artifactory for every project on the machine, run 'jf setup pipenv' instead - the two are independent and can even name different repositories.

Related: jf pipenv, jf pip-config, jf poetry-config, jf setup pipenv`
}
