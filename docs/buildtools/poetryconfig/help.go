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

Related: jf poetry, jf pip-config`
}
