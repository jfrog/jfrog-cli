package replicationdelete

const Description = "Remove a replication repository from Artifactory."

var Usage = []string{`jfrog rt rpldel <repository key>`}

const Arguments string = `	repository key
		The repository from which the replication will be deleted.`
