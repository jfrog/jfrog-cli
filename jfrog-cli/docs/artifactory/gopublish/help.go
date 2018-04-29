package gopublish

const Description = "Publish go package and/or it's dependencies to Artifactory"

var Usage = []string{`jfrog rt gp [command options] <target repository> <project version>`}

const Arguments string = `	target repository
		Target repository in Artifactory.
	project version
		Package version to be published.`
