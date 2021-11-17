package scan

const Description = "Scan files located on the local file-system with Xray."

var Usage = []string{"jfrog scan [command options] <source pattern> ",
	`jfrog scan [command options] --spec=<spec file> `}

const Arguments string = `	source pattern
		Specifies the local file system path of the files to be scanned.
		You can specify multiple files by using wildcards, Ant pattern or a regular expression.
		If you have specified that you are using regular expressions, then the first one used in the argument must be enclosed in parenthesis.`
