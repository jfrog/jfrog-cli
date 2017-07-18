package setprops

const Description = "Set properties on existing files in Artifactory."

var Usage = []string{"jfrog rt sp [command options] <artifacts pattern> <artifact properties>"}

const Arguments string =
`	artifacts pattern
		Artifacts that match the pattern will be set with the specified properties.

	artifact properties
		The list of properties, in the form of key1=value1;key2=value2,..., to be set on the matching artifacts.`