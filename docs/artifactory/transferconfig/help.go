package transferconfig

var Usage = []string{"rt transfer-config [command options] <source-server-id> <target-server-id>"}

func GetDescription() string {
	return `Copy the full configuration from a source Artifactory server to a target server`
}

func GetArguments() string {
	return `	source-server-id
		The source server ID. The configuration will be exported from this server.

	target-server-id
		The target server ID. The configuration will be imported to this server.
		[Warning] This action will wipe all Artifactory content in this target server.`
}

func GetAIDescription() string {
	return `Export the full Artifactory configuration from a source server and import it into a target server. Used as the configuration phase of a self-hosted-to-cloud (or vice versa) migration. The target's existing content is wiped.

When to use:
- Pre-step before 'jf rt transfer-files' in a cross-instance migration.
- One-shot configuration mirroring between two Artifactory instances.

Prerequisites:
- Both source and target server IDs must be configured via 'jf c add' with admin tokens.
- The target server must be empty or expected to be overwritten.
- Source and target Artifactory versions must be compatible (consult JFrog migration docs).

Common patterns:
  $ jf rt transfer-config source-prod target-cloud
  $ jf rt transfer-config source-prod target-cloud --include-repos="libs-*" --exclude-repos="*-deprecated"
  $ jf rt transfer-config source-prod target-cloud --force

Gotchas:
- DESTRUCTIVE on the target: existing repos, permissions, users, and other config are replaced.
- The command prompts for confirmation; --force skips it (dangerous).
- Run when the source server is quiet to avoid drift during the export.
- Does NOT transfer file content; pair with 'jf rt transfer-files' afterwards.

Related: jf rt transfer-config-merge, jf rt transfer-files, jf rt transfer-settings`
}
