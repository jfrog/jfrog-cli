package summary

var Usage = []string{"csm"}

func GetDescription() string {
	return `Generates a Summary of recorded CLI commands there were executed on the current machine.
	The report is generated in Markdown format and saved in the root directory of JFROG_CLI_COMMAND_SUMMARY_OUTPUT_DIR.
`
}
