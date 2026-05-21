package gradle

import "github.com/jfrog/jfrog-cli/docs/common"

var Usage = []string{"gradle <tasks and options> [command options]"}

var EnvVar = []string{common.JfrogCliReleasesRepo, common.JfrogCliDependenciesDir}

func GetDescription() string {
	return "Run Gradle build."
}

func GetArguments() string {
	return `	tasks and options
		Tasks and options to run with gradle command. For example, -b path/to/build.gradle.`
}

func GetAIDescription() string {
	return `Run a Gradle build with JFrog instrumentation: resolves dependencies and publishes artifacts via Artifactory, optionally collecting build-info. Wraps the local gradle binary; tasks and flags after the command are passed through.

When to use:
- Building Gradle projects that resolve and publish to Artifactory.
- Producing a JFrog build-info record alongside the publish.

Prerequisites:
- A local gradle binary on PATH (or use the project's gradlew).
- 'jf gradle-config' run once in the project directory.
- A configured server referenced by 'jf gradle-config'.

Common patterns:
  $ jf gradle clean build
  $ jf gradle artifactoryPublish --build-name=my-build --build-number=1
  $ jf gradle build -b path/to/build.gradle

Gotchas:
- 'jf gradle-config' must be run first; the command fails with a clear error if missing.
- --build-name and --build-number are required together for build-info.
- Gradle daemon caches can hide config changes; use --no-daemon when debugging.
- Set JFROG_CLI_RELEASES_REPO to fetch the Gradle extractor through a private repo (air-gapped builds).

Related: jf gradle-config, jf rt build-publish`
}
