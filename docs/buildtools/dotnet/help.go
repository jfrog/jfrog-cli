package dotnet

var Usage = []string{"dotnet <dotnet sub-command> [command options]"}

func GetDescription() string {
	return "Run .NET Core CLI."
}

func GetArguments() string {
	return `	dotnet sub-command
		 The dotnet sub-command to run, with its arguments and options.
		 Supported sub-commands: restore, build, publish, pack, add, and
		 'nuget push' (see 'Common patterns' below).`
}

func GetAIDescription() string {
	return `Run a .NET CLI command (restore, build, publish, pack, add, nuget push) through JFrog: package restoration is routed via Artifactory and optional build-info is collected.

When to use:
- Building .NET Core/SDK projects that consume NuGet packages from Artifactory.
- Publishing a .nupkg to Artifactory with 'jf dotnet nuget push'.
- Capturing build-info for .NET pipelines.

Prerequisites:
- The .NET SDK installed (dotnet on PATH).
- A configured server.
- Either JFROG_RUN_NATIVE=true (native/FlexPack mode, no per-project config needed), or
  'jf dotnet-config' run once in the project directory (legacy mode).

Common patterns:
  $ export JFROG_RUN_NATIVE=true
  $ jf dotnet restore MyApp.sln --repo-resolve my-nuget-virtual --server-id my-server
  $ jf dotnet build --build-name=my-app --build-number=4
  $ jf dotnet pack --configuration Release
  $ jf dotnet nuget push MyApp.1.0.0.nupkg --repo my-nuget-local --server-id my-server

Gotchas:
- 'jf dotnet-config' is optional when JFROG_RUN_NATIVE=true. In that mode a per-project
  .jfrog/projects/dotnet.yaml is ignored (a warning is printed) and the native path is used.
- Without JFROG_RUN_NATIVE=true, 'jf dotnet-config' must be run first, and the native-only
  flags --repo-resolve / --server-id are not supported.
- 'jf dotnet nuget push' is a two-token sub-command; plain 'jf dotnet push' is not a command.
- --repo-resolve routes the restore that a sub-command performs. restore, build, publish and
  pack all restore (publish and pack implicitly, unless --no-restore is passed), so all four
  honour it. 'dotnet add package' also restores, but the .NET SDK gives it no config-file
  option, so --repo-resolve cannot be applied there; run 'jf dotnet restore' first.
- Build-info dependencies are collected for restore-family sub-commands, and artifacts for
  pack and 'nuget push'. A pack that produces no package - for example every project already
  up to date - records an empty module rather than failing.
- Mixing 'jf nuget' and 'jf dotnet' configs in the same directory can create confused resolution.

Related: jf dotnet-config, jf nuget`
}
