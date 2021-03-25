package commands

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConvertBuildCmd(t *testing.T) {
	tests := []buildCmd{
		{"simpleMvn", "mvn clean install", "jfrog rt mvn clean install"},
		{"simpleGradle", "gradle clean build", "jfrog rt gradle clean build"},
		{"simpleNpm", "npm restore", "jfrog rt npm restore"},
		{"complex", "mvn clean install && gradle clean build", "jfrog rt mvn clean install && jfrog rt gradle clean build"},
		{"hiddenNpm", "mvn clean install -f \"hiddennpm/pom.xml\" && gradle clean build", "jfrog rt mvn clean install -f \"hiddennpm/pom.xml\" && jfrog rt gradle clean build"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			converted, err := convertBuildCmd(test.original)
			if err != nil {
				assert.NoError(t, err)
				return
			}
			assert.Equal(t, test.expected, converted)
		})
	}
}

type buildCmd struct {
	name     string
	original string
	expected string
}
