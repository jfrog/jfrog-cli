package yarnconfig

var Usage = []string{"yarn-config [command options]"}

func GetDescription() string {
	return "Generate Yarn configuration."
}

func GetAIDescription() string {
	return `Write a per-project Yarn configuration (.jfrog/projects/yarn.yaml) that points 'jf yarn' at an Artifactory npm/yarn repository for resolution.

When to use:
- First-time setup of a Yarn project to route through JFrog.

Prerequisites:
- A configured server.
- The Artifactory yarn/npm repository key.
- Run from the project root.

Common patterns:
  $ jf yarn-config --server-id-resolve=my-server --repo-resolve=npm-virtual

Gotchas:
- Interactive prompts run when required flags are missing.
- Yarn berry projects may need additional manual .yarnrc.yml tweaks beyond what this command writes.
- This does not configure the yarn client itself. It is read only by 'jf yarn' commands; a plain 'yarn install' keeps resolving from its own configuration, which this command never touches. To point the client itself at Artifactory for every project on the machine, run 'jf setup yarn' instead - the two are independent and can even name different repositories.

Related: jf yarn, jf npm-config, jf setup yarn`
}
