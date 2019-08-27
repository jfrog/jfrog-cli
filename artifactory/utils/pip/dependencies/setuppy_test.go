package dependencies

import (
	"bytes"
	logUtils "github.com/jfrog/jfrog-cli-go/utils/log"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"os"
	"testing"
)

func initTest() *setupExtractor {
	// Create log.
	newLog := log.NewLogger(logUtils.GetCliLogLevel(), nil)
	buffer := &bytes.Buffer{}
	newLog.SetOutputWriter(buffer)
	log.SetLogger(newLog)

	// Create extractor.
	newExtractor := &setupExtractor{setuppyFilePath: "/Users/barb/trash/devops-tools/setup.py", pythonExecutablePath: "/Users/barb/trash/venv-test2/bin/python"}

	return newExtractor
}

func TestExecuteEgginfoCommandWithOutput(t *testing.T) {
	pathVar := os.Getenv("PATH")
	os.Setenv("PATH", "/Users/barb/trash/venv-test2/bin:/usr/local/opt/node@8/bin")
	defer os.Setenv("PATH", pathVar)

	extractor := initTest()

	// Create temp dir.
	tempDirPath, err := fileutils.CreateTempDir()
	if err != nil {
		log.Info("Failed creating temp dir!")
		t.Error(err)
	}
	defer fileutils.RemoveTempDir(tempDirPath)
	err = os.Chdir(tempDirPath)
	if errorutils.CheckError(err) != nil {
		log.Info("Failed changing dir to temp!")
		t.Error(err)
	}

	// Get project name.
	pkgName, err := extractor.PackageName()
	if err != nil {
		log.Info("Error in execution: " + err.Error())
		t.Error(err)
	}

	log.Info("Project name: " + pkgName)

	// Get dep tree
	extractor.Pkg = "django-rss-plugin"
	depMap, err := BuildPipDependencyMap("/Users/barb/trash/venv-test2/bin/python", nil)
	if err != nil {
		t.Error("Failed BuildPipDependencyMap!")
	}
	extractor.extractRootDependencies(depMap)
}
