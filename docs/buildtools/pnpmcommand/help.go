package pnpmcommand

var Usage = []string{"pnpm <pnpm arguments> [command options]"}

func GetDescription() string {
	return "Run pnpm command."
}

func GetArguments() string {
	return `	install, i                Run pnpm install.
	publish, p                Packs and deploys the pnpm package to the designated npm repository.
	help, h`
}
