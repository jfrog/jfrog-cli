package trigger

var Usage = []string{"pl trigger"}

func GetDescription() string {
	return "Trigger a manual pipeline run."
}

func GetArguments() string {
	return `	pipeline name
	pipeline name to trigger manual run
	branch name
	branch name to trigger manual run`
}
