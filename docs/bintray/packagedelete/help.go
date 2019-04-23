package packagedelete

const Description string = "Delete Package."

var Usage = []string{"jfrog bt pd [command options] <target path>"}

const Arguments string = `	target path
		The path, in Bintray, to the package that should be deleted.
		Format: subject/repository/package.`
