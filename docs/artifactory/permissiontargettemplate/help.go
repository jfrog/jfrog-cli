package permissiontargettemplate

var Usage = []string{"rt ptt <template path>"}

func GetDescription() string {
	return "Create a JSON template for a permission target creation or replacement."
}

func GetArguments() string {
	return `	template path
		Specifies the local file system path for the template file.`
}

func GetAIDescription() string {
	return `Interactively generate a permission target JSON template at the specified local path. The template is the input for 'jf rt ptc' (create) and 'jf rt ptu' (update). Walks the user through repos, users, groups, and actions.

When to use:
- First step in scripting permission target creation: produce the template, then edit and apply via ptc/ptu.

Prerequisites:
- A configured Artifactory server.
- Admin privileges to read repo lists during the interactive flow.

Common patterns:
  $ jf rt ptt ./my-perm-target.json

Gotchas:
- This is an interactive command; does not work in non-TTY environments.
- The generated file is plain JSON; edit it before passing to ptc/ptu for production use.

Related: jf rt ptc, jf rt ptu, jf rt ptdel`
}
