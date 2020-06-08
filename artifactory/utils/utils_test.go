package utils

import (
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func TestGetHomeDir(t *testing.T) {
	homeDir, err := cliutils.GetJfrogHomeDir()
	assert.NoError(t, err)
	secPath, err := cliutils.GetJfrogSecurityDir()
	assert.NoError(t, err)
	secFile, err := cliutils.GetJfrogSecurityConfFilePath()
	assert.NoError(t, err)
	certsPath, err := cliutils.GetJfrogCertsDir()
	assert.NoError(t, err)

	assert.Equal(t, secPath, filepath.Join(homeDir, cliutils.JfrogSecurityDirName))
	assert.Equal(t, secFile, filepath.Join(secPath, cliutils.JfrogSecurityConfFile))
	assert.Equal(t, certsPath, filepath.Join(secPath, cliutils.JfrogCertsDirName))
}
