package auditpipenv

var Usage = []string{"xr audit-pipenv [command options]"}

func GetDescription() string {
	return "Execute an audit Pipenv command, using the configured Xray details."
}
