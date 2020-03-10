package repotemplate

const Description = "Create a new repository in Artifactory."

var Usage = []string{`jfrog rt rc <template path>`}

const Arguments string = `	template path
		Specifies the local file system path for the template file to bw used for the repository creation.`
