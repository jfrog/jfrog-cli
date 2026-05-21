package permissiontargetupdate

import "github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"

var Usage = []string{"rt ptu <template path>"}

func GetDescription() string {
	return "Update a permission target in the JFrog Platform."
}

func GetArguments() string {
	return `	template path
		Specifies the local file system path for the template file to be used for the permission target update. The template can be created using the "` + coreutils.GetCliExecutableName() + ` rt ptu" command.`
}

func GetAIDescription() string {
	return `Replace an existing permission target's definition from a local JSON template. Differs from ptc in that the target must already exist.

When to use:
- Modifying RBAC after the fact (adding repos, users, or actions to a permission target).
- Drift correction in GitOps-style RBAC management.

Prerequisites:
- A configured Artifactory server.
- Admin privileges.
- The permission target must already exist (use 'jf rt ptc' otherwise).

Common patterns:
  $ jf rt ptu ./my-perm-target.json
  $ jf rt ptu ./my-perm-target.json --vars=env=prod

Gotchas:
- The update is a full replacement; missing fields revert to defaults.
- Failures roll back partially; verify with 'jf api /artifactory/api/v2/security/permissions/<name>'.

Related: jf rt ptt, jf rt ptc, jf rt ptdel`
}
