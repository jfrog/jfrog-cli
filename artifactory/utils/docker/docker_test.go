package docker

import (
	"testing"
)

func TestGetImagePath(t *testing.T) {
	var imageTags = []struct {
		in       string
		expected string
	}{
		{"domain:8080/path:1.0", "/path/1.0"},
		{"domain:8080/path/in/artifactory:1.0", "/path/in/artifactory/1.0"},
		{"domain:8080/path/in/artifactory", "/path/in/artifactory/latest"},
		{"domain/path:1.0", "/path/1.0"},
		{"domain/path/in/artifactory:1.0", "/path/in/artifactory/1.0"},
		{"domain/path/in/artifactory", "/path/in/artifactory/latest"},
	}

	for _, v := range imageTags {
		result := New(v.in).Path()
		if result != v.expected {
			t.Errorf("Path(\"%s\") => '%s', want '%s'", v.in, result, v.expected)
		}
	}
}

func TestGetImageName(t *testing.T) {
	var imageTags = []struct {
		in       string
		expected string
	}{
		{"domain:8080/path:1.0", "path:1.0"},
		{"domain:8080/path/in/artifactory:1.0", "artifactory:1.0"},
		{"domain:8080/path/in/artifactory", "artifactory:latest"},
		{"domain/path:1.0", "path:1.0"},
		{"domain/path/in/artifactory:1.0", "artifactory:1.0"},
		{"domain/path/in/artifactory", "artifactory:latest"},
	}

	for _, v := range imageTags {
		result := New(v.in).Name()
		if result != v.expected {
			t.Errorf("Name(\"%s\") => '%s', want '%s'", v.in, result, v.expected)
		}
	}
}

func TestResolveRegistryFromTag(t *testing.T) {
	var imageTags = []struct {
		in             string
		expected       string
		expectingError bool
	}{
		{"domain:8080/path:1.0", "domain:8080", false},
		{"domain:8080/path/in/artifactory:1.0", "domain:8080/path", false},
		{"domain:8080/path/in/artifactory", "domain:8080/path", false},
		{"domain/path:1.0", "domain", false},
		{"domain/path/in/artifactory:1.0", "domain/path", false},
		{"domain/path/in/artifactory", "domain/path", false},
		{"domain:8081", "", true},
	}

	for _, v := range imageTags {
		result, err := ResolveRegistryFromTag(v.in)
		if err != nil && !v.expectingError {
			t.Error(err.Error())
		}
		if result != v.expected {
			t.Errorf("ResolveRegistryFromTag(\"%s\") => '%s', expected '%s'", v.in, result, v.expected)
		}
	}
}

func TestDockerClientApiVersionRegex(t *testing.T) {
	var versionStrings = []struct {
		in       string
		expected bool
	}{
		{"1", false},
		{"1.1", true},
		{"1.11", true},
		{"12.12", true},
		{"1.1.11", false},
		{"1.illegal", false},
		{"1 11", false},
	}

	for _, v := range versionStrings {
		result := ApiVersionRegex.Match([]byte(v.in))
		if result != v.expected {
			t.Errorf("Version(\"%s\") => '%v', want '%v'", v.in, result, v.expected)
		}
	}
}
