package httputils

import (
	"bufio"
	"errors"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
	"io"
	"net/http"
	"time"
)

type ConnectHandler func() (*http.Response, error)
type ErrorHandler func([]byte) error

// Retryable connection specific errors
var (
	timeoutErr         = errors.New("read timeout")
	exhaustedErr       = errors.New("connection error: exhausted retries")
	missingRespBodyErr = errors.New("missing response body")
)

type RetryableConnection struct {
	// ReadTimeout, if read timeout time passes without any data received from the server the connection will be closed.
	ReadTimeout time.Duration

	// RetriesNum represents the number of retries following a lost connection, -1 for unlimited
	RetriesNum int

	// StableConnectionWindow sets the duration of a stable connection after which the RetriesNum is reset.
	// If 0 RetriesNum is never reset.
	// It is recommended to use longer time than ReadTimeout.
	StableConnectionWindow time.Duration

	// SleepBetweenRetries sleep time between two retires.
	SleepBetweenRetries time.Duration

	// ConnectHandler will be called for connection retry, make sure response body is not closed.
	ConnectHandler ConnectHandler

	// ErrorHandler will be called after successful connection for content errors checks.
	ErrorHandler ErrorHandler
}

func (rt *RetryableConnection) checkErrors(err error, stableConnection bool, retryCounter *int) error {
	if err != nil {
		log.Info("Connection error:", err.Error()+",", "reconnecting...")
		time.Sleep(rt.SleepBetweenRetries)
		if stableConnection {
			*retryCounter = 0
		} else {
			*retryCounter++
		}
	}

	return err
}

func (rt *RetryableConnection) Do() ([]byte, error) {
	retry := 0
	for rt.RetriesNum == -1 || retry <= rt.RetriesNum {
		resp, err := rt.ConnectHandler()
		if rt.checkErrors(err, false, &retry) != nil {
			continue
		}

		timeNow := time.Now()
		monitor := monitor{
			resp:                   resp,
			readTimeout:            rt.ReadTimeout,
			stableConnectionWindow: rt.StableConnectionWindow,
			connectionTime:         timeNow,
			lastRead:               timeNow,
		}

		result, stableConnection, err := monitor.start()
		if rt.checkErrors(err, stableConnection, &retry) != nil {
			continue
		}

		// Check for content errors (only if there are no other errors)
		if rt.ErrorHandler != nil && rt.checkErrors(rt.ErrorHandler(result), stableConnection, &retry) != nil {
			continue
		}

		return result, err
	}
	return []byte{}, exhaustedErr
}

type monitor struct {
	resp *http.Response

	readTimeout            time.Duration
	stableConnectionWindow time.Duration

	connectionTime time.Time
	lastRead       time.Time
}

func (m *monitor) start() ([]byte, bool, error) {
	if m.resp == nil || m.resp.Body == nil {
		return []byte{}, false, errorutils.CheckError(missingRespBodyErr)
	}
	defer m.resp.Body.Close()

	result := []byte{}
	errChan := make(chan error, 1)
	bodyReader := bufio.NewReader(m.resp.Body)
	go func() {
		for {
			line, _, err := bodyReader.ReadLine()
			if err == io.EOF {
				errChan <- nil
				break
			}
			if err != nil {
				errChan <- err
				break
			}
			m.lastRead = time.Now()
			result = append(result, line...)
		}
	}()

	// timeout func
	go func() {
		defer close(errChan)
		for {
			// If last read exceeded the timeout signal for timeout error.
			if m.lastRead.Add(m.readTimeout).Before(time.Now()) {
				errChan <- timeoutErr
			} else {
				// Sleep the remaining time until another timeout check is required
				time.Sleep(m.readTimeout - time.Now().Sub(m.lastRead))
			}
		}
	}()

	// Receive error or nil for successful connection.
	err := <-errChan

	// Check whether connection time is longer the provided stableConnectionWindow duration.
	// If so the connection was stable.
	stable := false
	if m.stableConnectionWindow > 0 && m.stableConnectionWindow < m.lastRead.Sub(m.connectionTime) {
		stable = true
	}
	// receive the data or fail on timeout or error
	return result, stable, errorutils.CheckError(err)
}
