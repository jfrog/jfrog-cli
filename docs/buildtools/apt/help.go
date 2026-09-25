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
- Capturing build-info for installed packages (pass --build-name and --build-number).

Prerequisites:
- A Debian/Ubuntu host with apt-get available.
- A configured server.
- Root (or sudo) to modify apt state, unless the operation is read-only (e.g. apt-cache).

Common patterns:
  $ jf apt install curl --repo=ci-debian-local --dist=bookworm
  $ jf apt install curl vim --repo=ci-debian-local --dist=bookworm --component=main
  $ jf apt install curl --skip-login
  $ jf apt install curl --repo=ci-debian-local --dist=bookworm --build-name=my-build --build-number=1

Gotchas:
- --repo and --dist are required for on-the-fly auth; without them apt falls back to the system config.
- --skip-login bypasses auth injection and uses the existing sources.list.
- --build-name and --build-number must both be set to collect build-info; use 'jf rt bp' to publish.
- For persistent authentication, use 'jf setup apt' to write a managed sources.list entry instead.

Related: jf setup apt, jf rt bp`
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
		- jf apt install curl --repo=ci-debian-local --dist=bookworm \
		    --build-name=my-build --build-number=1   (collect build-info)
		- jf apt install curl --repo=ci-debian-local --dist=bookworm \
		    --build-name=my-build --build-number=1 --module=debian-pkgs --project=myproject

		Setup (persistent authentication):
		  jf setup apt --repo ci-debian-local --dist bookworm --component main

		Publish build-info after install:
		  jf rt bp my-build 1`
}
