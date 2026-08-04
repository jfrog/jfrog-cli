package poetryconfig

var Usage = []string{"poetry-config"}

func GetDescription() string {
	return "Generate poetry build configuration."
}

func GetAIDescription() string {
	return `Write a per-project Poetry configuration (.jfrog/projects/poetry.yaml) so 'jf poetry' resolves through an Artifactory PyPI repository.

When to use:
- Initial setup of a Poetry project to use a private PyPI index.

Prerequisites:
- A configured server.
- The Artifactory PyPI repository key.

Common patterns:
  $ jf poetry-config --server-id-resolve=my-server --repo-resolve=pypi-virtual

Gotchas:
- Interactive prompts trigger when required flags are missing.
- Updates the Poetry sources list in pyproject.toml; review the diff after running.
- This does not configure the poetry client itself. It is read only by 'jf poetry' commands; a plain 'poetry add' keeps resolving from its own configuration, which this command never touches. To point the client itself at Artifactory for every project on the machine, run 'jf setup poetry' instead - the two are independent and can even name different repositories.

Related: jf poetry, jf pip-config, jf setup poetry`
}
