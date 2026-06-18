package uvcommand

var Usage = []string{"uv <args> [command options]"}

func GetDescription() string {
	return "Run uv command"
}

func GetAIDescription() string {
	return `Run a uv (Python package manager) command through JFrog CLI so dependencies resolve against Artifactory and build-info can be recorded. Arguments are forwarded to the uv binary unchanged; the jf-specific flags are the build-info and server options.

When to use:
- Running 'uv pip install' / 'uv sync' / 'uv add' in a project that resolves through an Artifactory-backed index.
- Capturing build-info for a uv-based Python build by passing --build-name and --build-number.

Prerequisites:
- uv installed and available on PATH.
- A configured JFrog Platform server (jf c add or jf login), or pass --server-id.

Common patterns:
  $ jf uv pip install -r requirements.txt
  $ jf uv sync --build-name=my-app --build-number=7
  $ jf uv add requests --build-name=my-app --build-number=7 --module=api

Gotchas:
- All arguments after the uv sub-command are passed straight to uv (the command uses SkipFlagParsing); only --build-name, --build-number, --module, --project, and --server-id are interpreted by jf.
- Build-info is collected only when --build-name and --build-number are provided; publish it afterwards with 'jf rt build-publish'.

Related: jf rt build-publish, jf pip-install`
}

func GetArguments() string {
	return `	sub-command
		Arguments and options for the uv command.`
}
