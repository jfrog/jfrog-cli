package search

const Description = "Search files."

var Usage = []string{"jfrog rt s [command options] <search pattern>",
	"jfrog rt s --spec=<File Spec path> [command options]"}

const Arguments string = `	search pattern
		Specifies the search path in Artifactory, in the following format: <repository name>/<repository path>.
		You can use wildcards to specify multiple artifacts.`
