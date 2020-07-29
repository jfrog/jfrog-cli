package utils

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	logUtils "github.com/jfrog/jfrog-cli/utils/log"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/stretchr/testify/assert"
)

func TestPrintSearchResults(t *testing.T) {
	testdataPath, err := getTestDataPath()
	assert.NoError(t, err)
	reader := content.NewContentReader(filepath.Join(testdataPath, "search_results.json"), content.DefaultKey)

	previousLog := log.Logger
	newLog := log.NewLogger(logUtils.GetCliLogLevel(), nil)
	// Restore previous logger when the function returns.
	defer log.SetLogger(previousLog)

	// Set new logger with output redirection to buffer.
	buffer := &bytes.Buffer{}
	newLog.SetOutputWriter(buffer)
	log.SetLogger(newLog)

	// Print search result.
	err = PrintSearchResults(reader)
	assert.NoError(t, err)

	// Compare output.
	logOutput := buffer.Bytes()
	compareResult := bytes.Compare(logOutput, []byte(expectedLogOutput))
	assert.Equal(t, 0, compareResult)
}

func getTestDataPath() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	return filepath.Join(dir, "testdata"), nil
}

const expectedLogOutput = `[
  {
    "path": "jfrog-cli-tests-repo1-1595270324/a/b/c/c2.in",
    "type": "file",
    "size": 11,
    "created": "2020-07-20T21:39:38.374+03:00",
    "modified": "2020-07-20T21:39:38.332+03:00",
    "sha1": "a4f912be11e7d1d346e34c300e6d4b90e136896e",
    "md5": "82b6d565393a3fd1cc4778b1d53c0664",
    "props": {
      "c": [
        "3"
      ]
    }
  },
  {
    "path": "jfrog-cli-tests-repo1-1595270324/a/b/c/c3.in",
    "type": "file",
    "size": 11,
    "created": "2020-07-20T21:39:38.392+03:00",
    "modified": "2020-07-20T21:39:38.332+03:00",
    "sha1": "2d6ee506188db9b816a6bfb79c5df562fc1d8658",
    "md5": "d8020b86244956f647cf1beff5acdb90",
    "props": {
      "c": [
        "3"
      ]
    }
  },
  {
    "path": "jfrog-cli-tests-repo1-1595270324/a/b/b2.in",
    "type": "file",
    "size": 9,
    "created": "2020-07-20T21:39:38.413+03:00",
    "modified": "2020-07-20T21:39:38.410+03:00",
    "sha1": "3b60b837e037568856bedc1dd4952d17b3f06972",
    "md5": "6931271be1e5f98e36bdc7a05097407b",
    "props": {
      "b": [
        "1"
      ],
      "c": [
        "3"
      ]
    }
  },
  {
    "path": "jfrog-cli-tests-repo1-1595270324/a/b/b3.in",
    "type": "file",
    "size": 9,
    "created": "2020-07-20T21:39:38.413+03:00",
    "modified": "2020-07-20T21:39:38.410+03:00",
    "sha1": "ec6420d2b5f708283619b25e68f9ddd351f555fe",
    "md5": "305b21db102cf3a3d2d8c3f7e9584dba",
    "props": {
      "a": [
        "1"
      ],
      "b": [
        "2"
      ],
      "c": [
        "3"
      ]
    }
  },
  {
    "path": "jfrog-cli-tests-repo1-1595270324/a/a3.in",
    "type": "file",
    "size": 7,
    "created": "2020-07-20T21:39:38.430+03:00",
    "modified": "2020-07-20T21:39:38.428+03:00",
    "sha1": "29d38faccfe74dee60d0142a716e8ea6fad67b49",
    "md5": "73c046196302ff7218d47046cf3c0501",
    "props": {
      "a": [
        "1"
      ],
      "b": [
        "3"
      ],
      "c": [
        "3"
      ]
    }
  }
]
`
