package xray

import (
	"net/http"
	"io"
	"fmt"
	"github.com/buger/jsonparser"
	"io/ioutil"
	clienttests "github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/tests"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"os"
)

const (
	CleanScanBuildName  = "cleanBuildName"
	FatalScanBuildName  = "fatalBuildName"
	VulnerableBuildName = "vulnerableBuildName"
)

type flushWriter struct {
	f http.Flusher
	w io.Writer
}

func handler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buildName, err := jsonparser.GetString(body, "buildName")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch buildName {
	case CleanScanBuildName:
		fmt.Fprint(w, CleanXrayScanResponse)
		return
	case FatalScanBuildName:
		fmt.Fprint(w, FatalErrorXrayScanResponse)
		return
	case VulnerableBuildName:
		fmt.Fprint(w, VulnerableXrayScanResponse)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func StartXrayMockServer() int {
	handlers := clienttests.HttpServerHandlers{}
	handlers["/api/xray/scanBuild"] = handler
	handlers["/"] = http.NotFound

	port, err := clienttests.StartHttpServer(handlers)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	return port
}
