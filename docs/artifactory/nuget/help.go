package nuget

const Description = "Run NuGet."

var Usage = []string{`jfrog rt nuget [command options] <nuget args> <source repository name>`}

const Arguments string = `	nuget command
		The nuget command to run. For example, restore.

	source repository name
		The source NuGet repository. Can be a local, remote or virtual NuGet repository.`
