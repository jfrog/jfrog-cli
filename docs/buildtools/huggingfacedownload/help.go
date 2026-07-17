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

func GetAIDescription() string {
	return `Download a HuggingFace model or dataset from HuggingFace Hub through an Artifactory HuggingFaceML repository, which caches it for future requests.

When to use:
- Pulling a model in CI from a private HF mirror in Artifactory.
- Caching public HF models locally to reduce egress and ensure availability.

Prerequisites:
- A configured server with an Artifactory HF repository (--repo-key).
- HF_TOKEN if the source repo is private.

Common patterns:
  $ jf hf download bert-base-uncased --repo-key=hf-virtual
  $ jf hf download my-org/my-dataset --repo-key=hf-virtual --repo-type=dataset
  $ jf hf download my-org/my-model --repo-key=hf-virtual --revision=v1.0

Gotchas:
- --repo-key is mandatory.
- Default --repo-type is 'model'.
- The first download of a large model takes a long time; subsequent downloads are served from Artifactory cache.

Related: jf hf upload`
}
