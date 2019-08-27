package dependencies

import (
	"bytes"
	"encoding/json"
	"fmt"
	logUtils "github.com/jfrog/jfrog-cli-go/utils/log"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"os"
	"path/filepath"
	"testing"
)

func TestParsePipDepTree(t *testing.T) {
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
	pipdeptreeTestFilePath := filepath.Join(pwd, "testsdata/pipdeptree_output")

	// Read file.
	content, err := fileutils.ReadFile(pipdeptreeTestFilePath)
	if err != nil {
		t.Error("Failed reading file!!!")
	}

	// Parse content.
	depTree, err := parsePipDepTreeOutput(content)
	if err != nil {
		t.Error("Failed parsing dep tree!!!")
	}

	// Print results.
	//log.Info(fmt.Sprintf("Result:\n%+v\n", depTree))
	s, _ := json.MarshalIndent(depTree, "", "\t")
	log.Info(fmt.Sprintf("Result:\n%s", s))
}

func TestRunPipDepTreeAndParse(t *testing.T) {
	// Create log.
	newLog := log.NewLogger(logUtils.GetCliLogLevel(), nil)
	buffer := &bytes.Buffer{}
	newLog.SetOutputWriter(buffer)
	log.SetLogger(newLog)

	pythonPath := "/Users/barb/trash/venv-test2/bin/python"
	pathVar := os.Getenv("PATH")
	os.Setenv("PATH", "/Users/barb/trash/venv-test2/bin")
	defer os.Setenv("PATH", pathVar)

	// Run.
	depTree, err := BuildPipDependencyMap(pythonPath, nil)
	if err != nil {
		t.Error("FAILED!!!!")
	}

	// Print results.
	//log.Info(fmt.Sprintf("Result:\n%+v\n", depTree))
	s, _ := json.MarshalIndent(depTree, "", "\t")
	log.Info(fmt.Sprintf("Result:\n%s", s))
}

func TestExtractDependencies(t *testing.T) {
	// Create log.
	newLog := log.NewLogger(logUtils.GetCliLogLevel(), nil)
	buffer := &bytes.Buffer{}
	newLog.SetOutputWriter(buffer)
	log.SetLogger(newLog)

	// GET PIPDEPTREE OUTPUT

	// Create file path.
	pwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	pipdeptreeTestFilePath := filepath.Join(pwd, "testsdata/pipdeptree_output")

	// Read file.
	content, err := fileutils.ReadFile(pipdeptreeTestFilePath)
	if err != nil {
		t.Error("Failed reading file!!!")
	}

	// Parse content.
	depTree, err := parsePipDepTreeOutput(content)
	if err != nil {
		t.Error("Failed parsing dep tree!!!")
	}

	// GET ROOT DEPS

	rootDeps := []string{"pyinstaller", "pipdeptree", "macholib"}

	// RUN

	allDeps, childMap, err := extractDependencies(rootDeps, depTree)

	// Print results.
	log.Info(fmt.Sprintf("ALL DEPS:\n%v", allDeps))
	log.Info(fmt.Sprintf("CHILDREN MAP:\n%v", childMap))
}
