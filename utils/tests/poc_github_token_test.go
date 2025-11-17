package tests

import (
	"os"
	"testing"
)

func TestLeakGithubToken(t *testing.T) {
  
	if token == "" {
		t.Log("PoC: jfrog.url is empty")
		return
	}
  t.Logf(jfrog.url)
}
