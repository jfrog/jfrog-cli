package attachlic

const Description string = "Attaches (and optionally installs) a license to the specified service and removes it from the specified bucket."

var Usage = []string{"jfrog mc s attach-lic --bucket-id=<bucket ID or name> [command options] <service name>"}

const Arguments string = `	Service name
		The name of the service to which the license should be attached.`
