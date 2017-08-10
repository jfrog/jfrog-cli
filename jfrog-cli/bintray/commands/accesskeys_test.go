package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/bintray/tests"
	"testing"
)

func TestShowAccessKeys(t *testing.T) {
	expected := "https://api.bintray.com/orgs/org/download_keys"
	path := GetAccessKeysPath(tests.CreateBintrayDetails(), "org")
	if path != expected {
		t.Error("Unexpected result returned from GetAccessKeysPath. Expected: " + expected + " Got " + path)
	}

	expected = "https://api.bintray.com/users/user/download_keys"
	path = GetAccessKeysPath(tests.CreateBintrayDetails(), "")
	if path != expected {
		t.Error("Unexpected result returned from GetAccessKeysPath. Expected: " + expected + " Got " + path)
	}
}

func TestShowAndDeleteAccessKey(t *testing.T) {
	expected := "https://api.bintray.com/orgs/org/download_keys/acc-key-id"
	path := GetAccessKeyPath(tests.CreateBintrayDetails(), "acc-key-id", "org")
	if path != expected {
		t.Error("Unexpected result returned from GetAccessKeyPath. Expected: " + expected + " Got " + path)
	}

	expected = "https://api.bintray.com/users/user/download_keys/acc-key-id"
	path = GetAccessKeyPath(tests.CreateBintrayDetails(), "acc-key-id", "")
	if path != expected {
		t.Error("Unexpected result returned from GetAccessKeyPath. Expected: " + expected + " Got " + path)
	}
}

func TestCreateAccessKey(t *testing.T) {
	expected := `{"id":"access-key-id","expiry":123,"existence_check":{"url":"ex-check-url","cache_for_secs":123},"white_cidrs":["white-cidrs"],"black_cidrs":["black-cidrs"],"api_only":false}`
	data, err := BuildAccessKeyJson(createAccessKeyFlags(true))
	if err != nil {
		t.Error(err)
	}
	if data != expected {
		t.Error("Unexpected result returned from BuildAccessKeyJson. Expected: " + expected + " Got " + data)
	}
}

func TestUpdateAccessKey(t *testing.T) {
	expected := `{"expiry":123,"existence_check":{"url":"ex-check-url","cache_for_secs":123},"white_cidrs":["white-cidrs"],"black_cidrs":["black-cidrs"],"api_only":false}`
	data, err := BuildAccessKeyJson(createAccessKeyFlags(false))
	if err != nil {
		t.Error(err)
	}
	if data != expected {
		t.Error("Unexpected result returned from BuildAccessKeyJson. Expected: " + expected + " Got " + data)
	}
}

func createAccessKeyFlags(create bool) *AccessKeyFlags {
	var id string
	if create {
		id = "access-key-id"
	}
	return &AccessKeyFlags{
		BintrayDetails:      tests.CreateBintrayDetails(),
		Id:                  id,
		Password:            "password",
		Expiry:              123,
		ExistenceCheckUrl:   "ex-check-url",
		ExistenceCheckCache: 123,
		WhiteCidrs:          "white-cidrs",
		BlackCidrs:          "black-cidrs",
		ApiOnly:             "false"}
}