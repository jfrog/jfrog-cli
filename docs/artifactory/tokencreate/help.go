package tokencreate

const Description = "Creates an access token. By default an user-scoped token will be created, unless groups and/or admin-privileges are specified."

var Usage = []string{`jfrog rt tc <user name>`}

const Arguments string = `	user name
		The user name for which this token is created.`
