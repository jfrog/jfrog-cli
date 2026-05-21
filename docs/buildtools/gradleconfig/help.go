package gradleconfig

var Usage = []string{"gradle-config [command options]"}

func GetDescription() string {
	return "Generate gradle build configuration."
}

func GetAIDescription() string {
	return `Write a per-project Gradle configuration file (.jfrog/projects/gradle.yaml) that points 'jf gradle' at the right Artifactory server and repositories for resolution and deployment.

When to use:
- Initial setup of a Gradle project to build through JFrog.
- Switching an existing project to a different repository or server.

Prerequisites:
- A configured server.
- The resolver and deployer repository keys in Artifactory.
- Run from the project root.

Common patterns:
  $ jf gradle-config --server-id-resolve=my-server --repo-resolve=libs-release --repo-deploy=libs-snapshot
  $ jf gradle-config --global

Gotchas:
- Interactive prompts trigger when required flags are missing.
- --global writes to ~/.jfrog/projects/ and affects all projects on the machine.
- The generated config is consumed by 'jf gradle' in the same directory; running gradle from elsewhere will not find it.

Related: jf gradle, jf mvn-config`
}
