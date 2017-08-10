package detachlic

const Description string = "Detaches a license from the specified Artifactory instance and returns it to the specified bucket."

var Usage = []string{"jfrog mc rti detach-lic --bucket-id=<bucket ID or name> [command options] <Artifactory instance name>"}

const Arguments string =
`	Artifactory instance name
		The name of the Artifactory instance from which the license should be detached.`