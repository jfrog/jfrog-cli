package helmcommand

var Usage = []string{"helm <helm arguments> [command options]"}

func GetDescription() string {
	return "Run native Helm command with build-info collection support."
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
