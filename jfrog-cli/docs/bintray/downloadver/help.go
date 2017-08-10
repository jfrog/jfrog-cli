package downloadver

const Description string = "Download Version files."

var Usage = []string{"jfrog rt dlv [command options] <source path> [target path]"}

const Arguments string =
`	source path
		The path, in Bintray, to the version whose files should be downloaded.
		Format: subject/repository/package/version/.

	target path
		[Optional]
		This second optional argument lets you specify a path in local file system to where the version files should be downloaded.`