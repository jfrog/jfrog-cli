package npminstall

const Description = "Run npm install."

var Usage = []string{`jfrog rt npm-install <repo> [command options]`}

const Arguments string =
`	repo
		The source npm repository. Can be a local, remote or virtual npm repository.`