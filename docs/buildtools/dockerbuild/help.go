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
