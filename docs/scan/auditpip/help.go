package auditpip

var Usage = []string{"audit-pip [command options]"}

func GetDescription() string {
	return "Execute an audit Pip command, using the configured Xray details."
}
