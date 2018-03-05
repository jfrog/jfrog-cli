package utils

import "testing"

func TestToEncodedString(t *testing.T) {
	tests := []struct {
		props Properties
		expected   string
	}{
		{Properties{[]Property{{Key: "a", Value: "b"}}},"a=b"},
		{Properties{[]Property{{Key: "a;a", Value: "b;a"}}},"a%3Ba=b%3Ba"},
		{Properties{[]Property{{Key: "a", Value: "b"}}},"a=b"},
		{Properties{[]Property{{Key: ";a", Value: ";b"}}},"%3Ba=%3Bb"},
		{Properties{[]Property{{Key: ";a", Value: ";b"},{Key: ";a", Value: ";b"},{Key: "aaa", Value: "bbb"}}},"%3Ba=%3Bb;%3Ba=%3Bb;aaa=bbb"},
		{Properties{[]Property{{Key: "a;", Value: "b;"}}},"a%3B=b%3B"},
	}
	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			test.props.ToEncodedString()
			if test.expected != test.props.ToEncodedString() {
				t.Error("Failed to encode properties. The propertyes", test.props.ToEncodedString(), "expected to be encoded to", test.expected)
			}
		})
	}
}
