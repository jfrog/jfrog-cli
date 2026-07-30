package apidocs

var Usage = []string{"api docs search <query> [command options]"}

func GetDescription() string {
	return "Discover JFrog Platform REST API operations. Run 'jf api docs search <query>' to find the right endpoint before using 'jf api <path>'."
}

func GetAIDescription() string {
	return `Namespace for API-discovery subcommands. Run 'jf api docs search <query>' to look up a REST endpoint by keyword before guessing at 'jf api <path>'.

See 'jf api docs search --help' for the full set of options.`
}
