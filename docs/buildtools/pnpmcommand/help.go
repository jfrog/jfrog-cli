package pnpmcommand

var Usage = []string{"pnpm <pnpm arguments> [command options]"}

func GetDescription() string {
	return "Run pnpm command."
}

func GetArguments() string {
	return `	ci                        Run pnpm ci.
	publish, p                Packs and deploys the pnpm package to the designated npm repository.
	install, i, add           Run pnpm install.
	help, h`
}
