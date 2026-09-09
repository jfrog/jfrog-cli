package nuget

var Usage = []string{"nuget <nuget args> [command options]"}

func GetDescription() string {
	return "Run NuGet."
}

func GetArguments() string {
	return `	nuget command
		The nuget command to run. For example, restore.`
}

func GetAIDescription() string {
	return `Run a NuGet command (restore, pack, push) through JFrog: dependencies resolve via an Artifactory NuGet repository, optionally collecting build-info.

When to use:
- Restoring NuGet packages from an Artifactory NuGet repo.
- Producing build-info for .NET projects that use nuget.exe directly.

Prerequisites:
- A local nuget binary on PATH.
- A configured server.
- Either JFROG_RUN_NATIVE=true (native/FlexPack mode, no per-project config needed), or
  'jf nuget-config' run once in the project directory (legacy mode).

Common patterns:
  $ jf nuget restore MyApp.sln
  $ jf nuget restore --build-name=my-app --build-number=2
  $ export JFROG_RUN_NATIVE=true
  $ jf nuget restore MyApp.sln --repo-resolve my-nuget-virtual --server-id my-server

Gotchas:
- 'jf nuget-config' is optional when JFROG_RUN_NATIVE=true. In that mode a per-project
  .jfrog/projects/nuget.yaml is ignored (a warning is printed) and the native path is used.
- Without JFROG_RUN_NATIVE=true, 'jf nuget-config' must be run first, and the native-only
  flags --repo-resolve / --server-id are not supported.
- For .NET Core/SDK projects, prefer 'jf dotnet' instead.
- The nuget binary on Linux/macOS often comes from Mono and behaves differently than on Windows.

Related: jf nuget-config, jf dotnet`
}
