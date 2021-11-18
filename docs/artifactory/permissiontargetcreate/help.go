package permissiontargetcreate

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Create a new permission target in the JFrog Platform."

var Usage = []string{cliutils.CliExecutableName + " rt ptc <template path>"}

const Arguments string = `	template path
		Specifies the local file system path for the template file to be used for the permission target creation. The template can be created using the "` + cliutils.CliExecutableName + ` rt ptt" command.`
