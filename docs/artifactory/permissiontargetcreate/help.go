package permissiontargetcreate

import "github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"

var Usage = []string{"rt ptc <template path>"}

func GetDescription() string {
	return "Create a new permission target in the JFrog Platform."
}

func GetArguments() string {
	return `	template path
		Specifies the local file system path for the template file to be used for the permission target creation. The template can be created using the "` + coreutils.GetCliExecutableName() + ` rt ptt" command.`
}
