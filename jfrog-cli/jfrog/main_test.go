package main

import (
	"testing"
	"flag"
	"os"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/tests"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !*tests.TestArtifactoryProxy {
		InitBintrayTests()
		InitArtifactoryTests()
	}

	result := m.Run()

	if !*tests.TestArtifactoryProxy {
		CleanBintrayTests()
		CleanArtifactoryTests()
	}
	os.Exit(result)
}
