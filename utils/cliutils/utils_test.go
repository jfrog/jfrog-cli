package cliutils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSplitAgentNameAndVersion(t *testing.T) {
	tests := []struct {
		fullAgentName        string
		expectedAgentName    string
		expectedAgentVersion string
	}{
		{"abc/1.2.3", "abc", "1.2.3"},
		{"abc/def/1.2.3", "abc/def", "1.2.3"},
		{"abc\\1.2.3", "abc\\1.2.3", ""},
		{"abc:1.2.3", "abc:1.2.3", ""},
		{"", "", ""},
	}

	for _, test := range tests {
		actualAgentName, actualAgentVersion := splitAgentNameAndVersion(test.fullAgentName)
		assert.Equal(t, test.expectedAgentName, actualAgentName)
		assert.Equal(t, test.expectedAgentVersion, actualAgentVersion)
	}
}
