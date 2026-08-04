package transferfiles

var Usage = []string{"rt transfer-files [command options] <source-server-id> <target-server-id>"}

func GetDescription() string {
	return "Transfer files from one Artifactory to another."
}

func GetArguments() string {
	return `	source-server-id
		Server ID of the Artifactory instance to transfer from.

	target-server-id
		Server ID of the Artifactory instance to transfer to.`
}

func GetAIDescription() string {
	return `Stream artifact contents from a source Artifactory to a target Artifactory. Designed for large-scale migrations: resumable, throttleable, and parallelized. Run AFTER 'jf rt transfer-config' so repo definitions exist on the target.

When to use:
- Migrating petabytes of binary content from self-hosted to cloud (or between any two Artifactory instances).
- Resuming a previously interrupted transfer.

Prerequisites:
- Both source and target server IDs configured with admin tokens.
- 'jf rt transfer-config' (or transfer-config-merge) already run so repos exist on the target.
- 'jf rt transfer-plugin-install' installed on the source server's primary node.
- 'jf rt transfer-settings' tuned for throughput.

Common patterns:
  $ jf rt transfer-files source-prod target-cloud
  $ jf rt transfer-files source-prod target-cloud --include-repos="libs-*" --status
  $ jf rt transfer-files source-prod target-cloud --status   # check progress

Gotchas:
- Requires the data-transfer plugin on the source (install via 'jf rt transfer-plugin-install').
- Long-running; expect hours-to-days for large stores. Run inside a long-lived session (nohup, tmux, screen).
- Throughput is governed by 'jf rt transfer-settings'; defaults are conservative.
- Idempotent on file content but not on repo property changes; re-running after a target-side delete causes redelivery.

Related: jf rt transfer-config, jf rt transfer-settings, jf rt transfer-plugin-install`
}
