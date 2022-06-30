package audit

var Usage = []string{"audit [command options]"}

func GetDescription() string {
	return "Execute an audit command, using the configured Xray details."
}
