package huggingfaceupload

var Usage = []string{"hf upload <folder-path> <repo-id>"}

func GetDescription() string {
	return `Upload a model or dataset folder to HuggingFace Hub.`
}

func GetArguments() string {
	return `	folder-path
		Path to the folder to upload.

	repo-id
		The HuggingFace repository ID (e.g., 'username/model-name' or 'username/dataset-name').

	Command options:
		--repo-key
			[Mandatory] The Artifactory repository key to route the upload through.

		--revision
			[Optional] The revision (branch name, tag, or commit hash) to upload to. Default: 'main'.

		--repo-type
			[Optional] The repository type. Can be 'model' or 'dataset'. Default: 'model'.

		--hf-hub-etag-timeout
			[Optional] Timeout in seconds for ETag validation. Default: 86400 (24 hours).

		--hf-hub-download-timeout
			[Optional] Timeout in seconds for Download. Default: 86400 (24 hours).`
}

func GetAIDescription() string {
	return `Upload a HuggingFace model or dataset folder to HuggingFace Hub via an Artifactory HuggingFaceML repository. Files are streamed through the configured Artifactory server, which caches and tracks the artifacts.

When to use:
- Pushing a fine-tuned model to a private HF org behind Artifactory.
- Mirroring a dataset for offline access in regulated environments.

Prerequisites:
- A configured server with an Artifactory HF repository (--repo-key).
- A HuggingFace token in HF_TOKEN if the destination repo is private.
- The folder to upload exists locally.

Common patterns:
  $ jf hf upload ./my-model my-org/my-model --repo-key=hf-local
  $ jf hf upload ./my-dataset my-org/my-dataset --repo-key=hf-local --repo-type=dataset
  $ jf hf upload ./my-model my-org/my-model --repo-key=hf-local --revision=v1.0

Gotchas:
- --repo-key is mandatory; the upload fails fast without it.
- Default --repo-type is 'model'; pass 'dataset' for datasets.
- HF_TOKEN is required by upstream HuggingFace if the target repo is private.
- Large model uploads can take hours; tune --hf-hub-download-timeout and --hf-hub-etag-timeout.

Related: jf hf download`
}
