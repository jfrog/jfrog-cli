package conan

var Usage = []string{"conan <conan args> [command options]"}

func GetDescription() string {
	return "Run native conan command."
}

func GetArguments() string {
	return `	conan sub-command
		Arguments and options for the conan command.
		
		Examples:
		- jf conan install . --build=missing
		- jf conan create . --name=hello --version=1.0`
}
