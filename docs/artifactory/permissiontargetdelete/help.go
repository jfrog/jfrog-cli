package permissiontargetdelete

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Permanently delete a permission target."

var Usage = []string{cliutils.CliExecutableName + " rt ptdel <permission target name>"}

const Arguments string = `	permission target name
		Specifies the permission target that should be removed.`
