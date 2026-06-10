package gocommand

var Usage = []string{"go <go arguments> [command options]"}

func GetDescription() string {
	return "Runs go."
}

func GetArguments() string {
	return `	go commands
		Arguments and options for the go command.`
}

func GetAIDescription() string {
	return `Run a go command (build, test, mod download) through JFrog: module downloads resolve via an Artifactory Go repository (GOPROXY), with optional build-info collection.

When to use:
- Building Go projects that should resolve modules from a private Artifactory Go repo.
- Producing build-info for Go modules.

Prerequisites:
- A local go toolchain.
- 'jf go-config' run once in the project directory.
- A configured server.

Common patterns:
  $ jf go build ./...
  $ jf go test ./...
  $ jf go build --build-name=my-svc --build-number=12

Gotchas:
- 'jf go-config' must be run first; it sets GOPROXY for the go invocation.
- For private modules, GOPRIVATE should be set explicitly in your environment.

Related: jf go-config, jf go-publish, jf rt build-publish`
}
