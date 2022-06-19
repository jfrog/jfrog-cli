package cliutils

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/pkg/errors"

	commandUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/stretchr/testify/assert"
)

func TestSplitAgentNameAndVersion(t *testing.T) {
	tests := []struct {
		fullAgentName        string
		expectedAgentName    string
		expectedAgentVersion string
	}{
		{"abc/1.2.3", "abc", "1.2.3"},
		{"abc/def/1.2.3", "abc/def", "1.2.3"},
		{"abc\\1.2.3", "abc\\1.2.3", ""},
		{"abc:1.2.3", "abc:1.2.3", ""},
		{"", "", ""},
	}

	for _, test := range tests {
		actualAgentName, actualAgentVersion := splitAgentNameAndVersion(test.fullAgentName)
		assert.Equal(t, test.expectedAgentName, actualAgentName)
		assert.Equal(t, test.expectedAgentVersion, actualAgentVersion)
	}
}

func TestPrintCommandSummary(t *testing.T) {
	outputBuffer, stderrBuffer, previousLog := tests.RedirectLogOutputToBuffer()
	// Restore previous logger when the function returns
	defer log.SetLogger(previousLog)

	result := &commandUtils.Result{}
	result.SetSuccessCount(1)
	result.SetFailCount(0)
	teastdata := filepath.Join(tests.GetTestResourcesPath(), "reader", "printcommandsummary.json")
	tmpDir, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	err := fileutils.CopyFile(tmpDir, teastdata)
	assert.NoError(t, err)

	reader := content.NewContentReader(filepath.Join(tmpDir, "printcommandsummary.json"), content.DefaultKey)
	result.SetReader(reader)
	assert.NoError(t, err)
	tests := []struct {
		isDetailedSummary bool
		isDeploymentView  bool
		expectedString    string
		expectedError     error
	}{
		{true, false, `"status": "success",`, nil},
		{true, false, `"status": "failure",`, errors.New("test")},
		{false, true, "These files were uploaded:", nil},
		{false, true, ``, errors.New("test")},
	}
	for _, test := range tests {
		err = PrintCommandSummary(result, test.isDetailedSummary, test.isDeploymentView, false, test.expectedError)
		if test.expectedError != nil {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
		var output []byte
		if test.isDetailedSummary {
			output = outputBuffer.Bytes()
			outputBuffer.Truncate(0)
		} else {
			output = stderrBuffer.Bytes()
			stderrBuffer.Truncate(0)
		}
		assert.True(t, strings.Contains(string(output), test.expectedString), fmt.Sprintf("cant find '%s' in '%s'", test.expectedString, string(output)))
	}
}
