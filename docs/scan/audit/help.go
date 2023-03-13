package audit

var Usage = []string{"audit [command options]"}

func GetDescription() string {
	return "Audit your local project's dependencies, using the configured Xray details."
}
