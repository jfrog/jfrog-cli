package main

import (
	"testing"
	"flag"
	"os"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/tests"
)

func TestMain(m *testing.M) {
	flag.Parse()

	if *tests.TestBintray {
		InitBintrayTests()
	}
	if *tests.TestArtifactory && !*tests.TestArtifactoryProxy {
		InitArtifactoryTests()
	}

	result := m.Run()

	if *tests.TestBintray {
		CleanBintrayTests()
	}
	if *tests.TestArtifactory && !*tests.TestArtifactoryProxy {
		CleanArtifactoryTests()
	}

	os.Exit(result)
}
