package remove

const Description string = "Removes an Artifactory instance."

var Usage = []string{"jfrog mc rti remove [command options] <Artifactory instance name>"}

const Arguments string = `	Artifactory instance name
		The name of the Artifactory instance (as defined in Mission Control) to remove.`
