package dotnetconfig

var Usage = []string{"dotnet-config [command options]"}

func GetDescription() string {
	return "Generate dotnet configuration."
}

func GetAIDescription() string {
	return `Write a per-project .NET configuration (.jfrog/projects/dotnet.yaml) so 'jf dotnet' resolves NuGet packages through Artifactory.

When to use:
- First-time setup of a .NET SDK project.

Prerequisites:
- A configured server.
- The Artifactory NuGet repository key.

Common patterns:
  $ jf dotnet-config --server-id-resolve=my-server --repo-resolve=nuget-virtual

Gotchas:
- Interactive prompts trigger when required flags are missing.
- For older nuget.exe projects, use 'jf nuget-config' instead.

Related: jf dotnet, jf nuget-config`
}
