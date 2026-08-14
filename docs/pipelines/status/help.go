package status

var Usage = []string{"pl status"}

func GetDescription() string {
	return "Fetch the latest pipeline run status."
}

func GetAIDescription() string {
	return `Show the latest run status for one or more JFrog Pipelines. Filters by pipeline name and/or branch. Useful for polling a build from CI or surfacing the last run from a shell.

When to use:
- Verifying a pipeline finished before triggering a downstream job.
- Checking the latest status without opening the UI.

Prerequisites:
- A configured server with the pipelines URL set (jf c add captures it from the platform URL).
- Pipelines must be reachable from the calling network.

Common patterns:
  $ jf pl status --pipeline-name=my-pipeline --branch=main
  $ jf pl status --pipeline-name=my-pipeline --monitor
  $ jf pl status --pipeline-name=my-pipeline --format=json

Gotchas:
- Without --pipeline-name the command lists all pipelines visible to the caller, which can be slow.
- --monitor blocks until status changes; use carefully in CI.
- --single-branch limits to a single-branch pipeline; default is multi-branch.

Related: jf pl trigger, jf pl sync, jf pl sync-status`
}
