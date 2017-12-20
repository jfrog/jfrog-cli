package utils

import "testing"

func TestReadMissionControlHttpMessage (t *testing.T){
	// Test simple error message
	resp := []byte("{\"errors\":[{\"message\":\"HTTP 404 Not Found\",\"type\":\"Exception\"}]}")
	expected := "HTTP 404 Not Found"
	errorMessage := ReadMissionControlHttpMessage(resp)
	if expected != errorMessage {
		t.Error("Unexpected error message. Expected: " + expected + " Got " + errorMessage)
	}
	// Test multiple messages
	resp = []byte("{\"errors\":[{\"message\":\"HTTP 404 Not Found\",\"type\":\"Exception\"},{\"message\":\"HTTP 404 Not Found\",\"type\":\"Exception\"}]}")
	expected = "HTTP 404 Not Found, HTTP 404 Not Found"
	errorMessage = ReadMissionControlHttpMessage(resp)
	if expected != errorMessage {
		t.Error("Unexpected error message. Expected: " + expected + " Got " + errorMessage)
	}
	// Test none error response
	resp = []byte("{\"data\":[{\"message\":\"HTTP 404 Not Found\",\"type\":\"Exception\"},{\"message\":\"HTTP 404 Not Found\",\"type\":\"Exception\"}]}")
	expected = "{\"data\":[{\"message\":\"HTTP 404 Not Found\",\"type\":\"Exception\"},{\"message\":\"HTTP 404 Not Found\",\"type\":\"Exception\"}]}"
	errorMessage = ReadMissionControlHttpMessage(resp)
	if expected != errorMessage {
		t.Error("Unexpected error message. Expected: " + expected + " Got " + errorMessage)
	}
	// Test response with details
	resp = []byte("{\"errors\": [{\"message\": \"Validation constraint violation\",\"details\": [\"addInstance.req.url property must be a valid URL. Invalid value: 'the'\"],\"type\": \"Validation\"}]}")
	expected = "Validation constraint violation addInstance.req.url property must be a valid URL. Invalid value: 'the'"
	errorMessage = ReadMissionControlHttpMessage(resp)
	if expected != errorMessage {
		t.Error("Unexpected error message. Expected: \n" + expected + "\n Got \n" + errorMessage)
	}
	resp = []byte("{\"errors\": [{\"message\": \"Validation constraint violation\",\"details\": [\"addInstance.req.url property must be a valid URL. Invalid value: 'the'\" , \"test\"],\"type\": \"Validation\"}]}")
	expected = "Validation constraint violation addInstance.req.url property must be a valid URL. Invalid value: 'the' test"
	errorMessage = ReadMissionControlHttpMessage(resp)
	if expected != errorMessage {
		t.Error("Unexpected error message. Expected: \n" + expected + "\n Got \n" + errorMessage)
	}
}
