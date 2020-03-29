package repodelete

const Description = "Permanently delete a repository with all of its content from Artifactory."

var Usage = []string{`jfrog rt rdel <repository key>`}

const Arguments string = `	repository key
		Specifies the repository that should be removed.`
