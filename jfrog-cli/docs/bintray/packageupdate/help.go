package packageupdate

const Description string = "Update Package."

var Usage = []string{"jfrog bt pu [command options] <target path>"}

const Arguments string = `	target path
		The path, in Bintray, to the package that should be updated.
		Format: subject/repository/package.`
