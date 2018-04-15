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
		{"5.10.0", "5.5.2", 1},
		{"5.5.2", "5.15.2", -1},
		{"5.6.2", "5.50.2", -1},
		{"5.5.2", "5.0.2", 1},
		{"15.5.2", "6.0.2", 1},
		{"51.5.2", "6.0.2", 1},
		{"5.0.3", "5.0.20", -1},
		{"5.0.20", "5.0.3", 1},
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
