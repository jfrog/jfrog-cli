package mvnconfig

var Usage = []string{"mvn-config [command options]"}

func GetDescription() string {
	return "Generate maven build configuration."
}

func GetAIDescription() string {
	return `Write a per-project Maven configuration file (.jfrog/projects/maven.yaml) that tells 'jf mvn' which Artifactory server and repositories to use for resolution and deployment. Run once per project; commit the file or .gitignore it depending on team policy.

When to use:
- Initial setup of a Maven project to build through JFrog.
- Re-pointing an existing project at a different server or repo.

Prerequisites:
- A configured server (jf c add or jf login).
- Knowledge of the resolver (releases/snapshots) and deployer (releases/snapshots) repo keys in Artifactory.
- Run from the project root (the directory containing pom.xml).

Common patterns:
  $ jf mvn-config --server-id-resolve=my-server --repo-resolve-releases=libs-release --repo-resolve-snapshots=libs-snapshot
  $ jf mvn-config --global   # write to ~/.jfrog/projects/ instead of ./.jfrog/

Gotchas:
- Interactive prompts run by default if required flags are missing.
- The generated yaml is read by every subsequent 'jf mvn' from the same directory.
- --global affects all projects on the machine; prefer per-project config for multi-tenant scenarios.

Related: jf mvn, jf gradle-config, jf rt build-publish`
}
