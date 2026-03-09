package pnpmpublish

var Usage = []string{"pnpm publish [command options]"}

func GetDescription() string {
	return "Packs and deploys the pnpm package to the designated npm repository."
}
