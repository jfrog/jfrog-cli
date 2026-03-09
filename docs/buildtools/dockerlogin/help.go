package dockerlogin

var Usage = []string{"docker login [registry] [command options]",
	"docker login",
	"docker login --server-id jfrog-server",
	"docker login my-docker-registry.jfrog.io --server-id jfrog-server",
	"docker login my-docker-registry.jfrog.io --username my-username --password my-password"}

func GetDescription() string {
	return `Log in to an Artifactory Docker registry`
}

func GetArguments() string {
	return `	You can optionally specify a registry. If not provided, the registry from the JFrog config server-id is used.
	This argument is mandatory when logging in with a username and password.`
}
