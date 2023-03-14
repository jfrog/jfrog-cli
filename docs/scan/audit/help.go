package audit

var Usage = []string{"audit [command options]"}

func GetDescription() string {
	return "Audit your local project's dependencies by generating a dependency tree and scanning it with Xray."
}
