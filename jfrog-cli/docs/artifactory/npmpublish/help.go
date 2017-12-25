package npmpublish

const Description = "Run npm publish."

var Usage = []string{`jfrog rt npm-publish <repo> [command options]`}

const Arguments string =
`	repo
		The destination npm repository. Can be a local repository or a virtual repository with a 'Default Deployment Repository'.`