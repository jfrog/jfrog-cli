package permissiontargetcreate

const Description = "Create a new permission target in the JFrog Unified Platform."

var Usage = []string{`jfrog rt ptc <template path>`}

const Arguments string = `	template path
		Specifies the local file system path for the template file to be used for the permission target creation. The template can be created using the "jfrog rt ptt" command.`
