package fileutils

import (
	"strconv"
	"testing"
)

func TestIsSsh(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"http://some.url", false},
		{"https://some.url", false},
		{"sshd://wrong.url", false},
		{"assh://wrong.url", false},
		{"ssh://some.url", true},
		{"sSh://some.url/some/api", true},
		{"SSH://some.url/some/api", true},
	}
	for _, test := range tests {
		t.Run(test.url, func(t *testing.T) {
			if IsSshUrl(test.url) != test.expected {
				t.Error("Expected '"+strconv.FormatBool(test.expected)+"' Got: '"+strconv.FormatBool(!test.expected)+"' For URL:", test.url)
			}
		})
	}
}
