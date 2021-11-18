package ocstartbuild

var Usage = []string{"rt oc start-build <build config name | --from-build=<build name>> --repo=<target repository> [command options]"}

func GetDescription() string {
	return "Run OpenShift CLI (oc) start-build command."
}
