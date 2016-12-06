package helpers

import (
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/logger"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"time"
	"net/http"
	"io"
	"io/ioutil"
	"errors"
	"bufio"
	"encoding/json"
)

const BINTRAY_RECONNECT_HEADER = "X-Bintray-Stream-Reconnect-Id"

type StreamManager struct {
	HttpClientDetails ioutils.HttpClientDetails
	Url               string
	IncludeFilter     map[string]struct{}
	ReconnectId       string
}

func (sm *StreamManager) ReadStream(resp *http.Response, writer io.Writer, lastServerInteraction *time.Time) {
	ioReader := resp.Body
	bodyReader := bufio.NewReader(ioReader)
	sm.handleStream(bodyReader, writer, lastServerInteraction)
}

func (sm *StreamManager) handleStream(ioReader io.Reader, writer io.Writer, lastServerInteraction *time.Time) {
	bodyReader := bufio.NewReader(ioReader)
	pReader, pWriter := io.Pipe()
	defer pWriter.Close()
	go func() {
		defer pReader.Close()
		for {
			line, _, err := bodyReader.ReadLine()
			if err != nil {
				break
			}
			*lastServerInteraction = time.Now()
			_, err = pWriter.Write(line)
			if err != nil {
				break
			}
		}
	}()
	streamDecoder := json.NewDecoder(pReader)
	streamEncoder := json.NewEncoder(writer)
	sm.parseStream(streamDecoder, streamEncoder)
}

func (sm *StreamManager) parseStream(streamDecoder *json.Decoder, streamEncoder *json.Encoder) error {
	for {
		var decodedJson map[string]interface{}
		if e := streamDecoder.Decode(&decodedJson); e != nil {
			return e
		}
		if _, ok := sm.IncludeFilter[decodedJson["type"].(string)]; ok || len(sm.IncludeFilter) == 0 {
			if e := streamEncoder.Encode(&decodedJson); e != nil {
				return e
			}
		}
	}
}

func (sm *StreamManager) isReconnection() bool {
	return len(sm.ReconnectId) > 0
}

func (sm *StreamManager) setReconnectHeader() {
	if sm.HttpClientDetails.Headers == nil {
		sm.HttpClientDetails.Headers = map[string]string{}
	}
	sm.HttpClientDetails.Headers[BINTRAY_RECONNECT_HEADER] = sm.ReconnectId
}

func (sm *StreamManager) Connect() (bool, *http.Response) {
	if sm.isReconnection() {
		sm.setReconnectHeader()
	}
	logger.Logger.Info("Connecting...")
	resp, _, _, e := ioutils.Stream(sm.Url, sm.HttpClientDetails)
	if e != nil {
		return false, resp
	}
	if resp.StatusCode != 200 {
		cliutils.CheckError(errors.New("response: " + resp.Status))
		msgBody, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode > 400 && resp.StatusCode < 500 {
			cliutils.Exit(cliutils.ExitCodeError, string(msgBody))
		}
		return false, resp

	}
	sm.ReconnectId = resp.Header.Get(BINTRAY_RECONNECT_HEADER)
	logger.Logger.Info("Connected.")
	return true, resp
}
