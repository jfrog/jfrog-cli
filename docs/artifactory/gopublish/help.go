package gopublish

const Description = "Publish go package and/or its dependencies to Artifactory"

var Usage = []string{`jfrog rt gp [command options] <project version>`}

const Arguments string = `	project version
		Package version to be published.`
