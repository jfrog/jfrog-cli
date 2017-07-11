package attachlic

const Description string = "Attaches (and optionally installs) a license to the specified Artifactory instance and removes it from the specified bucket."

var Usage = []string{"jfrog mc rti attach-lic --bucket-id=<bucket ID or name> [command options] <Artifactory instance name>"}

const Arguments string =
`	Artifactory instance name
		The name of the Artifactory instance to which the license should be attached.`