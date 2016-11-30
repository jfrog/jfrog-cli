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

const BINTRAY_RECONNECT_HEADER = "X-Bintray-Hose-Reconnect-Id"

type FirehoseManager struct {
	HttpClientDetails ioutils.HttpClientDetails
	Url               string
	IncludeFilter     map[string]struct{}
	ReconnectId       string
}

func (fh *FirehoseManager) ReadStream(resp *http.Response, writer io.Writer, lastServerInteraction *time.Time) {
	ioReader := resp.Body
	bodyReader := bufio.NewReader(ioReader)
	fh.handleStream(bodyReader, writer, lastServerInteraction)
}

func (fh *FirehoseManager) handleStream(ioReader io.Reader, writer io.Writer, lastServerInteraction *time.Time) {
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
	fh.parseStream(streamDecoder, streamEncoder)
}

func (fh *FirehoseManager) parseStream(streamDecoder *json.Decoder, streamEncoder *json.Encoder) error {
	for {
		var decodedJson map[string]interface{}
		if e := streamDecoder.Decode(&decodedJson); e != nil {
			return e
		}
		if _, ok := fh.IncludeFilter[decodedJson["type"].(string)]; ok || len(fh.IncludeFilter) == 0 {
			if e := streamEncoder.Encode(&decodedJson); e != nil {
				return e
			}
		}
	}
}

func (fh *FirehoseManager) isReconnection() bool {
	return len(fh.ReconnectId) > 0
}

func (fh *FirehoseManager) setReconnectHeader() {
	if fh.HttpClientDetails.Headers == nil {
		fh.HttpClientDetails.Headers = map[string]string{}
	}
	fh.HttpClientDetails.Headers[BINTRAY_RECONNECT_HEADER] = fh.ReconnectId
}

func (fh *FirehoseManager) Connect() (bool, *http.Response) {
	if fh.isReconnection() {
		fh.setReconnectHeader()
	}
	logger.Logger.Info("Connecting to firehose...")
	resp, _, _, e := ioutils.Stream(fh.Url, fh.HttpClientDetails)
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
	fh.ReconnectId = resp.Header.Get(BINTRAY_RECONNECT_HEADER)
	logger.Logger.Info("Connected successfully...")
	return true, resp
}
