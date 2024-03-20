package export

var Usage = []string{"rbe <release bundle name> <release bundle version> [target pattern]"}

func GetDescription() string {
	return "Triggers the Export process and downloads the Release Bundle archive"
}

func GetArguments() string {
	return `	release bundle name
		Name of the Release Bundle to export.

	release bundle version
		Version of the Release Bundle to export.

	target pattern
		The third argument is optional and specifies the local file system target path.
			If the target path ends with a slash, the path is assumed to be a directory.
			For example, if you specify the target as "repo-name/a/b/", then "b" is assumed to be a directory into which files should be downloaded.
			If there is no terminal slash, the target path is assumed to be a file to which the downloaded file should be renamed.
			For example, if you specify the target as "a/b", the downloaded file is renamed to "b". `
}
