package config

const Description string = "Configure Mission Control details."

var Usage = []string{"jfrog mc c [command options]",
	"jfrog mc c show",
	"jfrog mc c [--interactive=<true|false>] clear"}

const Arguments string =
`	show
		Shows the stored configuration.

	clear
		Clears all stored configuration.`