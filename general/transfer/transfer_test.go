package transfer

import (
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

type isTransferRequiredTestSuite struct {
	testName          string
	transferTimestamp string
	lastModified      string
	expected          bool
}

func TestIsTransferRequired(t *testing.T) {
	tests := []isTransferRequiredTestSuite{
		{"wasn't transferred and wasn't modified", "", "", true},
		{"wasn't transferred and was modified before", "", "2022-03-26T10:11:11.872Z", true},
		{"transferred and wasn't modified", "1652614804", "", false},
		{"transferred and was modified before transfer", "1652614804", "2022-03-26T10:11:11.872Z", false},
		{"transferred and was modified after transfer", "1619608062", "2022-03-26T10:11:11.872Z", true},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			item := utils.ResultItem{Modified: test.lastModified, Properties: []utils.Property{{Key: fileTransferTimestampProperty, Value: test.transferTimestamp}}}
			required, err := isTransferRequired(item)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, required)
		})
	}
}
