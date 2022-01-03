package npminstall

var Usage = []string{"rt npmi [npm install args] [command options]"}

func GetDescription() string {
	return "Run npm install."
}

func GetArguments() string {
	return `	npm install args
		The npm install args to run npm install. For example, --global.`
}
