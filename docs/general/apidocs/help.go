package apidocs

var Usage = []string{"api docs search <query> [command options]", "api docs describe <method> <path> [command options]"}

func GetDescription() string {
	return "Discover JFrog Platform REST API operations. Run 'jf api docs search <query>' to find a candidate endpoint, then 'jf api docs describe <method> <path>' to see its full shape, before using 'jf api <path>'."
}

func GetAIDescription() string {
	return `Namespace for API-discovery subcommands. Run 'jf api docs search <query>' to look up a REST endpoint by keyword, then 'jf api docs describe <method> <path>' to see its parameters/request body/response codes, before guessing at 'jf api <path>'.

See 'jf api docs search --help' and 'jf api docs describe --help' for the full set of options.`
}
