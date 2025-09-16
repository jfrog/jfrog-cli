package dockerlogin

var Usage = []string{"docker login [registry] [command options]",
	"docker login",
	"docker login --server-id jfrog-server",
	"docker login my-docker-registry.jfrog.io --server-id jfrog-server",
	"docker login my-docker-registry.jfrog.io --username my-username --password my-password"}

func GetDescription() string {
	return `Login to a artifactory Docker registry`
}

func GetArguments() string {
	return `	Command accepts optional registry for login. If not provided, the registry from the jfrog config server-id will be used.
	This argument is mandatory when logging in using username and password.`
}
