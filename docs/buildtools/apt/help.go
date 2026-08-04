package apt

var Usage = []string{"apt <native-command> <args> [command options]"}

func GetDescription() string {
	return "Run apt-get commands against a JFrog Artifactory Debian repository."
}

func GetAIDescription() string {
	return `Run apt package-manager commands (install, apt-cache, dpkg-query, etc.) against a JFrog Artifactory Debian repository. Wraps the native apt-get/apt-cache/dpkg-query binaries and injects Artifactory authentication via a temporary sources.list that is removed after the command completes.

When to use:
- Installing Debian/Ubuntu packages from an Artifactory Debian repository with on-the-fly authentication.
- Running one-off apt commands against Artifactory without persistently editing system apt config.

Prerequisites:
- A Debian/Ubuntu host with apt-get available.
- A configured server.
- Root (or sudo) to modify apt state, unless the operation is read-only (e.g. apt-cache).

Common patterns:
  $ jf apt install curl --repo=ci-debian-local --dist=bookworm
  $ jf apt install curl vim --repo=ci-debian-local --dist=bookworm --component=main
  $ jf apt install curl --skip-login

Gotchas:
- --repo and --dist are required for on-the-fly auth; without them apt falls back to the system config.
- --skip-login bypasses auth injection and uses the existing sources.list.
- For persistent authentication, use 'jf setup apt' to write a managed sources.list entry instead.

Related: jf setup apt`
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
