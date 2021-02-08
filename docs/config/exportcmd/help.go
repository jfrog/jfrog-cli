package exportcmd

const Description string = `Creates a server configuration token. The generated token can be imported by the "jfrog rt c import <Server token>" command.`

var Usage = []string{"jfrog rt c export <server token>"}

const Arguments string = `	server ID
		A unique ID for the new server configuration.`
