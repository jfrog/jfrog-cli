package repodelete

const Description = "Permanently delete repositories with all of its content from Artifactory."

var Usage = []string{`jfrog rt rdel <repository pattern>`}

const Arguments string = `	repository pattern
		Specifies the repository that should be removed. You can use wildcards to specify multiple repositories.`
