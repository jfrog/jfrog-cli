package remove

var Usage = []string{"config rm",
	"config rm <server ID>"}

func GetDescription() string {
	return `Removes the stored configuration. In case this argument is followed by a configured server ID, then only this server's configurations is removed.`
}

func GetArguments() string {
	return `	server ID
		A unique ID for an existing JFrog configuration.`
}
