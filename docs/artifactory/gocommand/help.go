package gocommand

const Description = "Runs go"

var Usage = []string{`jfrog rt go [command options] <go arguments> <target repository>`}

const Arguments string = `	go commands
		Arguments and options for the go command.
	target repository
		Target repository in Artifactory. This will Set GOPROXY environment variable to resolve dependencies from this repository.`
