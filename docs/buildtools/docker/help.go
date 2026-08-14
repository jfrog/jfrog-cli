package docker

var Usage = []string{"docker <docker arguments> [command options]"}

func GetDescription() string {
	return `Run any docker command, including ‘jf docker scan’ for scanning a local image with JFrog Xray.`
}

func GetArguments() string {
	return `	login                       Log in to an Artifactory Docker registry.
	build                       Run docker build.
	push                        Run docker push.
	pull                        Run docker pull.
	scan                        Scan a local Docker image for security vulnerabilities with JFrog Xray.`
}

func GetAIDescription() string {
	return `Wrap the local docker binary for build, push, pull, login, and scan subcommands. Push and pull can collect build-info; scan delegates to Xray for a local image security scan.

When to use:
- Building and pushing Docker images to an Artifactory Docker registry with build-info.
- Logging into an Artifactory Docker registry using the configured server credentials.
- Scanning a locally built image with Xray before publishing.

Prerequisites:
- A local docker daemon and CLI.
- A configured server (for login/push/pull, the server's Docker registry URL must be set up).
- For scan: Xray reachable from the configured server.

Common patterns:
  $ jf docker login my-docker-registry.jfrog.io --server-id=my-server
  $ jf docker build -t my-docker-registry.jfrog.io/my-image:1.0 .
  $ jf docker push my-docker-registry.jfrog.io/my-image:1.0 --build-name=my-build --build-number=1
  $ jf docker pull my-docker-registry.jfrog.io/my-image:1.0 --build-name=my-build --build-number=1
  $ jf docker scan my-image:1.0

Gotchas:
- 'docker scan' is a JFrog extension, not the docker upstream 'scan' subcommand; behaviors differ.
- Push/pull collect build-info only when --build-name and --build-number are present.
- docker buildx (BuildKit) is supported but multi-platform manifests need extra care with build-info.

Related: jf dl, jf docker-promote, jf docker scan`
}
