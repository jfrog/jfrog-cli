package generic

import (
	"testing"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"net/http/httptest"
	"net/http"
	"fmt"
)

func TestPingSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	}))
	defer ts.Close()
	responseBytes,err := Ping(&config.ArtifactoryDetails{Url: ts.URL + "/"})
	if err != nil {
		t.Log("Error from artifactory %s", err)
		t.Fail()
	}
	responseString := string(responseBytes)
	if responseString != "OK" {
		t.Log("bad response string from artifactory: %s", responseString)
		t.Fail()
	}
}


func TestPingFailed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, `{"error":"error"}`)
	}))
	defer ts.Close()
	_,err := Ping(&config.ArtifactoryDetails{Url: ts.URL + "/"})
	if err == nil {
		t.Log("Expected error from artifactory")
		t.Fail()
	}
}
