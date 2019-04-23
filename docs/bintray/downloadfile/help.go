package downloadfile

const Description string = "Download file."

var Usage = []string{"jfrog bt dlf [command options] <source path> [target path]"}

const Arguments string = `	source path
		The path, in Bintray, to the file that should be downloaded.
		Format: subject/repository/path.

	target path
		[Optional]
 		This second optional argument lets you specify a path in local file system to where the file should be downloaded.`
