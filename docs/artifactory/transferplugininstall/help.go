package transferplugininstall

var Usage = []string{"rt transfer-plugin-install <server-id> [command options]"}

func GetDescription() string {
	return "[DEPRECATED] This command no longer installs the data-transfer user plugin and only prints a deprecation warning. 'jf rt transfer-files' does not require the plugin."
}

func GetArguments() string {
	return `	server-id
		Ignored.`
}

func GetAIDescription() string {
	return `[DEPRECATED] Informational command only. It no longer downloads or installs the data-transfer Artifactory user plugin; it only prints a deprecation warning and exits successfully. 'jf rt transfer-files' does not require the plugin on the source server. The command will be removed in a future release.

When to use:
- Do not use. Run 'jf rt transfer-files' directly.

Gotchas:
- All arguments and flags (--version, --dir, --home-dir) are accepted but ignored.

Related: jf rt transfer-files, jf rt transfer-settings

QA:
Q: Do I need to install the data-transfer user plugin before running 'jf rt transfer-files'?
A: No. The plugin is no longer required, and 'jf rt transfer-plugin-install' is deprecated and does nothing.
`
}
