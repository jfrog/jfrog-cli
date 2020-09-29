package utils

import (
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"io/ioutil"
)

// Gets all the installed plugins' names by looping over the plugins dir.
func GetAllPluginsNames() ([]string, error) {
	pluginsDir, err := coreutils.GetJfrogPluginsDir()
	if err != nil {
		return []string{}, err
	}
	exists, err := fileutils.IsDirExists(pluginsDir, false)
	if err != nil || !exists {
		return []string{}, err
	}

	files, err := ioutil.ReadDir(pluginsDir)
	if err != nil {
		return []string{}, errorutils.CheckError(err)
	}

	var plugins []string
	for _, f := range files {
		if f.IsDir() {
			logSkippablePluginsError("unexpected directory in plugins directory", f.Name(), nil)
			continue
		}
		plugins = append(plugins, f.Name())
	}
	return plugins, nil
}
