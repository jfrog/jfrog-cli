package mavensync

const Description string = "Sync Version Artifacts to Maven Central"

var Usage = []string{"jfrog bt mcs [command options] <target path>"}

const Arguments string = `	target path
		The path, in Bintray, to the version that should be synced.
		Format: subject/repository/package/version.`
