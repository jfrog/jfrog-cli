package jpdadd

var Usage = []string{"mc ja [command options] <config>"}

func GetDescription() string {
	return "Add a JPD to Mission Control."
}

func GetArguments() string {
	return `	Config
		Path to a JSON configuration file containing the JPD details.`
}
