package jetbrains

var Usage = []string{"jetbrains config <repository-url>"}

func GetDescription() string {
	return "Configure JetBrains IDEs to use JFrog Artifactory plugins repository"
}

func GetArguments() string {
	return `	repository-url
		The full URL to your JFrog Artifactory plugins repository.
		Example: http://productdemo.jfrog.io/artifactory/api/jetbrains/jetbrains-remote`
}
