package npmci

var Usage = []string{"rt npmci [npm ci args] [command options]"}

func GetDescription() string {
	return "Run npm ci."
}

func GetArguments() string {
	return `	npm ci args
		The npm ci args to run npm ci.`
}
