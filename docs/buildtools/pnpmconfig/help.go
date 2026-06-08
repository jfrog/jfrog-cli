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

Related: jf pnpm, jf npm-config`
}
