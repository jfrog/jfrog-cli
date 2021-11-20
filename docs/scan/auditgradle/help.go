package auditgradle

var Usage = []string{"audit-gradle [command options]"}

func GetDescription() string {
	return "Execute an audit Gradle command, using the configured Xray details."
}
