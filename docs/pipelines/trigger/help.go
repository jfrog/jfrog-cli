package trigger

var Usage = []string{"pl trigger"}

func GetDescription() string {
	return "Trigger a manual pipeline run."
}

func GetArguments() string {
	return `	pipeline name
		Pipeline name to trigger the manual run on.
	branch name
		Branch name to trigger the manual run on.`
}

func GetAIDescription() string {
	return `Trigger a manual run of a JFrog Pipeline on a specific branch. Returns 200 OK on success (the actual run executes asynchronously on the pipelines server).

When to use:
- Kicking off a build from a script or chat-ops automation.
- Re-running a pipeline after a configuration change.

Prerequisites:
- A configured server with the pipelines URL set.
- The pipeline must be visible to the caller and configured for manual triggering on the specified branch.

Common patterns:
  $ jf pl trigger my-pipeline main
  $ jf pl trigger my-pipeline feature/foo --single-branch
  $ jf pl trigger my-pipeline main --format=json

Gotchas:
- The CLI returns immediately once the trigger request is accepted; the run is not awaited. Poll 'jf pl status' to track it.
- --single-branch must match the pipeline's branch model; mismatched values produce a 4xx from the server.

Related: jf pl status, jf pl sync, jf pl version`
}
