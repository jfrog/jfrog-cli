package pnpmconfig

var Usage = []string{"pnpm-config [command options]"}

func GetDescription() string {
	return "Generate pnpm configuration."
}

func GetAIDescription() string {
	return `Write a per-project pnpm configuration (.jfrog/projects/pnpm.yaml) that routes 'jf pnpm' through an Artifactory npm repository.

When to use:
- Initial setup of a pnpm project for JFrog.

Prerequisites:
- A configured server.
- The Artifactory npm/pnpm repository key.
- Run from the project root.

Common patterns:
  $ jf pnpm-config --server-id-resolve=my-server --repo-resolve=npm-virtual

Gotchas:
- Interactive prompts run when required flags are missing.
- Workspace projects need the config in the workspace root, not each package.
- This does not configure the pnpm client itself. It is read only by 'jf pnpm' commands; a plain 'pnpm install' keeps resolving from its own configuration, which this command never touches. To point the client itself at Artifactory for every project on the machine, run 'jf setup pnpm' instead - the two are independent and can even name different repositories.

Related: jf pnpm, jf npm-config, jf setup pnpm`
}
