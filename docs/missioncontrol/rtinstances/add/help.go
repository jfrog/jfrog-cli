package add

const Description string = "Adds an Artifactory instance."

var Usage = []string{"jfrog mc rti add --rt-url=<Artifactory url> --rt-user=<Artifactory username> --rt-password=<Artifactory password> [command options] <Artifactory instance name>"}

const Arguments string =
`	Artifactory instance name
		The name that the Artifactory instance should be given in Mission Control.`