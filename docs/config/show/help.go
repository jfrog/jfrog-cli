package show

const Description string = `Shows the stored configuration. In case this argument is followed by a configured server ID, then only this server's configurations is shown.`

var Usage = []string{"jfrog config show <server ID>"}

const Arguments string = `	server ID
		The configured server ID.`
