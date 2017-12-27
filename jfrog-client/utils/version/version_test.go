package version

import "testing"

func TestCompare(t *testing.T) {
	tests := []struct {
		ver1     string
		ver2     string
		expected int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.1", "1.0.0", 1},
		{"1.0.0", "1.0.1", -1},
		{"1.0.x-SNAPSHOT", "1.0.x-SNAPSHOT", 0},
		{"1.1.x-SNAPSHOT", "1.0.x-SNAPSHOT", 1},
		{"2.0.x-SNAPSHOT", "1.0.x-SNAPSHOT", 1},
		{"1.0", "1.0.x-SNAPSHOT", -1},
		{"1.1", "1.0.x-SNAPSHOT", 1},
		{"1.0.x-SNAPSHOT", "1.0", 1},
		{"1.0.x-SNAPSHOT", "1.1", -1},
		{"1", "2", -1},
		{"1.0", "2.0", -1},
		{"2.1", "2.0", 1},
		{"2.a", "2.b", -1},
		{"b", "a", 1},
		{"1.0", "1", 0},
		{"1.1", "1", 1},
		{"1", "1.1", -1},
		{"", "1", -1},
		{"1", "", 1},
	}
	for _, test := range tests {
		t.Run(test.ver1+":"+test.ver2, func(t *testing.T) {
			result := Compare(test.ver1, test.ver2)
			if result != test.expected {
				t.Error("ver1:", test.ver1, "ver2:", test.ver2, "Expecting:", test.expected, "got:", result)
			}
		})
	}
}
