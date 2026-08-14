package helmcommand

var Usage = []string{"helm <helm arguments> [command options]"}

func GetDescription() string {
	return "Run native Helm command with build-info collection support."
}

func GetAIDescription() string {
	return `Run a Helm command (package, push, pull, dependency update) through JFrog with build-info collection. Wraps the local helm binary; arguments pass through.

When to use:
- Packaging and pushing Helm charts to an Artifactory OCI Helm repository.
- Resolving chart dependencies through Artifactory with build-info tracking.

Prerequisites:
- A local helm binary (Helm 3+).
- A configured server.
- For build-info: --build-name and --build-number passed together.

Common patterns:
  $ jf helm package ./mychart --build-name=my-build --build-number=1
  $ jf helm push mychart-0.1.0.tgz oci://myrepo.jfrog.io/helm-local --build-name=my-build --build-number=1
  $ jf helm dependency update ./mychart --build-name=my-build --build-number=1

Gotchas:
- Only 'package', 'push', and 'dependency' contribute to build-info; other commands are pure passthroughs.
- OCI registries require the URL prefix oci://; HTTP-based Helm repos use the helm-repo URL form.
- --repository-cache lets you override the chart cache location, useful in CI.

Related: jf rt build-publish, jf docker push`
}

func GetArguments() string {
	return `	helm arguments
		Helm command and arguments to execute. All standard Helm commands are supported.

	Command Options:
		--build-name              Build name.
		--build-number            Build number.
		--module                  Optional module name for the build-info.
		--project                 Project key for associating the build with a project.
		--server-id               Artifactory server ID configured using the config command.
		--repository-cache        Path to the Helm repository cache directory.
		--username                Artifactory username.
		--password                Artifactory password.

	Examples:

	  $ jf helm push mychart-0.1.0.tgz oci://myrepo.jfrog.io/helm-local --build-name=my-build --build-number=1

	  $ jf helm package ./mychart --build-name=my-build --build-number=1

	  $ jf helm dependency update ./mychart --build-name=my-build --build-number=1

	  $ jf helm package . --build-name=my-build --build-number=1 --server-id=my-server

	Supported Commands:

	  package                  Package a chart directory into a chart archive.
	                           Collects build-info including chart dependencies.

	  push                     Push a chart to a remote registry (OCI).
	                           Collects build-info for the pushed chart.

	  dependency               Manage a chart's dependencies.
	                           The 'dependency update' command collects build-info for downloaded dependencies.

	  pull                     Download a chart from a remote registry.

	  install                  Install a chart.

	  upgrade                  Upgrade a release.

	  repo                     Add, list, remove, update, and index chart repositories.

	  help, h                  Show help for any command.

	Build Info Collection:
		The helm command automatically collects build-info when used with --build-name and --build-number.
		Build-info is collected for:
		- Chart artifacts (when using 'push' commands)
		- Chart dependencies (when using 'package' or 'dependency update' commands)
		
		Use 'jf rt build-publish' to publish the collected build-info to Artifactory.`
}
