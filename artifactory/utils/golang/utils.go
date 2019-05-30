package golang

import (
	"github.com/jfrog/gocmd/cmd"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"path/filepath"
)

func LogGoVersion() error {
	output, err := cmd.GetGoVersion()
	if err != nil {
		return errorutils.CheckError(err)
	}
	log.Info("Using go:", output)
	return nil
}

// Checks if go.yaml file exists. First looks for the file in project dir. If not found, looks in JFrog home dir
func IsGoConfigExists() (configFilePath string, exists bool, err error) {
	projectDir, exists, err := fileutils.FindUpstream(".jfrog", fileutils.Dir)
	if err != nil {
		return
	}

	yamlFile := filepath.Join("projects", "go.yaml")
	if exists {
		// Check for the Go yaml configuration file
		// If exists, use the config.
		// If not fall back.
		configFilePath = filepath.Join(projectDir, ".jfrog", yamlFile)
		exists, err = fileutils.IsFileExists(configFilePath, false)
		if err != nil {
			return
		}

		if exists {
			return
		}
	}
	// If missing in the root project, check in the home dir
	jfrogHomeDir, err := config.GetJfrogHomeDir()
	if err != nil {
		return
	}
	configFilePath = filepath.Join(jfrogHomeDir, yamlFile)
	exists, err = fileutils.IsFileExists(configFilePath, false)
	return
}
