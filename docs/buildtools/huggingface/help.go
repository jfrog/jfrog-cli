package huggingface

var Usage = []string{"hf download <model-name>",
	"hf upload <folder-path> <repo-id>"}

func GetDescription() string {
	return `Download or upload models/datasets from/to HuggingFace Hub.`
}

func GetArguments() string {
	return `	download <model-name>
		Download a model/dataset from HuggingFace Hub.
		model-name
			The HuggingFace model repository ID (e.g., 'bert-base-uncased' or 'username/model-name').

	upload <folder-path> <repo-id>
		Upload a model or dataset folder to HuggingFace Hub.
		folder-path
			Path to the folder to upload.
		repo-id
			The HuggingFace repository ID (e.g., 'username/model-name' or 'username/dataset-name').

	Command options:
		--revision
			[Optional] The revision (commit hash, branch name, or tag) to download/upload. Defaults to main branch if not specified.

		--repo-type
			[Optional] The repository type. Can be 'model' or 'dataset'. Default: 'model'.

		--etag-timeout
			[Optional] [Download only] Timeout in seconds for ETag validation. Default: 86400 seconds (24 hours).`
}
