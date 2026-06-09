package dockerbuild

var Usage = []string{"docker build [command options] <docker build arguments>",
	"docker build -t my-image:tag .",
	"docker build -t my-image:tag -f Dockerfile .",
	"docker build -t my-image:tag --push .",
	"docker build -t my-image:tag --build-name my-build --build-number 1 .",
	"docker buildx build -t my-image:tag --platform linux/amd64 --push ."}

func GetDescription() string {
	return `Run Docker build command with build-info collection support.`
}

func GetArguments() string {
	return `	docker build args
		The docker build arguments to run docker build. Standard Docker build arguments are supported.

		All standard Docker build arguments are required and supported and are passed through to the Docker daemon.
		The command also supports 'docker buildx build' for multi-platform builds.`
}

func GetAIDescription() string {
	return `Build a Docker image (or buildx multi-platform image) and capture build-info for traceability. Wraps 'docker build' and 'docker buildx build'.

When to use:
- Building images that need to be tracked alongside other artifacts in a JFrog build.
- Multi-platform builds via buildx with --push to Artifactory.

Prerequisites:
- A local docker (or buildx) daemon.
- For build-info: --build-name and --build-number passed together.

Common patterns:
  $ jf docker build -t my-image:1.0 .
  $ jf docker build -t my-image:1.0 --build-name=my-build --build-number=1 .
  $ jf docker buildx build -t my-image:1.0 --platform linux/amd64,linux/arm64 --push .

Gotchas:
- Build-info captures layers and image digest at build time; subsequent pushes are tracked separately.
- buildx multi-platform builds require --push (cannot --load into local daemon).

Related: jf docker push, jf docker pull, jf rt build-publish`
}
