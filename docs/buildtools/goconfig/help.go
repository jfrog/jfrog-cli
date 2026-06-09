package goconfig

var Usage = []string{"go-config [command options]"}

func GetDescription() string {
	return "Generate go build configuration."
}

func GetAIDescription() string {
	return `Write a per-project Go configuration (.jfrog/projects/go.yaml) so 'jf go' uses an Artifactory Go repository for module resolution and 'jf go-publish' for publish.

When to use:
- Initial setup of a Go project to use a private module proxy.

Prerequisites:
- A configured server.
- The Artifactory Go repository key.

Common patterns:
  $ jf go-config --server-id-resolve=my-server --repo-resolve=go-virtual --repo-deploy=go-local

Gotchas:
- Interactive prompts trigger when required flags are missing.
- 'jf go' will set GOPROXY based on this config; other go invocations are unaffected.

Related: jf go, jf go-publish`
}
