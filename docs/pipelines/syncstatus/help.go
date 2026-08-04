package syncstatus

var Usage = []string{"pl sync-status"}

func GetDescription() string {
	return "Fetch pipeline resource sync status."
}

func GetAIDescription() string {
	return `Report the last sync status for a pipeline resource (repository + branch). Use after 'jf pl sync' to confirm the new definition is loaded, or to inspect a stuck sync.

When to use:
- Confirming a 'jf pl sync' completed cleanly.
- Diagnosing why a pipeline is running an old definition.

Prerequisites:
- A configured server with the pipelines URL set.
- --branch and --repository are both mandatory.

Common patterns:
  $ jf pl sync-status --repository=my-org/my-repo --branch=main
  $ jf pl sync-status --repository=my-org/my-repo --branch=main --format=json

Gotchas:
- Both --branch and --repository must be passed; the command errors otherwise.
- The status reflects the last sync attempt, not the current pipeline run state (use 'jf pl status' for that).

Related: jf pl sync, jf pl status, jf pl trigger`
}
