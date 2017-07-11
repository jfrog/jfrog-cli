package gpgsignver

const Description string = "GPG Sign Version."

var Usage = []string{"jfrog bt gsv [command options] <target path>"}

const Arguments string =
`	target path
		The path to the file which is being signed, formatted subject/repository/path/version.`