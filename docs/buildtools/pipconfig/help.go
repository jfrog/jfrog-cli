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

Related: jf pip, jf pipenv-config, jf poetry-config`
}
