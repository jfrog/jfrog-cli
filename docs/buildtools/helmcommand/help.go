package helmcommand

var Usage = []string{"helm <helm arguments> [command options]"}

func GetDescription() string {
	return "Run native Helm command"
}

func GetArguments() string {
	return `	Examples:

	  $ jf helm push mychart-0.1.0.tgz oci://myrepo.jfrog.io/helm-local --build-name=my-build --build-number=1

	  $ jf helm package ./mychart --build-name=my-build --build-number=1

	  $ jf helm dependency update ./mychart --build-name=my-build --build-number=1

	Commands:

	  install                  Install a chart.

	  upgrade                  Upgrade a release.

	  package                  Package a chart directory into a chart archive.

	  push                     Push a chart to remote.

	  pull                     Download a chart from remote.

	  repo                     Add, list, remove, update, and index chart repositories.

	  dependency               Manage a chart's dependencies.

	  help, h                  Show help for any command.`
}
