package show

var Usage = []string{"config show <server ID>"}

func GetDescription() string {
	return `Shows the stored configuration. In case this argument is followed by a configured server ID, then only this server's configurations is shown.`
}
