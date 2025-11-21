package helmcommand

var Usage = []string{"helm <helm arguments> [command options]"}

func GetDescription() string {
	return "Run native Helm command"
}

func GetArguments() string {
	return `	install                  Install a chart.
	upgrade                  Upgrade a release.
	package                  Package a chart directory into a chart archive.
	push                     Push a chart to remote.
	pull                     Download a chart from remote.
	repo                     Add, list, remove, update, and index chart repositories.
	dependency               Manage a chart's dependencies.
	help, h                  Show help for any command.`
}
