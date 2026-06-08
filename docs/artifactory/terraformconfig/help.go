package terraformconfig

var Usage = []string{"terraform-config [command options]"}

func GetDescription() string {
	return "Generate Terraform configuration"
}

func GetAIDescription() string {
	return `Write a per-project Terraform configuration (.jfrog/projects/terraform.yaml) so 'jf terraform' resolves modules through an Artifactory Terraform registry.

When to use:
- First-time setup of a Terraform project to use a private module registry.

Prerequisites:
- A configured server.
- The Artifactory Terraform repository key.

Common patterns:
  $ jf terraform-config --server-id-deploy=my-server --repo-deploy=terraform-local

Gotchas:
- Interactive prompts trigger when required flags are missing.
- The Terraform .terraformrc may need additional credentials for the Artifactory host; consult JFrog docs.

Related: jf terraform`
}
