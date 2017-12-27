package npminstall

const Description = "Run npm install."

var Usage = []string{`jfrog rt npmi [command options] <repository name>`}

const Arguments string =
`	repository name
		The source npm repository. Can be a local, remote or virtual npm repository.`