package sync

var Usage = []string{"pl sync <repository name>"}

func GetDescription() string {
	return "Sync a pipeline resource."
}

func GetArguments() string {
	return `	repository name
		Full repository name of the pipeline resource.
	branch name
		Branch name to trigger sync on.`
}
