package show

const Description string = `Shows the stored configuration. In case this argument is followed by a configured server ID, then only this server's configurations is shown.`

var Usage = []string{"jfrog rt c show <server ID>"}

const Arguments string = `	server ID
		A unique ID for the new server configuration.`
