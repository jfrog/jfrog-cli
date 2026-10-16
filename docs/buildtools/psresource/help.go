package psresource

import "strings"

// This package backs four separate top-level jf commands - Install-PSResource, Save-PSResource,
// Update-PSResource and Publish-PSResource - one per native PowerShell PSResourceGet cmdlet, rather
// than a single "jf psresource <verb>" wrapper (unlike jf choco, which does use a single-command,
// first-positional-argument dispatch). The four commands' help text is near-identical (same
// Prerequisites block, same shared Gotchas bullet, same overall shape), so rather than four
// hand-duplicated copies, it's rendered from one shared template (Usage/GetDescription/
// GetArguments/GetAIDescription below) parameterized by each cmdlet's full native name and, for
// GetAIDescription, the small set of values in cmdletMeta that actually differ between cmdlets.
// buildtools/cli.go calls these directly (one per cmdlet name) to register four distinct
// cli.Command entries that each read exactly like their own command's help.

// cmdletMeta holds the values that vary between the four PSResourceGet cmdlets' AI descriptions.
// Everything else in the rendered text (Prerequisites block, the shared build-info Gotchas bullet,
// bullet ordering) is identical across all four and lives directly in GetAIDescription.
type cmdletMeta struct {
	paramList           string   // native parameters shown in the description's first line, e.g. "-Name, -Version, -Repository, -Scope, ..."
	resultClause        string   // completes "...and optionally records <resultClause>."
	examples            []string // fully-formatted "  $ jf ..." example lines, in display order
	middleGotchas       []string // gotcha bullets specific to this cmdlet, between the shared build-info bullet and the shared registration bullet
	registrationGotcha  string   // registration gotcha bullet text (Install's differs from the other three)
	crossPlatformGotcha string   // cross-platform gotcha bullet text (Install/Publish differ from Save/Update)
}

// cmdletMetaByName holds the AI-description metadata for each of the four PSResourceGet cmdlets,
// keyed by their full native name (matching psresourcecommand.SubCommand* in
// jfrog-cli-artifactory, and buildtools/cli.go's cli.Command.Name for each).
var cmdletMetaByName = map[string]cmdletMeta{
	"Install-PSResource": {
		paramList:    "-Name, -Version, -Repository, -Scope, ...",
		resultClause: "build-info for the resolved package(s)",
		examples: []string{
			"  $ jf Install-PSResource -Name MyModule -Repository jfrt-acme.jfrog.io-psresource-virtual --build-name=app --build-number=1",
			"  $ jf Install-PSResource -Name MyModule -Version 2.0.0 -Repository jfrt-acme.jfrog.io-psresource-virtual --repo-resolve=psresource-virtual --build-name=app --build-number=1",
		},
		middleGotchas: []string{
			"Dependency build-info is direct-dependency-only: PSResourceGet has no lock file, so there is no transitive dependency graph, the same limitation 'jf choco install' has.",
			"Checksums for the recorded dependency are fetched with a HEAD request against Artifactory, not read from a local file.",
			"'--repo-resolve' records the resolution repository in build-info; it does not itself select which repository PSResourceGet resolves from - that is still driven by the native -Repository parameter (or by 'jf setup psresource').",
		},
		registrationGotcha:  "This command never writes PSResourceGet repository registration (PSResourceRepository.xml). Use 'jf setup psresource' for that.",
		crossPlatformGotcha: "Cross-platform: unlike 'jf choco' (Windows only), this runs on macOS, Linux and Windows wherever pwsh and PSResourceGet are installed.",
	},
	"Save-PSResource": {
		paramList:    "-Name, -Version, -Path, -Repository, ...",
		resultClause: "build-info for the saved package(s)",
		examples: []string{
			"  $ jf Save-PSResource -Name MyModule -Path ./local-modules -Repository jfrt-acme.jfrog.io-psresource-virtual --build-name=app --build-number=1",
		},
		middleGotchas: []string{
			"Dependency build-info is direct-dependency-only, matching Install-PSResource - no transitive graph.",
			"When '-Path' is given, the saved .nupkg is written locally, so its checksums are computed from that file; otherwise checksums fall back to a HEAD request against Artifactory.",
		},
		registrationGotcha:  "This command never writes PSResourceGet repository registration. Use 'jf setup psresource' for that.",
		crossPlatformGotcha: "Cross-platform: macOS, Linux and Windows, wherever pwsh and PSResourceGet are installed.",
	},
	"Update-PSResource": {
		paramList:    "-Name, -Version, -Repository, ...",
		resultClause: "build-info for the resolved package(s) after the update",
		examples: []string{
			"  $ jf Update-PSResource -Name MyModule -Repository jfrt-acme.jfrog.io-psresource-virtual --build-name=app --build-number=1",
		},
		middleGotchas: []string{
			"Dependency build-info is direct-dependency-only, matching Install-PSResource - no transitive graph.",
			"The recorded version is the one resolved after the update completes, not any version implied on the command line.",
			"Checksums are fetched with a HEAD request against Artifactory.",
		},
		registrationGotcha:  "This command never writes PSResourceGet repository registration. Use 'jf setup psresource' for that.",
		crossPlatformGotcha: "Cross-platform: macOS, Linux and Windows, wherever pwsh and PSResourceGet are installed.",
	},
	"Publish-PSResource": {
		paramList:    "-Path, -Repository, ...",
		resultClause: "artifact build-info and stamps build properties on the published package",
		examples: []string{
			"  $ jf Publish-PSResource -Path ./MyModule -Repository jfrt-acme.jfrog.io-psresource-local --repo=psresource-local --build-name=app --build-number=1",
		},
		middleGotchas: []string{
			"Publish-PSResource compiles and uploads the .nupkg internally - it is never written to local disk - so the artifact is confirmed and its checksums are fetched from Artifactory via a HEAD request rather than read from a local file.",
			"'--repo' identifies the deployment repository for build-info purposes; it must agree with whatever -Repository resolves to in Artifactory.",
		},
		registrationGotcha:  "This command never writes PSResourceGet repository registration. Use 'jf setup psresource' for that.",
		crossPlatformGotcha: "Cross-platform: unlike 'jf choco' (Windows only), this runs on macOS, Linux and Windows wherever pwsh and PSResourceGet are installed.",
	},
}

// Usage returns the usage line for the top-level jf command wrapping the given PSResourceGet
// cmdlet (e.g. cmdletName "Install-PSResource").
func Usage(cmdletName string) []string {
	return []string{cmdletName + " <cmdlet args> [command options]"}
}

// GetDescription returns the short (non-AI) description for the given PSResourceGet cmdlet.
func GetDescription(cmdletName string) string {
	return "Run the native PowerShell " + cmdletName + " cmdlet with optional JFrog build-info collection."
}

// GetArguments returns the UsageText argument description for the given PSResourceGet cmdlet.
func GetArguments(cmdletName string) string {
	return "\t" + cmdletName + " cmdlet args\n\t\t\tArguments and options for the native " + cmdletName + " cmdlet."
}

// GetAIDescription renders the full AI-oriented description for the given PSResourceGet cmdlet
// (one of "Install-PSResource", "Save-PSResource", "Update-PSResource", "Publish-PSResource"),
// interpolating its cmdletMeta into the prose shared by all four cmdlets.
func GetAIDescription(cmdletName string) string {
	m := cmdletMetaByName[cmdletName]
	var b strings.Builder
	b.WriteString("Run the native PowerShell " + cmdletName + " cmdlet through JFrog. The command forwards " + cmdletName +
		"'s own parameters unchanged (" + m.paramList + ") and optionally records " + m.resultClause + ".\n\n")
	b.WriteString("Prerequisites:\n")
	b.WriteString("- PowerShell 7+ (`pwsh`) with the Microsoft.PowerShell.PSResourceGet module installed.\n")
	b.WriteString("- Run 'jf setup psresource' first to register an authenticated PSResourceGet repository against Artifactory, or pass the native -Repository/-Credential parameters yourself.\n\n")
	b.WriteString("Examples:\n")
	b.WriteString(strings.Join(m.examples, "\n"))
	b.WriteString("\n\n")
	b.WriteString("Gotchas:\n")
	b.WriteString("- Build-info is collected only when both '--build-name' and '--build-number' are given; supplying just one is an error.\n")
	for _, g := range m.middleGotchas {
		b.WriteString("- " + g + "\n")
	}
	b.WriteString("- " + m.registrationGotcha + "\n")
	b.WriteString("- " + m.crossPlatformGotcha)
	return b.String()
}
