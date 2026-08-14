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
- This does not configure the dotnet client itself. It is read only by 'jf dotnet' commands; a plain 'dotnet restore' keeps resolving from its own configuration, which this command never touches. To point the client itself at Artifactory for every project on the machine, run 'jf setup dotnet' instead - the two are independent and can even name different repositories.

Related: jf dotnet, jf nuget-config, jf setup dotnet`
}
