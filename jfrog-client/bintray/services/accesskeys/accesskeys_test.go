package accesskeys

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray/services/utils/tests"
	"testing"
)

func TestShowAccessKeys(t *testing.T) {
	expected := "https://api.bintray.com/orgs/org/access_keys"
	path := getAccessKeysPath(tests.CreateBintrayDetails(), "org")
	if path != expected {
		t.Error("Unexpected result returned from GetAccessKeysPath. Expected: " + expected + " Got " + path)
	}

	expected = "https://api.bintray.com/users/user/access_keys"
	path = getAccessKeysPath(tests.CreateBintrayDetails(), "")
	if path != expected {
		t.Error("Unexpected result returned from GetAccessKeysPath. Expected: " + expected + " Got " + path)
	}
}

func TestShowAndDeleteAccessKey(t *testing.T) {
	expected := "https://api.bintray.com/orgs/org/access_keys/acc-key-id"
	path := getAccessKeyPath(tests.CreateBintrayDetails(), "acc-key-id", "org")
	if path != expected {
		t.Error("Unexpected result returned from GetAccessKeyPath. Expected: " + expected + " Got " + path)
	}

	expected = "https://api.bintray.com/users/user/access_keys/acc-key-id"
	path = getAccessKeyPath(tests.CreateBintrayDetails(), "acc-key-id", "")
	if path != expected {
		t.Error("Unexpected result returned from GetAccessKeyPath. Expected: " + expected + " Got " + path)
	}
}

func TestCreateAccessKey(t *testing.T) {
	expected := `{"id":"access-key-id","expiry":123,"existence_check":{"url":"ex-check-url","cache_for_secs":123},"white_cidrs":["white-cidrs"],"black_cidrs":["black-cidrs"],"api_only":false}`
	data, err := buildAccessKeyJson(createAccessKeyFlags(true))
	if err != nil {
		t.Error(err)
	}
	if string(data) != expected {
		t.Error("Unexpected result returned from BuildAccessKeyJson. Expected: " + expected + " Got " + string(data))
	}
}

func TestUpdateAccessKey(t *testing.T) {
	expected := `{"expiry":123,"existence_check":{"url":"ex-check-url","cache_for_secs":123},"white_cidrs":["white-cidrs"],"black_cidrs":["black-cidrs"],"api_only":false}`
	data, err := buildAccessKeyJson(createAccessKeyFlags(false))
	if err != nil {
		t.Error(err)
	}
	if string(data) != expected {
		t.Error("Unexpected result returned from BuildAccessKeyJson. Expected: " + expected + " Got " + string(data))
	}
}

func createAccessKeyFlags(create bool) *Params {
	var id string
	if create {
		id = "access-key-id"
	}

	return &Params{
		Id:                  id,
		Password:            "password",
		Expiry:              123,
		ExistenceCheckUrl:   "ex-check-url",
		ExistenceCheckCache: 123,
		WhiteCidrs:          "white-cidrs",
		BlackCidrs:          "black-cidrs",
		ApiOnly:             false}
}
