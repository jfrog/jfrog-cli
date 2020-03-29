package repoupdate

const Description = "Update an exiting repository configuration in Artifactory."

var Usage = []string{`jfrog rt ru <template path>`}

const Arguments string = `	template path
		Specifies the local file system path for the template file to be used for the repository update. The template can be created using the "jfrog rt rpt" command.`
