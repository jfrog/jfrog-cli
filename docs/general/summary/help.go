package summary

var Usage = []string{"gsm"}

func GetDescription() string {
	return `Generate a summary of recorded CLI commands that were executed on the current machine. The report is generated in Markdown format and saved in the directory stored in the JFROG_CLI_COMMAND_SUMMARY_OUTPUT_DIR environment variable.`
}

func GetAIDescription() string {
	return `Finalize and render a Markdown summary of CLI commands executed during the current run. Designed for CI integration: the report can be picked up by GitHub Actions $GITHUB_STEP_SUMMARY or similar systems. Reads recorded command artifacts from the directory specified in JFROG_CLI_COMMAND_SUMMARY_OUTPUT_DIR.

When to use:
- At the end of a CI job to publish a human-readable summary of uploads, builds, scans, and other tracked operations.

Prerequisites:
- The JFROG_CLI_COMMAND_SUMMARY_OUTPUT_DIR environment variable must be set during the prior jf commands so they record summary data.

Common patterns:
  $ JFROG_CLI_COMMAND_SUMMARY_OUTPUT_DIR=$RUNNER_TEMP/jfrog-summary jf gsm

Gotchas:
- If JFROG_CLI_COMMAND_SUMMARY_OUTPUT_DIR was not set during the recorded commands, there is nothing to summarize and the output is empty.
- Output is always Markdown; no JSON or table format option.

Related: jf rt upload, jf rt build-publish`
}
