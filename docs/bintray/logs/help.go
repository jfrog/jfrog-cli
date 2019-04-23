package logs

const Description string = "Download available log files for a package."

var Usage = []string{"jfrog bt l [command options] <target path>",
	"jfrog bt l download [command options] <target path> <log name>"}

const Arguments string = `	target path
		The path to the package for which you want the logs, formatted subject/repository/package.

	log name
		The name of the log file to download.`
