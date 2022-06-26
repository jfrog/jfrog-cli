package progressbar

import (
	"github.com/jfrog/jfrog-cli-core/v2/utils/progressbar"
	"testing"
)

func TestBuildProgressDescription(t *testing.T) {
	// Set an arbitrary terminal width
	terminalWidth = 100
	tests := getTestCases()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			desc := buildProgressDescription(test.prefix, test.path, test.extraCharsLen)

			// Validate result
			if desc != test.expectedDesc {
				t.Errorf("Expected value of: \"%s\", got: \"%s\".", test.expectedDesc, desc)
			}
		})
	}
}

func getTestCases() []testCase {
	prefix := "  downloading"
	path := "/a/path/to/a/file"
	separator := " | "

	fullDesc := " " + prefix + separator + path + separator
	emptyPathDesc := " " + prefix + separator + "..." + separator
	shortenedDesc := " " + prefix + separator + "...ggggg/path/to/a/file" + separator

	widthMinusProgress := terminalWidth - progressbar.ProgressBarWidth*2
	return []testCase{
		{"commonUseCase", prefix, path, 17, fullDesc},
		{"zeroExtraChars", prefix, path, 0, fullDesc},
		{"minDescLength", prefix, path, widthMinusProgress - len(emptyPathDesc), emptyPathDesc},
		{"longPath", prefix, "/a/longggggggggggggggggggggg/path/to/a/file", 17, shortenedDesc},
		{"longPrefix", "longggggggggggggggggggggggggg prefix", path, 17, ""},
		{"manyExtraChars", prefix, path, widthMinusProgress - len(emptyPathDesc) + 1, ""},
	}
}

type testCase struct {
	name          string
	prefix        string
	path          string
	extraCharsLen int
	expectedDesc  string
}
