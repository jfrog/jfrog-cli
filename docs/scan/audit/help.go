package audit

var Usage = []string{"audit [command options]"}

func GetDescription() string {
	return "Audit your local project's dependencies by generating a dependency tree for the sources, and scans it with Xray. The command should be executed while inside the root directory of the project."
}
