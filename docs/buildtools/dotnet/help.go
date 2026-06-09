package dotnet

var Usage = []string{"dotnet <dotnet sub-command> [command options]"}

func GetDescription() string {
	return "Run .NET Core CLI."
}

func GetArguments() string {
	return `	dotnet sub-command
		 Arguments and options for the dotnet command.`
}

func GetAIDescription() string {
	return `Run a .NET CLI command (restore, build, pack, push) through JFrog: package restoration is routed via Artifactory and optional build-info is collected.

When to use:
- Building .NET Core/SDK projects that consume NuGet packages from Artifactory.
- Capturing build-info for .NET pipelines.

Prerequisites:
- The .NET SDK installed (dotnet on PATH).
- 'jf dotnet-config' run once in the project directory.
- A configured server.

Common patterns:
  $ jf dotnet restore MyApp.sln
  $ jf dotnet build --build-name=my-app --build-number=4
  $ jf dotnet pack --configuration Release

Gotchas:
- 'jf dotnet-config' must be run first.
- Mixing 'jf nuget' and 'jf dotnet' configs in the same directory can create confused resolution.

Related: jf dotnet-config, jf nuget`
}
