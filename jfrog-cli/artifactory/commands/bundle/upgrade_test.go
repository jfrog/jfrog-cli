package bundle

import (
	"github.com/jfrog/jfrog-client-go/artifactory/services/bundle"
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestGetLatestVersionCompleted(t *testing.T) {
	versions := []bundle.Version{
		{"1.0.0", "", bundle.Complete},
	}
	assert.Equal(t, getLatestMatchedVersion("*", versions), "1.0.0")
	assert.Equal(t, getLatestMatchedVersion("1", versions), "1.0.0")
	assert.Equal(t, getLatestMatchedVersion("1.*", versions), "1.0.0")
	assert.Equal(t, getLatestMatchedVersion("1.1.*", versions), "")
	assert.Equal(t, getLatestMatchedVersion("0.5", versions), "")
}

func TestGetLatestVersionNotCompleted(t *testing.T) {
	versions := []bundle.Version{
		{"1.0.0", "", bundle.Failed},
	}
	assert.Equal(t, getLatestMatchedVersion("*", versions), "")
	assert.Equal(t, getLatestMatchedVersion("1", versions), "")
	assert.Equal(t, getLatestMatchedVersion("1.*", versions), "")
	assert.Equal(t, getLatestMatchedVersion("1.1.*", versions), "")
	assert.Equal(t, getLatestMatchedVersion("0.5", versions), "")
}

func TestGetLatestVersionEmpty(t *testing.T) {
	var versions []bundle.Version
	assert.Equal(t, getLatestMatchedVersion("*", versions), "")
	assert.Equal(t, getLatestMatchedVersion("1", versions), "")
	assert.Equal(t, getLatestMatchedVersion("1.*", versions), "")
	assert.Equal(t, getLatestMatchedVersion("1.1.*", versions), "")
	assert.Equal(t, getLatestMatchedVersion("0.5", versions), "")
}

func TestGetLatestVersionMultiple(t *testing.T) {
	versions := []bundle.Version{
		{"1.0.0", "", bundle.Complete},
		{"1.0.1", "", bundle.Complete},
		{"1.0.1", "", bundle.Failed},
		{"1.0.2", "", bundle.NotDistributed},
		{"1.0.x-SNAPSHOT", "", bundle.InProgress},
	}
	assert.Equal(t, getLatestMatchedVersion("*", versions), "1.0.1")
	assert.Equal(t, getLatestMatchedVersion("1", versions), "1.0.0")
	assert.Equal(t, getLatestMatchedVersion("1.*", versions), "1.0.1")
	assert.Equal(t, getLatestMatchedVersion("1.1.*", versions), "")
	assert.Equal(t, getLatestMatchedVersion("0.5", versions), "")
}

func TestGetLatestVersionSnapshot(t *testing.T) {
	versions := []bundle.Version{
		{"1.0.x-SNAPSHOT", "", bundle.Complete},
	}
	assert.Equal(t, getLatestMatchedVersion("*", versions), "1.0.x-SNAPSHOT")
	assert.Equal(t, getLatestMatchedVersion("1", versions), "")
	assert.Equal(t, getLatestMatchedVersion("1.*", versions), "1.0.x-SNAPSHOT")
	assert.Equal(t, getLatestMatchedVersion("1.0.*", versions), "1.0.x-SNAPSHOT")
	assert.Equal(t, getLatestMatchedVersion("0.5", versions), "")
	assert.Equal(t, getLatestMatchedVersion("0.5.*", versions), "")
	assert.Equal(t, getLatestMatchedVersion("0.*", versions), "")
	assert.Equal(t, getLatestMatchedVersion("2", versions), "")
	assert.Equal(t, getLatestMatchedVersion("2.*", versions), "")
	assert.Equal(t, getLatestMatchedVersion("2.0.*", versions), "")
}
