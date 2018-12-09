package upgradeBundleConfig

const Description string = "Configures a release bundle-config."

var Usage = []string{"jfrog rt ubc [command options] [bundle-config ID]",
	"jfrog rt ubc show [bundle-config ID]",
	"jfrog rt ubc [--interactive=<true|false>] delete <bundle-config ID>",
	"jfrog rt ubc [--interactive=<true|false>] clear"}

const Arguments string = `	bundle-config ID
		A unique ID for the new Bundle configuration.

	show
		Shows the stored configuration.
		In case this argument is followed by a configured bundle-config ID, then only this bundle-config is shown.

	delete
		This argument should be followed by a configured bundle-config ID. The configuration for this bundle-config ID will be deleted.

	clear
		Clears all stored configuration.`
