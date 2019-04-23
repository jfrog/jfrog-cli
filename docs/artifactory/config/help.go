package config

const Description string = "Configure Artifactory details."

var Usage = []string{"jfrog rt c [command options] [server ID]",
	"jfrog rt c show [server ID]",
	"jfrog rt c [--interactive=<true|false>] delete <server ID>",
	"jfrog rt c [--interactive=<true|false>] clear"}

const Arguments string = `	server ID
		A unique ID for the new Artifactory server configuration.

	show
		Shows the stored configuration.
		In case this argument is followed by a configured server ID, then only this server's configurations is shown.

	delete
		This argument should be followed by a configured server ID. The configuration for this server ID will be deleted.

	clear
		Clears all stored configuration.`
