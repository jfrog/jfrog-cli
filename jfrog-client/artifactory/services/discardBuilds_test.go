package services

import (
	"testing"
	"time"
)

func TestCalculateMinimumBuildDate(t *testing.T) {

	layout := "2006-01-02T15:04:05.000-0700"
	time1, _ := time.Parse(layout, "2018-05-07T17:34:49.729+0300")
	time2, _ := time.Parse(layout, "2018-05-07T17:34:49.729+0300")
	time3, _ := time.Parse(layout, "2018-05-07T17:34:49.729+0300")

	tests := []struct {
		testName      string
		startingDate  time.Time
		maxDaysString string
		expectedTime  string
	}{
		{"test_max_days=3", time1, "3", "2018-05-04T17:34:49.729+0300"},
		{"test_max_days=0", time2, "0", "2018-05-07T17:34:49.729+0300"},
		{"test_max_days=-1", time3, "-3", "2018-05-10T17:34:49.729+0300"},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			actual, _ := calculateMinimumBuildDate(test.startingDate, test.maxDaysString)
			if test.expectedTime != actual {
				t.Errorf("Test name: %s: Expected: %s, Got: %s", test.testName, test.expectedTime, actual)
			}
		})
	}
}
