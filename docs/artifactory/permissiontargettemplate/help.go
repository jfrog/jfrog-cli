package permissiontargettemplate

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Create a JSON template for a permission target creation or replacement."

var Usage = []string{cliutils.CliExecutableName + " rt ptt <template path>"}

const Arguments string = `	template path
		Specifies the local file system path for the template file.`
