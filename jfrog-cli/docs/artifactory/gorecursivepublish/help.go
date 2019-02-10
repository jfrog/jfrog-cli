package gorecursivepublish

const Description = "Recursively discovers all project dependencies, both direct and indirect, and publishes them to Artifactory"

var Usage = []string{`jfrog rt grp <target repository>`}

const Arguments string = `	target repository
		Target repository in Artifactory. Publish the dependencies to this repository. Also, this will Set GOPROXY environment variable to resolve dependencies from this repository.`
