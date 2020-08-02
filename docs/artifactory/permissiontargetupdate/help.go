package permissiontargetupdate

const Description = "Update a permission target in the JFrog Platform."

var Usage = []string{`jfrog rt ptu <template path>`}

const Arguments string = `	template path
		Specifies the local file system path for the template file to be used for the permission target update. The template can be created using the "jfrog rt ptu" command.`
