package generic

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPingSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	}))
	defer ts.Close()
	responseBytes, err := new(PingCommand).SetRtDetails(&config.ArtifactoryDetails{Url: ts.URL + "/"}).Ping()
	if err != nil {
		t.Log(fmt.Sprintf("Error received from Artifactory following ping request: %s", err))
		t.Fail()
	}
	responseString := string(responseBytes)
	if responseString != "OK" {
		t.Log(fmt.Sprintf("Non 'OK' response received from Artifactory following ping request:: %s", responseString))
		t.Fail()
	}
}

func TestPingFailed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, `{"error":"error"}`)
	}))
	defer ts.Close()
	_, err := new(PingCommand).SetRtDetails(&config.ArtifactoryDetails{Url: ts.URL + "/"}).Ping()
	if err == nil {
		t.Log("Expected error from artifactory")
		t.Fail()
	}
}
