package scan

var Usage = []string{"scan [command options] <source pattern> ",
	"scan [command options] --spec=<spec file> "}

func GetDescription() string {
	return "Scan files located on the local file-system with Xray."
}

func GetArguments() string {
	return `	source pattern
		Specifies the local file system path of the files to be scanned.
		You can specify multiple files by using wildcards, Ant pattern or a regular expression.
		If you have specified that you are using regular expressions, then the first one used in the argument must be enclosed in parenthesis.`
}
