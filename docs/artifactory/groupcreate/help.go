package groupcreate

var Usage = []string{"rt gc <group name>"}

func GetDescription() string {
	return "Create a new user group"
}

func GetArguments() string {
	return `	group name
		The name of the new group.`
}

func GetAIDescription() string {
	return `Create an empty user group on the configured Artifactory server. Add users to it with 'jf rt gau' afterward.

When to use:
- Setting up new team groups before binding them to permission targets.

Prerequisites:
- A configured Artifactory server.
- Admin privileges.

Common patterns:
  $ jf rt gc readers
  $ jf rt gc release-managers

Gotchas:
- Group names must be unique on the server.
- The group is created empty; member assignment is a separate 'jf rt gau' call.

Related: jf rt gau, jf rt gdel, jf rt ptc

QA:
Q: Could you guide me through the process of creating a group in JFrog Artifactory considering that the group name is 'group1'?
A: jf rt gc group1

Q: Could you elucidate the approach for forming a group in JFrog Artifactory provided the group name is 'group3'?
A: jf rt gc group3

Q: Could you outline the procedure for constructing a group in JFrog Artifactory with the group name being 'group5'?
A: jf rt gc group5
`
}
