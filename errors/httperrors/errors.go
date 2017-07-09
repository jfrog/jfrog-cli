package httperrors

import (
	"net/http"
	"io/ioutil"
)

type ResponseStatusError struct {
	status string
	body   string
}

func CheckResponseStatusError(resp *http.Response, body []byte, expectedStatusCodes... int) error {
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
	return &ResponseStatusError{status:resp.Status, body:string(errorBody)}
}

func IsResponseStatusError(err error) bool {
	_, isRespErr := err.(*ResponseStatusError)
	return isRespErr
}

func (e *ResponseStatusError) Error() string {
	return e.status + " " + e.body
}
