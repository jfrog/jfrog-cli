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

func GetAIDescription() string {
	return `Authenticate the local docker daemon against an Artifactory Docker registry, reusing credentials from the configured jf server. Skips re-typing username/password and avoids leaving them in shell history.

When to use:
- Setting up Docker auth on a developer machine after 'jf c add'.
- CI runners that need a docker login before push/pull.

Prerequisites:
- A configured server (jf c add or jf login).
- A docker daemon running.

Common patterns:
  $ jf docker login --server-id=my-server
  $ jf docker login my-docker-registry.jfrog.io --server-id=my-server
  $ jf docker login my-docker-registry.jfrog.io --username=alice --password=secret

Gotchas:
- The registry host must be a Docker repository host on the platform; the platform base URL is NOT a docker registry.
- Credentials are written into the local docker config (~/.docker/config.json) in plaintext or via a credential helper.

Related: jf docker push, jf docker pull, jf c add`
}
