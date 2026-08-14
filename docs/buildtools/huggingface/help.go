package huggingface

var Usage = []string{"hf download <model-name>",
	"hf upload <folder-path> <repo-id>"}

func GetDescription() string {
	return `Download or upload models/datasets from/to HuggingFace Hub.`
}

func GetAIDescription() string {
	return `Parent command for HuggingFace Hub operations routed through an Artifactory HuggingFaceML repository. Use the 'upload' and 'download' subcommands.

When to use:
- Caching ML models or datasets through Artifactory for compliance/auditability.
- Mirroring private HF repos into Artifactory.

Prerequisites:
- A configured server with an HF-type Artifactory repo.

Common patterns:
  $ jf hf download bert-base-uncased --repo-key=hf-virtual
  $ jf hf upload ./my-model my-org/my-model --repo-key=hf-local

Gotchas:
- 'jf hf' alone shows subcommand help; an explicit subcommand is required to do work.

Related: jf hf upload, jf hf download`
}
