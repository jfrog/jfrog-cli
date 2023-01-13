package sync

var Usage = []string{"pl sync <repository name>"}

func GetDescription() string {
	return "Sync a pipeline resource."
}

func GetArguments() string {
	return `	repository name
	full repository name of the pipeline resource
	branch name
	branch name to trigger sync on
`
}
