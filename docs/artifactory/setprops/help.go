package setprops

const Description = "Set properties."

var Usage = []string{"jfrog rt sp [command options] <artifacts pattern> <artifact properties>"}

const Arguments string =
`    artifacts pattern
		Specifies the artifacts path in Artifactory, on which the properties are should be set.

	artifact properties
		List of properties in the form of key1=value1;key2=value2,... to be set on the matching artifacts.
`