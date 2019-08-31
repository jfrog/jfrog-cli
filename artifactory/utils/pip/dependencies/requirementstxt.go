package dependencies

import (
	"bufio"
	"fmt"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"os"
	"regexp"
	"strings"
)

// Dependencies extractor for requirements.txt
type requirementsExtractor struct {
	allDependencies      map[string]*buildinfo.Dependency
	childrenMap          map[string][]string
	rootDependencies     []string
	regExps              []requirementRegExp
	skipRegexp           *regexp.Regexp
	requirementsFilePath string
	pythonExecutablePath string
}

func NewRequirementsExtractor(requirementsFilePath, pythonExecutablePath string) Extractor {
	newExtractor := &requirementsExtractor{requirementsFilePath: requirementsFilePath, pythonExecutablePath: pythonExecutablePath}
	// Init regexps.
	newExtractor.initializeRegExps()
	return newExtractor
}

func (extractor *requirementsExtractor) Extract() error {
	// Parse requirements.txt, add to rootDependencies.
	dependencies, err := extractor.parseRequirementsFile()
	if errorutils.CheckError(err) != nil {
		return err
	}
	extractor.rootDependencies = dependencies

	// Get installed packages tree.
	environmentPackages, err := BuildPipDependencyMap(extractor.pythonExecutablePath)
	if err != nil {
		return nil
	}

	// Extract all project dependencies.
	allDeps, childMap, err := extractDependencies(extractor.rootDependencies, environmentPackages)
	if err != nil {
		return err
	}

	// Update extracted dependencies.
	extractor.allDependencies = allDeps
	extractor.childrenMap = childMap

	return nil
}

// Parse the provided requirements file.
// Return all found package names.
func (extractor *requirementsExtractor) parseRequirementsFile() ([]string, error) {
	var dependencies []string

	// Read file.
	file, err := os.Open(extractor.requirementsFilePath)
	if errorutils.CheckError(err) != nil {
		return nil, err
	}
	defer file.Close()

	// Check line by line for match.
	var line, previousLine, lineToConsume string
	var shouldContinue bool
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line = scanner.Text()

		lineToConsume, shouldContinue = prepareLineForConsumption(line, previousLine)
		if shouldContinue {
			previousLine = lineToConsume
			continue
		}
		previousLine = ""

		// Check line for pattern matching.
		depName, err := extractor.consumeLine(lineToConsume)
		if err != nil {
			return nil, err
		}

		// Check if failed parsing line.
		if depName == "" {
			// No match found.
			log.Info(fmt.Sprintf("Failed parsing requirement: '%s' in file: '%s'.", line, extractor.requirementsFilePath))
			continue
		}

		// Append dependency.
		dependencies = append(dependencies, strings.ToLower(depName))
	}

	// Check for scanner error.
	if err := scanner.Err(); err != nil {
		errorutils.CheckError(err)
		return nil, err
	}

	return dependencies, nil
}

func prepareLineForConsumption(line, previousLine string) (lineToConsume string, shouldContinue bool) {
	// Remove spaces from start and end of line.
	trimmedLine := strings.TrimSpace(line)

	// Check if this line continues previous line.
	if previousLine != "" {
		// If line starts with '#', ignore it and consume previousLine.
		if strings.HasPrefix(trimmedLine, "#") {
			// Consume only previous line.
			return previousLine, false
		}

		// Concatenate lines.
		lineToConsume = concatenateLines(previousLine, line)

		// Check if line ends with '\'.
		if strings.HasSuffix(line, "\\") {
			// Don't consume this line, concatenate next line.
			return lineToConsume, true
		}

		// Consume this concatenated line.
		return lineToConsume, false
	}

	// This is a new line.
	// If line starts with '#' (comment), continue.
	if strings.HasPrefix(trimmedLine, "#") {
		return "", true
	}

	// If line ends with '\', need to concatenate to next line and consume together.
	if strings.HasSuffix(line, "\\") {
		return trimmedLine, true
	}

	// Consume this line.
	return trimmedLine, false
}

func concatenateLines(firstLine, secondLine string) string {
	// Remove all trailing '\' from firtLine.
	for strings.HasSuffix(firstLine, "\\") {
		firstLine = strings.TrimSuffix(firstLine, "\\")
	}

	// Concatenate lines.
	return firstLine + secondLine
}

// Iterate over requirementRegExp until match is found.
func (extractor *requirementsExtractor) consumeLine(line string) (string, error) {
	for _, regexp := range extractor.regExps {
		matched := regexp.regExp.Match([]byte(line))
		if !matched {
			continue
		}

		// Matched.
		matchedResults := regexp.regExp.FindStringSubmatch(line)
		if len(matchedResults) < regexp.matchGroup + 1 {
			// Expecting matchResults size to be at least 'regexp.matchGroup'.
			return "", nil
		}

		return matchedResults[regexp.matchGroup], nil
	}

	// No matches found.
	return "", nil
}

// In the requirements.txt file, line ending with unescaped '\' means that the next line is a continuance.
// Thus should skip the next line when parsing.
func (extractor *requirementsExtractor) shouldSkipNextRequirementsLine(line string) bool {
	matched := extractor.skipRegexp.Match([]byte(line))
	if matched {
		// Should skip the next line.
		return true
	}
	return false
}

func (extractor *requirementsExtractor) initializeRegExps() error {
	// Order is important! pattern '^\w[\w-\.]+' matches for all regexps, thus must be last.
	// Go doesn't support Lookaheads in regexps, thus this won't work: '^(?!(git\+)|(git:)|(https?:\/\/)|(hg\+)|(svn\+)|(bzr\+))\w[\w-\.]+'
	var requirementRegExps = []requirementRegExp{
		{regExpString: `^((-e\s)?(git\+)|(git:\/\/))\w.*?\w.*\#egg=([\w-]+)`, matchGroup: 5}, // match git+, git://
		{regExpString: `^((-e\s)?hg\+)\w.*?\w.*\#egg=([\w-]+)`, matchGroup: 3},               // match hg+
		{regExpString: `^((-e\s)?svn\+)\w.*?\w.*\#egg=([\w-]+)`, matchGroup: 3},              // match svn+
		{regExpString: `^((-e\s)?bzr\+)\w.*?\w.*\#egg=([\w-]+)`, matchGroup: 3},              // match bzr+
		{regExpString: `^\w[\w-\.]+`, matchGroup: 0},                                         // match pkg ids not starting with git:, git+, http://, https://, hg+, svn+, bzr+
	}

	var err error
	// Calculate regular expressions.
	for i, regexp := range requirementRegExps {
		requirementRegExps[i].regExp, err = utils.GetRegExp(regexp.regExpString)
		if err != nil {
			return err
		}
	}
	extractor.regExps = append(extractor.regExps, requirementRegExps...)

	// Calculate skip regexp.
	extractor.skipRegexp, err = utils.GetRegExp(`.*\s\\$`)
	if err != nil {
		return err
	}

	return nil
}

type requirementRegExp struct {
	regExp       *regexp.Regexp
	regExpString string
	matchGroup   int
}

func (extractor *requirementsExtractor) AllDependencies() map[string]*buildinfo.Dependency {
	return extractor.allDependencies
}

func (extractor *requirementsExtractor) DirectDependencies() []string {
	return extractor.rootDependencies
}

func (extractor *requirementsExtractor) ChildrenMap() map[string][]string {
	return extractor.childrenMap
}

func (extractor *requirementsExtractor) PackageName() (string, error) {
	return "", nil
}
