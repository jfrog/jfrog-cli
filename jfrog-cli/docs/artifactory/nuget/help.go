package nuget

const Description = "Run NuGet."

var Usage = []string{`jfrog rt nuget [command options] <nuget args> <source repository name>`}

const Arguments string = `	nuget args
		Arguments to run with NuGet command.	

	source repository name
		The source NuGet repository. Can be a local, remote or virtual NuGet repository.`
