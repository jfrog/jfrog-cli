package transferplugininstall

var Usage = []string{"rt transfer-plugin-install <server-id> [command options]"}

func GetDescription() string {
	return "Download and install the data-transfer user plugin on the primary node of Artifactory, which is running on this local machine."
}

func GetArguments() string {
	return `	server-id
		The ID of the source server, on which the plugin should be installed.`
}

func GetAIDescription() string {
	return `Download the data-transfer Artifactory user plugin and install it on the primary node of the source Artifactory. Required as a one-time prep before 'jf rt transfer-files'. The plugin handles the actual byte streaming on the server side.

When to use:
- One-time setup on the source server before starting a self-hosted-to-cloud migration.

Prerequisites:
- Must be run on the same host as the source Artifactory's primary node (or with filesystem access to its etc/artifactory/plugins/ directory).
- Admin privileges on the source server.
- The source server ID is configured locally.

Common patterns:
  $ jf rt transfer-plugin-install source-prod
  $ jf rt transfer-plugin-install source-prod --dir=/opt/jfrog/artifactory/var/etc/artifactory/plugins

Gotchas:
- Must run as the user that owns the Artifactory installation; otherwise the file write fails.
- The plugin requires a Groovy plugins admin restart (or wait for the scheduled reload).
- This command only targets self-hosted instances; cloud-hosted sources need a different transfer mechanism.

Related: jf rt transfer-files, jf rt transfer-settings`
}
