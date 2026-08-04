package conanconfig

var Usage = []string{"conan-config [command options]"}

func GetDescription() string {
	return "Generate conan build configuration."
}

func GetAIDescription() string {
	return `Write a per-project Conan configuration (.jfrog/projects/conan.yaml) and configure the local Conan remote so 'jf conan' resolves through an Artifactory Conan repository.

When to use:
- First-time setup of a C/C++ project using Conan against a private Artifactory Conan repo.

Prerequisites:
- A configured server.
- The Artifactory Conan repository key.

Common patterns:
  $ jf conan-config --server-id-resolve=my-server --repo-resolve=conan-virtual

Gotchas:
- Modifies the local Conan remotes configuration (~/.conan/remotes.json or equivalent).
- Interactive prompts trigger when required flags are missing.

Related: jf conan, jf rt build-publish`
}
