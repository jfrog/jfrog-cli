package remove

const Description string = "Removes a service from Mission Control."

var Usage = []string{"jfrog mc s remove [command options] <service name>"}

const Arguments string = `	Service name
		The name of the service (as defined in Mission Control) to remove.`
