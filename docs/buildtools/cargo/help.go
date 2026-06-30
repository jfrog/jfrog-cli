package cargo

var Usage = []string{"jf cargo <cargo args> [command options]"}

func GetDescription() string {
	return "Run Cargo (Rust) commands with JFrog build-info collection."
}

func GetArguments() string {
	return `	cargo command
		The cargo command to run, e.g. build, publish, package, add.`
}

func GetAIDescription() string {
	return "Runs native cargo and collects build-info for Rust projects."
}
