package httperrors

import (
	"io/ioutil"
	"net/http"
)

type ResponseStatusError struct {
	status string
	body   string
}

func (e *ResponseStatusError) Error() string {
	return e.status + " " + e.body
}

// Check expected status codes and return error if needed
func CheckResponseStatus(resp *http.Response, body []byte, expectedStatusCodes ...int) error {
	for _, statusCode := range expectedStatusCodes {
		if statusCode == resp.StatusCode {
			return nil
		}
	}

	var errorBody []byte
	if len(body) > 0 {
		errorBody = body
	} else {
		errorBody, _ = ioutil.ReadAll(resp.Body)
	}
	return &ResponseStatusError{status: resp.Status, body: string(errorBody)}
}

// Check if the error is instance of ResponseStatusError
func IsResponseStatusError(err error) bool {
	_, isRespErr := err.(*ResponseStatusError)
	return isRespErr
}
