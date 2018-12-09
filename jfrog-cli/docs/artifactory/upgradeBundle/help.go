package upgradeBundle

const Description = "Checks for new release-bundle's version in the edge node. If new version exists, this command runs an installation script."

var Usage = []string{"jfrog rt ub [command options] <bundle-config ID>"}

const Arguments string = `	bundle-config ID
		 Bundle-config to run. Bundle-configs can be configured with 'jfrog rt ubc'.`
