package docker

import "testing"

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
		in       string
		expected string
	}{
		{"domain:8080/path:1.0", "domain:8080"},
		{"domain:8080/path/in/artifactory:1.0", "domain:8080/path"},
		{"domain:8080/path/in/artifactory", "domain:8080/path"},
		{"domain/path:1.0", "domain"},
		{"domain/path/in/artifactory:1.0", "domain/path"},
		{"domain/path/in/artifactory", "domain/path"},
	}

	for _, v := range imageTags {
		result, _ := ResolveRegistryFromTag(v.in)
		if result != v.expected {
			t.Errorf("Name(\"%s\") => '%s', want '%s'", v.in, result, v.expected)
		}
	}
}
