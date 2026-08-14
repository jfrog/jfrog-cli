package twinedocs

var Usage = []string{"twine <twine arguments> [command options]"}

func GetDescription() string {
	return "Run twine"
}

func GetArguments() string {
	return `	twine commands
		Arguments and options for the twine command.`
}

func GetAIDescription() string {
	return `Run a Twine command (upload) through JFrog to publish Python distributions to an Artifactory PyPI repository, with optional build-info collection.

When to use:
- Uploading wheels or sdists to a private PyPI repo.

Prerequisites:
- A local twine binary (pip install twine).
- 'jf pip-config' or equivalent has the deploy repo set.
- A configured server.
- Built distributions in dist/ (use 'jf poetry build' or python setup.py sdist bdist_wheel).

Common patterns:
  $ jf twine upload dist/*
  $ jf twine upload dist/* --build-name=my-pkg --build-number=1

Gotchas:
- The Artifactory PyPI repo must be the destination; the command does not auto-detect.
- Twine uses credentials from .pypirc; setting them via jf config is not automatic.

Related: jf poetry, jf pip, jf rt build-publish`
}
