package terraformdocs

var Usage = []string{"terraform <terraform arguments> [command options]"}

func GetDescription() string {
	return "Run Terraform"
}

func GetArguments() string {
	return `	terraform commands
		Arguments and options for the terraform command.`
}

func GetAIDescription() string {
	return `Run a Terraform command (publish, etc.) through JFrog with module resolution against an Artifactory Terraform registry and optional build-info collection.

When to use:
- Publishing Terraform modules to an Artifactory Terraform repo.
- Capturing build-info for IaC pipelines.

Prerequisites:
- A local terraform binary.
- 'jf terraform-config' run once in the project directory.
- A configured server.

Common patterns:
  $ jf terraform publish --build-name=my-iac --build-number=3

Gotchas:
- 'jf terraform-config' must be run first.
- This command focuses on publish; for plan/apply, use terraform directly or wrap with --build-info collection.

Related: jf terraform-config, jf rt build-publish`
}
