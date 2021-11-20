package jpddelete

var Usage = []string{"mc jd [command options] <jpd id>"}

func GetDescription() string {
	return "Delete a JPD from Mission Control."
}

func GetArguments() string {
	return `	JPD ID
		The ID of the JPD to be removed from Mission Control.`
}
