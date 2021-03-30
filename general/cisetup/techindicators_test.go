package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTechIndicator(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected Technology
	}{
		{"simpleMavenTest", "pom.xml", "maven"},
		{"npmTest", "../package.json", "npm"},
		{"windowsGradleTest", "c://users/test/package/build.gradle", "gradle"},
		{"noTechTest", "pomxml", ""},
	}
	indicators := GetTechIndicators()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var detectedTech Technology
			for _, indicator := range indicators {
				if indicator.Indicates(test.filePath) {
					detectedTech = indicator.GetTechnology()
					break
				}
			}
			assert.Equal(t, test.expected, detectedTech)
		})
	}
}

type techTest struct {
}
