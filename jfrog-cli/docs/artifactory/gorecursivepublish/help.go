package gorecursivepublish

const Description = "Populates and publish dependencies and transitive dependencies of the project"

var Usage = []string{`jfrog rt grp <target repository>`}

const Arguments string = `	target repository
		Target repository in Artifactory. Publish the dependencies to this repository. Also, this will Set GOPROXY environment variable to resolve dependencies from this repository.`
