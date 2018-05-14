package detachlic

const Description string = "Detaches a license from the specified service and returns it to the specified bucket."

var Usage = []string{"jfrog mc s detach-lic --bucket-id=<bucket ID or name> [command options] <service name>"}

const Arguments string = `	Service name
		The name of the service from which the license should be detached.`
