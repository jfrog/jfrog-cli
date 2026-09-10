package choco

var Usage = []string{"choco <choco args> [command options]"}

func GetDescription() string {
	return "Run Chocolatey with optional JFrog build-info collection."
}

func GetArguments() string {
	return `	choco sub-command
		Arguments and options for the native Chocolatey command.`
}

func GetAIDescription() string {
	return `Run a native Chocolatey command through JFrog. The command forwards Chocolatey arguments unchanged and optionally records build-info for pack, push, install, and upgrade.

Prerequisites:
- Chocolatey installed on Windows.
- Run 'jf setup choco' to add authenticated Artifactory sources, or pass the native Chocolatey source/authentication options yourself.

Examples:
  $ jf choco install git -s=jfrt-acme.jfrog.io-choco-virtual --build-name=image --build-number=1
  $ jf choco push tool.1.0.0.nupkg -s=jfrt-acme.jfrog.io-choco-local --repo=choco-local --build-name=tool --build-number=1

Gotchas:
- Chocolatey runs on Windows only.
- Build-info is collected only when both '--build-name' and '--build-number' are given; supplying just one is an error.
- The native '-s' source and JFrog '--repo' must identify the same Artifactory endpoint for push build-info.
- Install and upgrade record the requested packages and their transitive dependencies, read from Chocolatey's lib directory. A dependency served by a Chocolatey special source (ruby, cygwin, python, windowsfeatures) never lands there, so it is reported as not found and left out of the build-info.
- In CI, set '--execution-timeout' explicitly. Chocolatey's default of 2700 seconds is mishandled by Chocolatey itself and becomes a five-hour timeout, so a hung install stalls the build instead of failing it. '--execution-timeout=2700' is enough to avoid this.
- This command never writes Chocolatey configuration. Use 'jf setup choco' for that.`
}
