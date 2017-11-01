package main

import (
	"testing"
	"flag"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/tests"
	"os/exec"
	"strings"
	"bufio"
	"regexp"
	"os"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
)

const (
	VENDOR_TESTS = "vendor"
	INTEGRATION_TESTS = "jfrog-cli-go/jfrog-cli/jfrog"
	DOCS_TEST = "jfrog-cli-go/jfrog-cli/docs"
)

func TestMain(m *testing.M) {
	runUnitTests()
	setupIntegrationTests()
	result := m.Run()
	tearDownIntegrationTests()
	os.Exit(result)
}

func runUnitTests() {
	unitTests := append([]string{"test"}, listUnitTests()...)
	cmd := exec.Command("go", unitTests...)
	res, err := cmd.Output()
	if err != nil {
		log.Error(err)
	}
	log.Info(string(res))
	if err != nil || strings.Contains(string(res), "FAIL") {
		os.Exit(1)
	}
}

func listUnitTests() []string {
	cmd := exec.Command("go", "list", "../../...")
	res, _ := cmd.Output()
	scanner := bufio.NewScanner(strings.NewReader(string(res)))
	var unitTests []string
	for scanner.Scan() {
		excludedTest, _ := regexp.MatchString(VENDOR_TESTS + "|" + INTEGRATION_TESTS + "|" + DOCS_TEST, scanner.Text())
		if !excludedTest {
			unitTests = append(unitTests, scanner.Text())
		}
	}
	return unitTests
}

func setupIntegrationTests() {
	flag.Parse()
	if *tests.TestBintray {
		InitBintrayTests()
	}
	if *tests.TestArtifactory && !*tests.TestArtifactoryProxy {
		InitArtifactoryTests()
	}
	if *tests.TestBuildTools {
		InitBuildToolsTests()
	}
}

func tearDownIntegrationTests() {
	if *tests.TestBintray {
		CleanBintrayTests()
	}
	if *tests.TestArtifactory && !*tests.TestArtifactoryProxy {
		CleanArtifactoryTests()
	}
	if *tests.TestBuildTools {
		CleanBuildToolsTests()
	}
}
