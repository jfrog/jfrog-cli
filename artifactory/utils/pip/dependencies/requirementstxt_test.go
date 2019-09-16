package dependencies

import (
	"github.com/jfrog/jfrog-cli-go/utils/log"
	"github.com/jfrog/jfrog-cli-go/utils/tests"
	"os"
	"path/filepath"
	"testing"
)


func TestConcatenateLines(t *testing.T) {
	tests := []struct {
		firstLine     string
		secondLine     string
		expected string
	}{
		{"pkg1==1.0.1 \\", "--hash=sha256:abcdefg1", "pkg1==1.0.1 --hash=sha256:abcdefg1"},
		{"pkg2==1.0.2 \\\\\\", "--hash=sha256:abcdefg2", "pkg2==1.0.2 --hash=sha256:abcdefg2"},
		{"pk\\", "g3", "pkg3"},
		{"pkg4==1.0.4\\\\\\", "\\", "pkg4==1.0.4\\"},
	}

	for _, test := range tests {
		actualValue := concatenateLines(test.firstLine, test.secondLine)
		if actualValue != test.expected {
			t.Errorf("Expected value: %s, got: %s.", test.expected, actualValue)
		}
	}
}

func TestParseRequirementsFile(t *testing.T) {
	log.SetDefaultLogger()
	expectedResult := []string {
		"pk1._test",
		"pkg2_test",
		"pkg3",
		"pkg4",
		"pkg5",
		"pkg6",
		"pkg7",
		"pkg8",
		"pk-g9",
		"pkg10",
		"pkg11",
		"pkg13",
		"pkg14",
		"pkg17",
		"pkg18",
		"pkg19",
		"pkg21",
		"pkg22"}

	// Create file path.
	pwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	requirementsTestFilePath := filepath.Join(pwd, "testsdata/requirements.txt")

	newExtractor := &requirementsExtractor{requirementsFilePath: requirementsTestFilePath}
	err = newExtractor.initializeRegExps()
	if err != nil {
		t.Error(err)
	}

	// Parse the file.
	parseResult, err := newExtractor.parseRequirementsFile()

	// Check results.
	if err != nil {
		t.Error(err)
	}

	err = tests.ValidateListsIdentical(expectedResult, parseResult)
	if err != nil {
		t.Error(err)
	}
}
