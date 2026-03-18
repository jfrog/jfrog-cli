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
			[Optional] The repository type. Can be 'model' or 'dataset'. Default: 'model'.`
}
