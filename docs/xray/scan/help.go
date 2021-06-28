package scan

const Description = "Execute an Xray scan command on the binaries found according to the given file spec, using the configured Xray details. "

var Usage = []string{"jfrog xr scan [command options] <source pattern> ",
	`jfrog xr scan [command options] --spec=<spec file> `}

const Arguments string = `	source pattern
		Specifies the local file system path to binaries which should be scanned by Xray.
		You can specify multiple files by using wildcards, Ant pattern or a regular expression.
		If you have specified that you are using regular expressions, then the first one used in the argument must be enclosed in parenthesis.`
