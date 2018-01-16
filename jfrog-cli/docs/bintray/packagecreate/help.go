package packagecreate

const Description string = "Create Package."

var Usage = []string{"jfrog bt pc [command options] <target path>"}

const Arguments string = `	target path
		The path, in Bintray, to the package that should be created.
		Format: subject/repository/package.`
