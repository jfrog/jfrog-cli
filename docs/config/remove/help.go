package remove

const Description string = `Removes the stored configuration. In case this argument is followed by a configured server ID, then only this server's configurations is removed.`

var Usage = []string{"jfrog config rm",
	"jfrog config rm <server ID>"}

const Arguments string = `	server ID
		A unique ID for an existing JFrog configuration.`
