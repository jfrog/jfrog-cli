package conan

var Usage = []string{"conan <conan args> [command options]"}

func GetDescription() string {
	return "Run native conan command."
}

func GetArguments() string {
	return `	conan sub-command
		Arguments and options for the conan command.

		Examples:
		- jf conan install . --build=missing
		- jf conan create . --name=hello --version=1.0`
}

func GetAIDescription() string {
	return `Run a Conan C/C++ package manager command through JFrog, with remote resolution against an Artifactory Conan repository and optional build-info collection.

When to use:
- Resolving Conan dependencies from a private Conan repo.
- Producing build-info for C/C++ projects.

Prerequisites:
- A local conan binary (Conan 1.x or 2.x).
- 'jf conan-config' run once in the project directory.
- A configured server.

Common patterns:
  $ jf conan install . --build=missing
  $ jf conan create . --name=hello --version=1.0
  $ jf conan upload mypkg/1.0@ --build-name=my-build --build-number=1

Gotchas:
- 'jf conan-config' must be run first.
- Conan 1.x and 2.x have different command syntax; jf passes args through verbatim.

Related: jf conan-config, jf rt build-publish`
}
