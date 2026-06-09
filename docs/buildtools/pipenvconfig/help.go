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

Related: jf pipenv, jf pip-config, jf poetry-config`
}
