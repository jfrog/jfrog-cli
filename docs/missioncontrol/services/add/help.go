package add

const Description string = "Adds a service in Mission Control."

var Usage = []string{"jfrog mc s add --service-url=<service url> --service-user=<service username> --service-password=<service password> [command options] <ARTIFACTORY | EDGE | XRAY | DISTRIBUTION> <service name>"}

const Arguments string = `	Service type
		Can be either ARTIFACTORY, EDGE, XRAY or DISTRIBUTION.

	Service name
		The name the new service should be get in Mission Control.`
