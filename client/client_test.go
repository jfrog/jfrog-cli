package client

import (
    "testing"
    "fmt"
    "encoding/json"
)

func TestInit(t *testing.T) {
    const dummyUrl = "http://dummy"
    bintray := New("user", "apkiKey", dummyUrl)
    if bintray.ApiUrl != dummyUrl {
        t.Fatalf("unexpected value expected %s:, received: %s", dummyUrl, bintray.ApiUrl)
    }
}

func TestSearchRepos(t *testing.T) {
    bintray := New("user", "apkiKey", "", nil)
    repos := bintray.GetRepositories("yoavl")
    buf, err := json.MarshalIndent(repos, "", "  ")
    perror(err)
    fmt.Printf("%s\n", buf)
}

