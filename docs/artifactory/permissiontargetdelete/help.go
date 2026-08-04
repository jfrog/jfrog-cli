package permissiontargetdelete

var Usage = []string{"rt ptdel <permission target name>"}

func GetDescription() string {
	return "Permanently delete a permission target."
}

func GetArguments() string {
	return `	permission target name
		Specifies the permission target that should be removed.`
}

func GetAIDescription() string {
	return `Permanently remove a permission target from Artifactory. Affected users/groups lose the binding immediately on next request.

When to use:
- Removing an obsolete permission target during cleanup.
- Resetting an RBAC misconfiguration before re-creating it.

Prerequisites:
- A configured Artifactory server.
- Admin privileges.

Common patterns:
  $ jf rt ptdel my-perm-target
  $ jf rt ptdel my-perm-target --quiet

Gotchas:
- No undo; the permission target is gone.
- Removing a permission target does NOT delete the repos, users, or groups it referenced.
- --quiet skips the confirmation prompt; useful in CI.

Related: jf rt ptc, jf rt ptu`
}
