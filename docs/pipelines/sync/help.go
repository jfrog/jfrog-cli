package sync

var Usage = []string{"pl sync"}

func GetDescription() string {
	return "Sync a pipeline resource."
}

func GetArguments() string {
	return `	repository name
		Full repository name of the pipeline resource.
	branch name
		Branch name to trigger sync on.`
}

func GetAIDescription() string {
	return `Trigger a sync of a pipeline resource (re-read the pipeline definition from the source repository for the given branch). Used after editing pipeline YAML or moving branches.

When to use:
- Forcing a re-load of pipelines.yml after a configuration change.
- Recovering from a stale sync (pipeline UI shows old config).

Prerequisites:
- A configured server with the pipelines URL set.
- The repository and branch must already be wired into Pipelines as a pipeline source.

Common patterns:
  $ jf pl sync my-org/my-repo main
  $ jf pl sync my-org/my-repo feature/x --format=json

Gotchas:
- The repository argument must be the full <owner>/<name> form, matching how the source is registered.
- Sync runs asynchronously. Use 'jf pl sync-status' to confirm completion.

Related: jf pl sync-status, jf pl status, jf pl trigger`
}
