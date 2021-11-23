package repoupdate

import "github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"

var Usage = []string{"rt ru <template path>"}

func GetDescription() string {
	return "Update an exiting repository configuration in Artifactory."
}

func GetArguments() string {
	return `	template path
		Specifies the local file system path for the template file to be used for the repository update. The template can be created using the "` + coreutils.GetCliExecutableName() + ` rt rpt" command.`
}
