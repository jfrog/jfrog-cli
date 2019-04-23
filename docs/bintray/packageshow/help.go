package packageshow

const Description string = "Show Package details."

var Usage = []string{"jfrog bt ps [command options] <source path>"}

const Arguments string = `	source path
		The path, in Bintray, to the package whose details should be retrieved.
		Format: subject/repository/package.`
