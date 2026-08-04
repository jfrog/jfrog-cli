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
- 'jf nuget-config' run once in the project directory.
- A configured server.

Common patterns:
  $ jf nuget restore MyApp.sln
  $ jf nuget restore --build-name=my-app --build-number=2

Gotchas:
- 'jf nuget-config' must be run first.
- For .NET Core/SDK projects, prefer 'jf dotnet' instead.
- The nuget binary on Linux/macOS often comes from Mono and behaves differently than on Windows.

Related: jf nuget-config, jf dotnet`
}
