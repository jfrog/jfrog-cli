package vscode

var Usage = []string{"vscode config <repository-url> [command options]"}

func GetDescription() string {
	return "Configure VSCode to use JFrog Artifactory extensions repository"
}

func GetArguments() string {
	return `	repository-url
		The full URL to your JFrog Artifactory extensions repository.
		Example: http://productdemo.jfrog.io/artifactory/api/vscodeextensions/extensions-remote/_apis/public/gallery`
}
