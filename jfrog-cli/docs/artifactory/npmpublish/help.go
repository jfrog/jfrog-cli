package npmpublish

const Description = "Packs and deploys the npm package to the designated npm repository."

var Usage = []string{`jfrog rt npmp [command options] <repository name>`}

const Arguments string = `	repository name
		The destination npm repository. Can be a local repository or a virtual repository with a 'Default Deployment Repository'.`
