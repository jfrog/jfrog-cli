package pipconfig

var Usage = []string{"pip-config"}

func GetDescription() string {
	return "Generate pip build configuration."
}

func GetAIDescription() string {
	return `Write a per-project pip configuration (.jfrog/projects/pip.yaml) so 'jf pip install' resolves through an Artifactory PyPI repository.

When to use:
- First-time setup of a Python project to use a private PyPI index.

Prerequisites:
- A configured server.
- The Artifactory PyPI repository key.

Common patterns:
  $ jf pip-config --server-id-resolve=my-server --repo-resolve=pypi-virtual

Gotchas:
- Interactive prompts trigger when required flags are missing.
- Affects only 'jf pip'; native pip invocations still use the system index.
- This does not configure the pip client itself. It is read only by 'jf pip' commands; a plain 'pip install' keeps resolving from its own configuration, which this command never touches. To point the client itself at Artifactory for every project on the machine, run 'jf setup pip' instead - the two are independent and can even name different repositories.

Related: jf pip, jf pipenv-config, jf poetry-config, jf setup pip`
}
