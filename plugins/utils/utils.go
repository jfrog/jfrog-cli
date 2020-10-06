package utils

import "github.com/jfrog/jfrog-cli-core/utils/coreutils"

func GetPluginExecutableName(plugin string) string {
	if coreutils.IsWindows() {
		return plugin + ".exe"
	}
	return plugin
}
