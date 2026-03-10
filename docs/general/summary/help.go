package summary

var Usage = []string{"gsm"}

func GetDescription() string {
	return `Generate a summary of recorded CLI commands that were executed on the current machine. The report is generated in Markdown format and saved in the directory stored in the JFROG_CLI_COMMAND_SUMMARY_OUTPUT_DIR environment variable.`
}
