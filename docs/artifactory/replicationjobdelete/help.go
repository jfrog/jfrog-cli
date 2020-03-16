package replicationdelete

const Description = "Remove a replication repository job from Artifactory."

var Usage = []string{`jfrog rt rjd <repository key>`}

const Arguments string = `	repository key
		The repository for which the replication configuration will be delete.`
