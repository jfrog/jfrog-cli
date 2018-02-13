package buildadddependencies

const Description = "Adds dependencies from the local file-system to the build info."

var Usage = []string{"jfrog rt bad [command options] <build name> <build number> <pattern>",
	"jfrog rt bad --spec=<File Spec path> [command options] <build name> <build number>"}

const Arguments string = `	build name
		Build name.

	build number
		Build number.

	pattern
		Specifies the local file system path to dependencies which should be added to the build info.
		You can specify multiple dependencies by using wildcards or a regular expression as designated by the --regexp command option.
		If you have specified that you are using regular expressions, then the first one used in the argument must be enclosed in parenthesis.`
