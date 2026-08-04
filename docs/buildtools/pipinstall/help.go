package pipinstall

var Usage = []string{"pip <pip arguments> [command options]"}

func GetDescription() string {
	return "Run pip install."
}

func GetArguments() string {
	return `	pip sub-command
		Arguments and options for the pip command.`
}

func GetAIDescription() string {
	return `Run pip install (or another pip subcommand) through JFrog, routing package downloads via an Artifactory PyPI virtual repository, with optional build-info collection.

When to use:
- Installing Python dependencies through a private PyPI repo (caching, curation, isolation).

Prerequisites:
- A local pip on PATH (in the active venv or system).
- 'jf pip-config' run once in the project directory.
- A configured server.

Common patterns:
  $ jf pip install -r requirements.txt
  $ jf pip install -r requirements.txt --build-name=my-svc --build-number=5

Gotchas:
- 'jf pip-config' must be run first; otherwise pip uses the public PyPI index.
- Always activate the target virtualenv before running 'jf pip', or pass --user explicitly.
- Curation policies (if enabled) can block install on packages flagged by Xray.

Related: jf pip-config, jf pipenv, jf poetry, jf twine`
}
