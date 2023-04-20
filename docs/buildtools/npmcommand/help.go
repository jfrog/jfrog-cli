package npmcommand

var Usage = []string{"npm <npm arguments> [command options]"}

func GetDescription() string {
	return "Run npm command."
}

func GetArguments() string {
	return `	ci                        Run npm ci.
	publish, p                Packs and deploys the npm package to the designated npm repository.
	install, i, isntall, add  Run npm install.
	help, h`
}
