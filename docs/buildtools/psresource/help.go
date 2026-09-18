package psresource

// This package backs four separate top-level jf commands - Install-PSResource, Save-PSResource,
// Update-PSResource and Publish-PSResource - one per native PowerShell PSResourceGet cmdlet, rather
// than a single "jf psresource <verb>" wrapper (unlike jf choco, which does use a single-command,
// first-positional-argument dispatch). Each cmdlet gets its own Usage/GetXDescription/
// GetXArguments/GetXAIDescription set below so buildtools/cli.go can register four distinct
// cli.Command entries that each read like their own command's help.

var InstallUsage = []string{"Install-PSResource <cmdlet args> [command options]"}

func GetInstallDescription() string {
	return "Run the native PowerShell Install-PSResource cmdlet with optional JFrog build-info collection."
}

func GetInstallArguments() string {
	return `	Install-PSResource cmdlet args
			Arguments and options for the native Install-PSResource cmdlet.`
}

func GetInstallAIDescription() string {
	return `Run the native PowerShell Install-PSResource cmdlet through JFrog. The command forwards Install-PSResource's own parameters unchanged (-Name, -Version, -Repository, -Scope, ...) and optionally records build-info for the resolved package(s).

Prerequisites:
- PowerShell 7+ (` + "`pwsh`" + `) with the Microsoft.PowerShell.PSResourceGet module installed.
- Run 'jf setup psresource' first to register an authenticated PSResourceGet repository against Artifactory, or pass the native -Repository/-Credential parameters yourself.

Examples:
  $ jf Install-PSResource -Name MyModule -Repository jfrt-acme.jfrog.io-psresource-virtual --build-name=app --build-number=1
  $ jf Install-PSResource -Name MyModule -Version 2.0.0 -Repository jfrt-acme.jfrog.io-psresource-virtual --repo-resolve=psresource-virtual --build-name=app --build-number=1

Gotchas:
- Build-info is collected only when both '--build-name' and '--build-number' are given; supplying just one is an error.
- Dependency build-info is direct-dependency-only: PSResourceGet has no lock file, so there is no transitive dependency graph, the same limitation 'jf choco install' has.
- Checksums for the recorded dependency are fetched with a HEAD request against Artifactory, not read from a local file.
- '--repo-resolve' records the resolution repository in build-info; it does not itself select which repository PSResourceGet resolves from - that is still driven by the native -Repository parameter (or by 'jf setup psresource').
- This command never writes PSResourceGet repository registration (PSResourceRepository.xml). Use 'jf setup psresource' for that.
- Cross-platform: unlike 'jf choco' (Windows only), this runs on macOS, Linux and Windows wherever pwsh and PSResourceGet are installed.`
}

var SaveUsage = []string{"Save-PSResource <cmdlet args> [command options]"}

func GetSaveDescription() string {
	return "Run the native PowerShell Save-PSResource cmdlet with optional JFrog build-info collection."
}

func GetSaveArguments() string {
	return `	Save-PSResource cmdlet args
			Arguments and options for the native Save-PSResource cmdlet.`
}

func GetSaveAIDescription() string {
	return `Run the native PowerShell Save-PSResource cmdlet through JFrog. The command forwards Save-PSResource's own parameters unchanged (-Name, -Version, -Path, -Repository, ...) and optionally records build-info for the saved package(s).

Prerequisites:
- PowerShell 7+ (` + "`pwsh`" + `) with the Microsoft.PowerShell.PSResourceGet module installed.
- Run 'jf setup psresource' first to register an authenticated PSResourceGet repository against Artifactory, or pass the native -Repository/-Credential parameters yourself.

Examples:
  $ jf Save-PSResource -Name MyModule -Path ./local-modules -Repository jfrt-acme.jfrog.io-psresource-virtual --build-name=app --build-number=1

Gotchas:
- Build-info is collected only when both '--build-name' and '--build-number' are given; supplying just one is an error.
- Dependency build-info is direct-dependency-only, matching Install-PSResource - no transitive graph.
- When '-Path' is given, the saved .nupkg is written locally, so its checksums are computed from that file; otherwise checksums fall back to a HEAD request against Artifactory.
- This command never writes PSResourceGet repository registration. Use 'jf setup psresource' for that.
- Cross-platform: macOS, Linux and Windows, wherever pwsh and PSResourceGet are installed.`
}

var UpdateUsage = []string{"Update-PSResource <cmdlet args> [command options]"}

func GetUpdateDescription() string {
	return "Run the native PowerShell Update-PSResource cmdlet with optional JFrog build-info collection."
}

func GetUpdateArguments() string {
	return `	Update-PSResource cmdlet args
			Arguments and options for the native Update-PSResource cmdlet.`
}

func GetUpdateAIDescription() string {
	return `Run the native PowerShell Update-PSResource cmdlet through JFrog. The command forwards Update-PSResource's own parameters unchanged (-Name, -Version, -Repository, ...) and optionally records build-info for the resolved package(s) after the update.

Prerequisites:
- PowerShell 7+ (` + "`pwsh`" + `) with the Microsoft.PowerShell.PSResourceGet module installed.
- Run 'jf setup psresource' first to register an authenticated PSResourceGet repository against Artifactory, or pass the native -Repository/-Credential parameters yourself.

Examples:
  $ jf Update-PSResource -Name MyModule -Repository jfrt-acme.jfrog.io-psresource-virtual --build-name=app --build-number=1

Gotchas:
- Build-info is collected only when both '--build-name' and '--build-number' are given; supplying just one is an error.
- Dependency build-info is direct-dependency-only, matching Install-PSResource - no transitive graph.
- The recorded version is the one resolved after the update completes, not any version implied on the command line.
- Checksums are fetched with a HEAD request against Artifactory.
- This command never writes PSResourceGet repository registration. Use 'jf setup psresource' for that.
- Cross-platform: macOS, Linux and Windows, wherever pwsh and PSResourceGet are installed.`
}

var PublishUsage = []string{"Publish-PSResource <cmdlet args> [command options]"}

func GetPublishDescription() string {
	return "Run the native PowerShell Publish-PSResource cmdlet with optional JFrog build-info collection."
}

func GetPublishArguments() string {
	return `	Publish-PSResource cmdlet args
			Arguments and options for the native Publish-PSResource cmdlet.`
}

func GetPublishAIDescription() string {
	return `Run the native PowerShell Publish-PSResource cmdlet through JFrog. The command forwards Publish-PSResource's own parameters unchanged (-Path, -Repository, ...) and optionally records artifact build-info and stamps build properties on the published package.

Prerequisites:
- PowerShell 7+ (` + "`pwsh`" + `) with the Microsoft.PowerShell.PSResourceGet module installed.
- Run 'jf setup psresource' first to register an authenticated PSResourceGet repository against Artifactory, or pass the native -Repository/-Credential parameters yourself.

Examples:
  $ jf Publish-PSResource -Path ./MyModule -Repository jfrt-acme.jfrog.io-psresource-local --repo=psresource-local --build-name=app --build-number=1

Gotchas:
- Build-info is collected only when both '--build-name' and '--build-number' are given; supplying just one is an error.
- Publish-PSResource compiles and uploads the .nupkg internally - it is never written to local disk - so the artifact is confirmed and its checksums are fetched from Artifactory via a HEAD request rather than read from a local file.
- '--repo' identifies the deployment repository for build-info purposes; it must agree with whatever -Repository resolves to in Artifactory.
- This command never writes PSResourceGet repository registration. Use 'jf setup psresource' for that.
- Cross-platform: unlike 'jf choco' (Windows only), this runs on macOS, Linux and Windows wherever pwsh and PSResourceGet are installed.`
}
