package dockerpush

var Usage = []string{"docker push <image tag> [command options]"}

func GetDescription() string {
	return `Run Docker push command.`
}

func GetArguments() string {
	return `	docker push args
		The docker push args to run docker push.`
}

func GetAIDescription() string {
	return `Push a Docker image to an Artifactory Docker registry and optionally record it in build-info.

When to use:
- Publishing built images to a private Docker repo.
- Capturing the image digest in build-info for downstream traceability.

Prerequisites:
- A local docker daemon.
- 'jf docker login' or equivalent docker credentials.
- For build-info: --build-name and --build-number together.

Common patterns:
  $ jf docker push my-docker-registry.jfrog.io/my-image:1.0
  $ jf docker push my-docker-registry.jfrog.io/my-image:1.0 --build-name=my-build --build-number=1

Gotchas:
- Image tag must include the full registry host that matches the configured server's Docker registry.
- buildx multi-platform manifests push differently; prefer 'jf docker build' with --push for those.

Related: jf docker pull, jf docker build, jf rt build-publish`
}
