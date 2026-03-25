package huggingface

var Usage = []string{"hf download <model-name>",
	"hf upload <folder-path> <repo-id>"}

func GetDescription() string {
	return `Download or upload models/datasets from/to HuggingFace Hub.`
}
