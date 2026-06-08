package npmconfig

var Usage = []string{"npm-config [command options]"}

func GetDescription() string {
	return "Generate npm configuration."
}

func GetAIDescription() string {
	return `Write a per-project npm configuration (.jfrog/projects/npm.yaml) that binds 'jf npm' to a server and repository. Required before 'jf npm install' or 'jf npm publish' will work as intended.

When to use:
- Initial setup of an npm project to build through JFrog.
- Switching to a different npm repo or server.

Prerequisites:
- A configured server (jf c add or jf login).
- The npm repository key in Artifactory (a virtual repo for resolve, a local repo for deploy).
- Run from the project root (where package.json lives).

Common patterns:
  $ jf npm-config --server-id-resolve=my-server --repo-resolve=npm-virtual --repo-deploy=npm-local
  $ jf npm-config --global

Gotchas:
- Prompts run when required flags are missing; pass --repo-resolve / --repo-deploy to avoid them.
- --global writes to ~/.jfrog/projects/ and affects all subsequent jf npm runs on the machine.
- A separate npm-config is needed per project; running 'jf npm' from a directory without one fails.

Related: jf npm, jf yarn-config, jf pnpm-config`
}
