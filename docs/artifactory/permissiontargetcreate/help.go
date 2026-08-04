package permissiontargetcreate

import "github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"

var Usage = []string{"rt ptc <template path>"}

func GetDescription() string {
	return "Create a new permission target in the JFrog Platform."
}

func GetArguments() string {
	return `	template path
		Specifies the local file system path for the template file to be used for the permission target creation. The template can be created using the "` + coreutils.GetCliExecutableName() + ` rt ptt" command.`
}

func GetAIDescription() string {
	return `Create a new permission target on the configured Artifactory server from a local JSON template. Permission targets bind repositories to users/groups with specified actions (read, deploy, delete, manage).

When to use:
- Provisioning a new permission target during environment setup.
- Scripting RBAC alongside repository creation.

Prerequisites:
- A configured Artifactory server.
- Admin privileges.
- A permission target JSON template (generate with 'jf rt ptt' or hand-write per the Artifactory schema).

Common patterns:
  $ jf rt ptc ./my-perm-target.json
  $ jf rt ptc ./my-perm-target.json --vars=env=prod

Gotchas:
- Fails if a permission target with the same name already exists; use 'jf rt ptu' to update.
- The JSON schema is strict; missing required fields produce a 400 from the server.
- --vars allows templated variable substitution (key=value pairs).

Related: jf rt ptt, jf rt ptu, jf rt ptdel`
}
