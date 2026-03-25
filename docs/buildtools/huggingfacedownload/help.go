package huggingfacedownload

var Usage = []string{"hf download <model-name>"}

func GetDescription() string {
	return `Download a model or dataset from HuggingFace Hub.`
}

func GetArguments() string {
	return `	model-name
		The HuggingFace model repository ID (e.g., 'bert-base-uncased' or 'username/model-name').

	Command options:
		--repo-key
			[Mandatory] The Artifactory repository key to route the download through.

		--revision
			[Optional] The revision (commit hash, branch name, or tag) to download. Default: 'main'.

		--repo-type
			[Optional] The repository type. Can be 'model' or 'dataset'. Default: 'model'.

		--hf-hub-etag-timeout
			[Optional] Timeout in seconds for ETag validation. Default: 86400 (24 hours).

		--hf-hub-download-timeout
			[Optional] Timeout in seconds for Download. Default: 86400 (24 hours).`
}
