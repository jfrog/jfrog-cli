package gpgsignfile

const Description string = "GPG Sign file."

var Usage = []string{"jfrog bt gsf [command options] <target path>"}

const Arguments string =
`	target path
		The path to the file which is being signed, formatted subject/repository/path.`