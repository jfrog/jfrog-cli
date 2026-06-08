package nugetconfig

var Usage = []string{"nuget-config [command options]"}

func GetDescription() string {
	return "Generate nuget configuration."
}

func GetAIDescription() string {
	return `Write a per-project NuGet configuration (.jfrog/projects/nuget.yaml) so 'jf nuget' resolves through Artifactory.

When to use:
- First-time setup of a NuGet project.

Prerequisites:
- A configured server.
- The Artifactory NuGet repository key.

Common patterns:
  $ jf nuget-config --server-id-resolve=my-server --repo-resolve=nuget-virtual

Gotchas:
- Interactive prompts trigger when required flags are missing.
- Does NOT affect .NET Core/SDK projects; for those use 'jf dotnet-config'.

Related: jf nuget, jf dotnet-config`
}
