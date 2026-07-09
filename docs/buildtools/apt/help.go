package apt

var Usage = []string{"apt <native-command> <args> [command options]"}

func GetDescription() string {
	return "Run apt-get commands against a JFrog Artifactory Debian repository."
}

func GetArguments() string {
	return `	apt native-command
		Wraps apt-get/apt-cache/dpkg-query commands with JFrog Artifactory
		authentication. Credentials are injected via a temporary sources.list
		file that is removed after the command completes.

		Examples:
		- jf apt install curl --repo=ci-debian-local --dist=bookworm
		- jf apt install curl vim --repo=ci-debian-local --dist=bookworm --component=main
		- jf apt install curl --skip-login   (uses existing sources.list auth)

		Setup (persistent authentication):
		  jf setup apt --repo ci-debian-local --dist bookworm --component main`
}
