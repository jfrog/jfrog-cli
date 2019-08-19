package dependencies

import (
	"bytes"
	"fmt"
	logUtils "github.com/jfrog/jfrog-cli-go/utils/log"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseRequirementsTxt(t *testing.T) {
	expectedResult := []string {"pk._g3", "SomeDependency", "MyProject", "MyProject", "pkg2", "ProjectB", "ProjectA", "ProjectC", "pk-g1", "subdir", "Some-Dependency"}

	// Create log.
	newLog := log.NewLogger(logUtils.GetCliLogLevel(), nil)
	buffer := &bytes.Buffer{}
	newLog.SetOutputWriter(buffer)
	log.SetLogger(newLog)

	// Create file path.
	pwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	requirementsTestFilePath := filepath.Join(pwd, "testsdata/requirements.txt")

	newExtractor := &requirementsExtractor{requirementsFilePath: requirementsTestFilePath}
	newExtractor.initializeRegExps()
	// Parse the file.
	parseResult, err := newExtractor.parseRequirementsFile()

	// Print results.
	if err != nil {
		log.Info("Error: " + err.Error())
	}
	log.Info(fmt.Sprintf("Result:\n%v", parseResult))

	if !reflect.DeepEqual(expectedResult, parseResult) {
		t.Error("Not equal!!!")
	}
}
