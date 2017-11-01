package httputils

import (
	"testing"
	"fmt"
	"time"
	"io"
	"net/http"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"errors"
	"strconv"
	"io/ioutil"
	"os"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/tests"
)

const SUCCESS_RESPONSE = "successful response"

type testContext struct {
	tryNum int
}

var mockServerPort int

func TestMain(m *testing.M) {
	mockServerPort = startMockServer()
	result := m.Run()
	os.Exit(result)
}

func TestSimpleSuccessful(t *testing.T) {
	t.Parallel()
	port := mockServerPort
	ctx := &testContext{}

	connection := RetryableConnection{
		ReadTimeout:            time.Second * 3,
		RetriesNum:             3,
		StableConnectionWindow: time.Second * 4,
		SleepBetweenRetries:    time.Second * 1,
		ConnectHandler: func() (*http.Response, error) {
			return execGet(port, "/simple", ctx)
		},
		ErrorHandler: func(content []byte) error {
			return nil
		},
	}

	res, err := connection.Do()
	if err != nil {
		t.Error(err)
		return
	}

	if string(res) != SUCCESS_RESPONSE {
		t.Error(fmt.Errorf("expected, %s, got: %s", SUCCESS_RESPONSE, string(res)))
	}
}

func TestSimpleExceedConnectionRetries(t *testing.T) {
	t.Parallel()
	port := mockServerPort
	c := &testContext{}

	connection := RetryableConnection{
		ReadTimeout:            time.Second * 3,
		RetriesNum:             3,
		StableConnectionWindow: time.Second * 4,
		SleepBetweenRetries:    time.Second * 1,
		ConnectHandler: func() (*http.Response, error) {
			return execGet(port, "/exceed/retries", c)
		},
		ErrorHandler: func(content []byte) error {
			return nil
		},
	}

	_, err := connection.Do()
	if err != exhaustedErr {
		t.Error(err)
		return
	}
}

func TestRetryStableWindowConnection(t *testing.T) {
	t.Parallel()
	port := mockServerPort
	c := &testContext{}

	connection := RetryableConnection{
		ReadTimeout:            time.Second * 3,
		RetriesNum:             3,
		StableConnectionWindow: time.Second * 8,
		SleepBetweenRetries:    time.Second * 1,
		ConnectHandler: func() (*http.Response, error) {
			return execGet(port, "/window", c)
		},
		ErrorHandler: func(content []byte) error {
			return nil
		},
	}

	res, err := connection.Do()
	if err != nil {
		t.Error(err)
		return
	}

	if string(res) != SUCCESS_RESPONSE {
		t.Error(fmt.Errorf("expected, %s, got: %s", SUCCESS_RESPONSE, string(res)))
	}
}

// Testing for stable connection retries.
// Each retry context will be updated, so windowHandler will execute different behaviour.
func TestRetryExceedUnstableWindowConnection(t *testing.T) {
	t.Parallel()
	port := mockServerPort
	c := &testContext{}

	connection := RetryableConnection{
		ReadTimeout:            time.Second * 3,
		RetriesNum:             3,
		StableConnectionWindow: time.Second * 9,
		SleepBetweenRetries:    time.Second * 1,
		ConnectHandler: func() (*http.Response, error) {
			return execGet(port, "/window", c)
		},
		ErrorHandler: func(content []byte) error {
			return nil
		},
	}

	_, err := connection.Do()
	if err != exhaustedErr {
		t.Error(err)
		return
	}
}

func TestRetryExceededWithNoStableConnectionWindow(t *testing.T) {
	t.Parallel()
	port := mockServerPort
	c := &testContext{}

	connection := RetryableConnection{
		ReadTimeout:         time.Second * 3,
		RetriesNum:          3,
		SleepBetweenRetries: time.Second * 1,
		ConnectHandler: func() (*http.Response, error) {
			return execGet(port, "/window", c)
		},
		ErrorHandler: func(content []byte) error {
			return nil
		},
	}

	_, err := connection.Do()
	if err != exhaustedErr {
		t.Error(err)
		return
	}
}

func TestErrorHandler(t *testing.T) {
	t.Parallel()
	port := mockServerPort
	c := &testContext{}

	retErr := errors.New("error to return")
	connection := RetryableConnection{
		ReadTimeout:            time.Second * 3,
		RetriesNum:             3,
		StableConnectionWindow: time.Second * 9,
		SleepBetweenRetries:    time.Second * 1,
		ConnectHandler: func() (*http.Response, error) {
			return execGet(port, "/simple", c)
		},
		ErrorHandler: func(content []byte) error {
			return retErr
		},
	}

	_, err := connection.Do()
	if err != exhaustedErr {
		t.Error(err)
		return
	}
}

// Send post request with context value in the request body.
func execGet(port int, path string, c *testContext) (*http.Response, error) {
	resp, _, _, err := Send("POST", "http://localhost:"+strconv.Itoa(port)+path,
		[]byte(strconv.Itoa(c.tryNum)), true, false, HttpClientDetails{})
	if err != nil {
		return resp, err
	}
	c.tryNum++

	if resp.StatusCode != 200 {
		err = errorutils.CheckError(errors.New("Response: " + resp.Status))
	}

	return resp, err
}

type flushWriter struct {
	f http.Flusher
	w io.Writer
}

func (fw *flushWriter) Write(p []byte) (n int, err error) {
	n, err = fw.w.Write(p)
	if fw.f != nil {
		fw.f.Flush()
	}
	return
}

// Simple handler will send \r\n 4 times with 1 sec in between
func simpleHandler(w http.ResponseWriter, r *http.Request) {
	fw := &flushWriter{w: w}
	if f, ok := w.(http.Flusher); ok {
		fw.f = f
	}

	sendIdleAndSleep(fw, 4, 1)
	fmt.Fprint(fw, SUCCESS_RESPONSE)
}

// Retry handler will send \r\n 4 times with 4 sec in between.
func exceedRetriesHandler(w http.ResponseWriter, r *http.Request) {
	fw := &flushWriter{w: w}
	if f, ok := w.(http.Flusher); ok {
		fw.f = f
	}

	sendIdleAndSleep(fw, 4, 4)
	fmt.Fprint(fw, SUCCESS_RESPONSE)
}

// Response handler with context according to the request body.
// For example:
// Sending body with 0 will send \r\n once with 10 secs sleep.
// Sending body with 1 will send NotFound response.
func windowHandler(w http.ResponseWriter, r *http.Request) {
	fw := &flushWriter{w: w}
	if f, ok := w.(http.Flusher); ok {
		fw.f = f
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	iteration, err := strconv.Atoi(string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch iteration {
	case 0:
		sendIdleAndSleep(fw, 1, 10)
		return
	case 1, 2:
		http.NotFound(w, r)
		return
	case 3:
		sendIdleAndSleep(fw, 4, 2)
		sendIdleAndSleep(fw, 1, 4)
		return
	case 4, 5:
		http.NotFound(w, r)
		return
	case 6:
		fmt.Fprint(fw, SUCCESS_RESPONSE)
		return
	}
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func sendIdleAndSleep(fw *flushWriter, iterations, sec int) {
	for i := 0; i < iterations; i++ {
		fmt.Fprint(fw, "\r\n")
		time.Sleep(time.Second * time.Duration(sec))
	}
}

func startMockServer() int {
	handlers := tests.HttpServerHandlers{}
	handlers["/simple"] = simpleHandler
	handlers["/exceed/retries"] = exceedRetriesHandler
	handlers["/window"] = windowHandler
	handlers["/"] = http.NotFound

	port, err := tests.StartHttpServer(handlers)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	return port
}
