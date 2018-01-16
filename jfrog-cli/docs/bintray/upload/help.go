package upload

const Description string = "Upload files."

var Usage = []string{"jfrog bt u [command options] <source path> <target location> [target path]"}

const Arguments string = `	source path
		The first argument specifies the local file system path to artifacts which should be uploaded to Bintray.
		You can specify multiple artifacts by using wildcards or a regular expression as designated by the --regexp command option.
		If you have specified that you are using regular expressions, then the first one used in the argument must be enclosed in parenthesis.

	target location
		The second argument specifies the location within Bintray for uploaded file, in the form of subject/repository/package/version.

	target path
		[Optional]
		This third optional argument lets you specify a path under the target location to where the files should be uploaded.
		NOTE: that the path should end with a slash (/) to indicate a directory, otherwise the last element
		in your path will be interpreted as a file name, and the file you upload will be renamed to that filename.
		For flexibility in specifying the upload path, you can include placeholders in the form of {1}, {2} etc.
		which are replaced by corresponding tokens in the source path that are enclosed in parenthesis.`
