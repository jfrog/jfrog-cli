package delete

const Description string = `Deletes the stored configuration. In case this argument is followed by a configured server ID, then only this server's configurations is deleted.`

var Usage = []string{"jfrog rt c del",
	"jfrog rt c del <server ID>"}

const Arguments string = `	server ID
		A unique ID for an existing JFrog configuration.`
