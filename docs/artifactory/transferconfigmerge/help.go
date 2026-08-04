package transferconfigmerge

var Usage = []string{"rt transfer-config-merge [command options] <source-server-id> <target-server-id>"}

func GetDescription() string {
	return "Merge projects and repositories from a source Artifactory instance to a target Artifactory instance, if no conflicts are found"
}

func GetArguments() string {
	return `	source-server-id
		The source server ID. The configuration will be exported from this server.

	target-server-id
		The target server ID. The configuration will be imported to this server.`
}

func GetAIDescription() string {
	return `Merge repository and project definitions from a source Artifactory instance into a target instance non-destructively. Unlike 'transfer-config', this preserves existing target content and aborts on naming conflicts.

When to use:
- Migrating selected projects/repos into an existing Artifactory deployment.
- Bringing two instances closer together without overwriting either side.

Prerequisites:
- Both source and target server IDs configured with admin credentials.
- Compatible Artifactory versions on both sides.

Common patterns:
  $ jf rt transfer-config-merge source-prod target-cloud
  $ jf rt transfer-config-merge source-prod target-cloud --include-projects="proj-*" --exclude-projects="legacy-*"

Gotchas:
- Aborts and reports conflicts; resolve them manually before re-running.
- Only repos and projects are merged; permission targets, users, and other config are NOT.
- Run during low-traffic windows on the source to avoid mid-export drift.

Related: jf rt transfer-config, jf rt transfer-files`
}
