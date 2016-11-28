package main

import (
	"testing"
	"flag"
	"os"
	"github.com/jfrogdev/jfrog-cli-go/utils/tests"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
	"strings"
)

func TestMain(m *testing.M) {
	flag.Parse()

	if isArtifactoryTested() {
		err := tests.CreateReposIfNeeded()
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
	}

	if isBintrayTested() {
		//fmt.Println("########## bintray init")
	}

	result := m.Run()
	tests.InitTest()
	os.Exit(result)
}

func isArtifactoryTested() bool {
	return strings.Contains("Artifactory", flag.Lookup("test.run").Value.String()) || flag.Lookup("test.run").Value.String() == ""
}

func isBintrayTested() bool {
	return strings.Contains("Bintray", flag.Lookup("test.run").Value.String()) || flag.Lookup("test.run").Value.String() == ""
}
