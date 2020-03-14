package project

import (
	"testing"
)

func TestParseModuleName(t *testing.T) {
	expected := "github.com/jfrog/go-example"
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{"moduleNameWithoutQuotationMarks", `

module github.com/jfrog/go-example

go 1.14

require (
        github.com/Sirupsen/logrus v1.0.6
        golang.org/x/crypto v0.0.0-20180802221240-56440b844dfe // indirect
        golang.org/x/sys v0.0.0-20180802203216-0ffbfd41fbef // indirect
        rsc.io/quote v1.5.2
)
	`, expected},
		{"ModuleNameWithQuotationMarks", `

module "github.com/jfrog/go-example"

require (
        github.com/Sirupsen/logrus v1.0.6
        golang.org/x/crypto v0.0.0-20180802221240-56440b844dfe // indirect
        golang.org/x/sys v0.0.0-20180802203216-0ffbfd41fbef // indirect
        rsc.io/quote v1.5.2
)
	`, expected},
		{"ModuleNameWithoutSlash", `

module go4.org

require (
        github.com/Sirupsen/logrus v1.0.6
        golang.org/x/crypto v0.0.0-20180802221240-56440b844dfe // indirect
        golang.org/x/sys v0.0.0-20180802203216-0ffbfd41fbef // indirect
        rsc.io/quote v1.5.2
)
	`, "go4.org"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := parseModuleName(test.content)

			if err != nil {
				t.Error(err)
			}

			if actual != test.expected {
				t.Errorf("Expected: %s, got: %s.", test.expected, actual)
			}
		})
	}
}
