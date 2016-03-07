package commands

import (
    "github.com/JFrogDev/jfrog-cli-go/bintray/tests"
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
    expected := "{\"id\": \"access-key-id\",\"password\": \"password\",\"expiry\": \"expiry\",\"existence_check\": {\"url\": \"ex-check-url\",\"cache_for_secs\": \"123\"},\"white_cidrs\": [\"white-cidrs\"],\"black_cidrs\": [\"black-cidrs\"]}"
    data := BuildAccessKeyJson(createAccessKeyFlags(), true)
    if data != expected {
        t.Error("Unexpected result returned from BuildAccessKeyJson. Expected: " + expected + " Got " + data)
    }
}

func TestUpdateAccessKey(t *testing.T) {
    expected := "{\"password\": \"password\",\"expiry\": \"expiry\",\"existence_check\": {\"url\": \"ex-check-url\",\"cache_for_secs\": \"123\"},\"white_cidrs\": [\"white-cidrs\"],\"black_cidrs\": [\"black-cidrs\"]}"
    data := BuildAccessKeyJson(createAccessKeyFlags(), false)
    if data != expected {
        t.Error("Unexpected result returned from BuildAccessKeyJson. Expected: " + expected + " Got " + data)
    }
}

func createAccessKeyFlags() *AccessKeyFlags {
	return &AccessKeyFlags{
		BintrayDetails:      tests.CreateBintrayDetails(),
		Id:                  "access-key-id",
		Password:            "password",
		Expiry:              "expiry",
		ExistenceCheckUrl:   "ex-check-url",
		ExistenceCheckCache: 123,
		WhiteCidrs:          "white-cidrs",
		BlackCidrs:          "black-cidrs"}
}