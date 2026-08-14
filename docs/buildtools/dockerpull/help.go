package dockerpull

var Usage = []string{"docker pull <image tag> [command options]"}

func GetDescription() string {
	return `Run Docker pull command.`
}

func GetArguments() string {
	return `	docker pull args
		The docker pull args to run docker pull.`
}

func GetAIDescription() string {
	return `Pull a Docker image from an Artifactory Docker registry and optionally record it in build-info.

When to use:
- Pulling images from a private registry with build-info attached.
- Tracking what was pulled during a CI job.

Prerequisites:
- A local docker daemon.
- 'jf docker login' or equivalent docker credentials.
- For build-info: --build-name and --build-number together.

Common patterns:
  $ jf docker pull my-docker-registry.jfrog.io/my-image:1.0
  $ jf docker pull my-docker-registry.jfrog.io/my-image:1.0 --build-name=my-build --build-number=1

Gotchas:
- Without authentication to the registry, the daemon falls back to anonymous and fails on private repos.
- Image must be tagged with the full registry host for jf to associate it with the right server.

Related: jf docker push, jf docker login`
}
