package project

import "testing"

func TestParseModuleName(t *testing.T) {
	content := `

module github.com/jfrog/go-example

require (
        github.com/Sirupsen/logrus v1.0.6
        golang.org/x/crypto v0.0.0-20180802221240-56440b844dfe // indirect
        golang.org/x/sys v0.0.0-20180802203216-0ffbfd41fbef // indirect
        rsc.io/quote v1.5.2
)
	`

	expected := "github.com/jfrog/go-example"
	actual, err := parseModuleName(content)

	if err != nil {
		t.Error(err)
	}

	if actual != expected {
		t.Errorf("Expected: %s, got: %s.", expected, actual)
	}
}

func TestParseModuleNameWithoutSlash(t *testing.T) {
	content := `

module go4.org

require (
        github.com/Sirupsen/logrus v1.0.6
        golang.org/x/crypto v0.0.0-20180802221240-56440b844dfe // indirect
        golang.org/x/sys v0.0.0-20180802203216-0ffbfd41fbef // indirect
        rsc.io/quote v1.5.2
)
	`

	expected := "go4.org"
	actual, err := parseModuleName(content)

	if err != nil {
		t.Error(err)
	}

	if actual != expected {
		t.Errorf("Expected: %s, got: %s.", expected, actual)
	}
}
