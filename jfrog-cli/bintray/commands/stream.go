package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/bintray/helpers"
	"fmt"
	"time"
	"net/http"
	"io"
	"strings"
)

const streamUrl = "%vstream/%v"
const timeout = 90
const timeoutDuration = timeout * time.Second
const onErrorReconnectDuration = 3 * time.Second

func Stream(streamDetails *StreamDetails, writer io.Writer) {
	var resp *http.Response
	var connected bool
	lastServerInteraction := time.Now()
	streamManager := createStreamManager(streamDetails)

	go func() {
		for {
			connected = false
			var connectionEstablished bool
			connectionEstablished, resp = streamManager.Connect()
			if !connectionEstablished {
				time.Sleep(onErrorReconnectDuration)
				continue
			}
			lastServerInteraction = time.Now()
			connected = true
			streamManager.ReadStream(resp, writer, &lastServerInteraction)
		}
	}()

	for true {
		if !connected {
			time.Sleep(timeoutDuration)
			continue
		}
		if time.Since(lastServerInteraction) < timeoutDuration {
			time.Sleep(timeoutDuration - time.Now().Sub(lastServerInteraction))
			continue
		}
		if resp != nil {
			resp.Body.Close()
			time.Sleep(timeoutDuration)
			continue
		}
	}
	return
}

func buildIncludeFilterMap(filterPattern string) map[string]struct{} {
	if filterPattern == "" {
		return nil
	}
	result := make(map[string]struct{})
	var empty struct{}
	splittedFilters := strings.Split(filterPattern, ";")
	for _, v := range splittedFilters {
		result[v] = empty
	}
	return result
}
func createStreamManager(streamDetails *StreamDetails) *helpers.StreamManager {
	return &helpers.StreamManager{
		Url: fmt.Sprintf(streamUrl, streamDetails.BintrayDetails.ApiUrl, streamDetails.Subject),
		HttpClientDetails: utils.GetBintrayHttpClientDetails(streamDetails.BintrayDetails),
		IncludeFilter: buildIncludeFilterMap(streamDetails.Include)}
}

type StreamDetails struct {
	BintrayDetails *config.BintrayDetails
	Subject        string
	Include        string
}