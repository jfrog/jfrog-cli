package vgopublish

const Description = "Publish vgo project"

var Usage = []string{`jfrog rt vp [command options] <target repository> <project version>`}

const Arguments string = `	target repository
		Target repository in Artifactory.
	project version
		Project version to be published.`