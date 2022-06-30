package auditpipenv

var Usage = []string{"audit-pipenv [command options]"}

func GetDescription() string {
	return "Execute an audit Pipenv command, using the configured Xray details."
}
