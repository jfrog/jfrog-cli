package deleteprops

const Description = "Delete properties on existing files in Artifactory."

var Usage = []string{"jfrog rt delp [command options] <artifacts pattern> <artifact properties>",
	"jfrog rt delp <artifact properties> --spec=<File Spec path> [command options]"}

const Arguments string = `	artifacts pattern
		Properties of artifacts that match this pattern will be removed.

	artifact properties
		The list of properties, in the form of key1,key2,..., to be removed from the matching artifacts.`
