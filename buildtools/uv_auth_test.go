package buildtools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUvURLHasEmbeddedCredentials is a table-driven test for uvURLHasEmbeddedCredentials.
// The function returns true when the URL contains a userinfo component (user or user:pass).
func TestUvURLHasEmbeddedCredentials(t *testing.T) {
	cases := []struct {
		url      string
		expected bool
	}{
		{"https://user:pass@host.example.com/path", true},
		{"https://user@host.example.com/path", true}, // user present, no password still counts
		{"https://host.example.com/path", false},
		{"", false},
		{"not-a-url", false},
		{"https://host.example.com/api/pypi/repo/simple", false},
		{"http://user:secret@10.0.0.1:8080/simple", true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.url, func(t *testing.T) {
			assert.Equal(t, tc.expected, uvURLHasEmbeddedCredentials(tc.url),
				"uvURLHasEmbeddedCredentials(%q)", tc.url)
		})
	}
}

// TestUvNetrcHasCredentials writes a temp .netrc file and verifies that
// uvNetrcHasCredentials correctly matches the hostname.
func TestUvNetrcHasCredentials(t *testing.T) {
	tmpDir := t.TempDir()
	netrcPath := filepath.Join(tmpDir, ".netrc")
	err := os.WriteFile(netrcPath, []byte(
		"machine myhost.example.com\nlogin user\npassword pass\n",
	), 0600)
	assert.NoError(t, err)

	t.Setenv("HOME", tmpDir)

	// matching host — should return true
	assert.True(t, uvNetrcHasCredentials("https://myhost.example.com/simple"),
		"should match machine entry in .netrc")

	// non-matching host — should return false
	assert.False(t, uvNetrcHasCredentials("https://other.example.com/simple"),
		"should not match when host is absent from .netrc")

	// empty URL — should return false
	assert.False(t, uvNetrcHasCredentials(""),
		"empty URL should return false")

	// URL with no host component — should return false
	assert.False(t, uvNetrcHasCredentials("not-a-url"),
		"unparseable/hostless URL should return false")
}

// TestUvNetrcHasCredentials_CustomPath verifies that the NETRC env var is
// respected for a custom netrc file location (same as UV and curl behavior).
func TestUvNetrcHasCredentials_CustomPath(t *testing.T) {
	tmpDir := t.TempDir()
	customPath := filepath.Join(tmpDir, "custom_netrc")
	err := os.WriteFile(customPath, []byte("machine customhost.example.com\nlogin u\npassword p\n"), 0600)
	assert.NoError(t, err)

	// Point $NETRC at the custom file, HOME at an empty dir (no ~/.netrc)
	t.Setenv("NETRC", customPath)
	t.Setenv("HOME", t.TempDir())

	assert.True(t, uvNetrcHasCredentials("https://customhost.example.com/simple"),
		"should find credentials via $NETRC custom path")
	assert.False(t, uvNetrcHasCredentials("https://other.example.com/simple"),
		"non-matching host should return false even with $NETRC set")
}

// TestUvNetrcHasCredentials_NoFile verifies that when ~/.netrc does not exist,
// the function returns false rather than erroring.
func TestUvNetrcHasCredentials_NoFile(t *testing.T) {
	// Point HOME at an empty dir so there is no .netrc
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	assert.False(t, uvNetrcHasCredentials("https://myhost.example.com/simple"),
		"should return false when .netrc file does not exist")
}

// TestUvNetrcHasCredentials_MultipleEntries verifies correct entry selection
// when .netrc contains multiple machine entries.
func TestUvNetrcHasCredentials_MultipleEntries(t *testing.T) {
	tmpDir := t.TempDir()
	netrcContent := "machine first.example.com\nlogin u1\npassword p1\n" +
		"machine second.example.com\nlogin u2\npassword p2\n"
	err := os.WriteFile(filepath.Join(tmpDir, ".netrc"), []byte(netrcContent), 0600)
	assert.NoError(t, err)
	t.Setenv("HOME", tmpDir)

	assert.True(t, uvNetrcHasCredentials("https://first.example.com/path"))
	assert.True(t, uvNetrcHasCredentials("https://second.example.com/path"))
	assert.False(t, uvNetrcHasCredentials("https://third.example.com/path"))
}

// TestUvIndexEnvName is a table-driven test for uvIndexEnvName.
// UV uppercases the name and replaces hyphens, dots, and spaces with underscores.
func TestUvIndexEnvName(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"pypi-virtual", "PYPI_VIRTUAL"},
		{"my.index", "MY_INDEX"},
		{"MyIndex", "MYINDEX"},
		{"index-with-multiple-hyphens", "INDEX_WITH_MULTIPLE_HYPHENS"},
		{"index.with.dots", "INDEX_WITH_DOTS"},
		{"index name", "INDEX_NAME"},
		{"jfrog-pypi-virtual", "JFROG_PYPI_VIRTUAL"},
		{"ALREADY_UPPER", "ALREADY_UPPER"},
		{"mixed.case-name", "MIXED_CASE_NAME"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.expected, uvIndexEnvName(tc.input),
				"uvIndexEnvName(%q)", tc.input)
		})
	}
}

// TestUvIndexHasNativeCredentials verifies the composite logic of
// uvIndexHasNativeCredentials: env var takes priority over URL credentials,
// which take priority over .netrc.
func TestUvIndexHasNativeCredentials(t *testing.T) {
	t.Run("env var set", func(t *testing.T) {
		t.Setenv("UV_INDEX_MY_INDEX_USERNAME", "testuser")
		assert.True(t, uvIndexHasNativeCredentials(
			"https://host.example.com/simple",
			"UV_INDEX_MY_INDEX_USERNAME",
		))
	})

	t.Run("env var not set, URL has credentials", func(t *testing.T) {
		// ensure env var is not set (t.Setenv restores on cleanup)
		assert.True(t, uvIndexHasNativeCredentials(
			"https://user:pass@host.example.com/simple",
			"UV_INDEX_NONEXISTENT_VAR_XYZ",
		))
	})

	t.Run("no env var, no URL credentials, netrc match", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := os.WriteFile(filepath.Join(tmpDir, ".netrc"),
			[]byte("machine netrchost.example.com\nlogin u\npassword p\n"), 0600)
		assert.NoError(t, err)
		t.Setenv("HOME", tmpDir)
		assert.True(t, uvIndexHasNativeCredentials(
			"https://netrchost.example.com/simple",
			"UV_INDEX_NONEXISTENT_VAR_XYZ2",
		))
	})

	t.Run("none of the above", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir) // no .netrc
		assert.False(t, uvIndexHasNativeCredentials(
			"https://host.example.com/simple",
			"UV_INDEX_NONEXISTENT_VAR_XYZ3",
		))
	})
}
