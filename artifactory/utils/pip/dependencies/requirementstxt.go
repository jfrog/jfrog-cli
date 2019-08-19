package dependencies

import (
	"bufio"
	"fmt"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Dependencies extractor for requirements.txt
type requirementsExtractor struct {
	allDependencies  map[string]*buildinfo.Dependency
	childrenMap      map[string][]string
	rootDependencies []string
	regExps          []requirementRegExp
	skipRegexp       *regexp.Regexp

	requirementsFilePath string
	pythonExecutablePath string
}

func NewRequirementsExtractor(fileName, projectRoot, pythonExecutablePath string) Extractor {
	newExtractor := &requirementsExtractor{requirementsFilePath: filepath.Join(projectRoot, fileName), pythonExecutablePath: pythonExecutablePath}
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
	environmentPackages, err := BuildPipDependencyMap(extractor.pythonExecutablePath, nil)
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
	var line, previousLine string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		previousLine = line
		line = scanner.Text()

		// Check if line is commented out.
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Check line for pattern matching.
		depName, err := extractor.consumeLine(line)
		if err != nil {
			return nil, err
		}

		// Check if failed parsing line.
		if depName == "" {
			// No match found for line, check if previous line ends with unescaped '\' meaning skip this line.
			// TODO: check that the line doesn't get commented by looking for ' #'.
			shouldSkip := extractor.shouldSkipNextRequirementsLine(previousLine)
			if shouldSkip {
				continue
			}

			// No match found.
			log.Info(fmt.Sprintf("Failed parsing requirement: '%s' in file: '%s'.", line, extractor.requirementsFilePath))
			continue
		}

		// Append dependency.
		dependencies = append(dependencies, depName)
	}

	// Check for scanner error.
	if err := scanner.Err(); err != nil {
		errorutils.CheckError(err)
		return nil, err
	}

	return dependencies, nil
}

// Iterate over requirementRegExp until match is found.
func (extractor *requirementsExtractor) consumeLine(line string) (string, error) {
	for _, regexp := range extractor.regExps {
		matched := regexp.regExp.Match([]byte(line))
		if !matched {
			continue
		}

		// We have a match.
		matchedResults := regexp.regExp.FindStringSubmatch(line)
		if len(matchedResults) < regexp.matchGroup+1 {
			// Expecting matchResults size to be at least 'regexp.matchGroup'.
			return "", nil
		}

		return matchedResults[regexp.matchGroup], nil
	}

	// No matches found.
	return "", nil
}

// In the requirements.txt file, line ending with unescaped '/' means that the next line is a continuance.
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
	// TODO: regexps 2-5 can be replaced by: ^((((git)|(hg)|(svn)|(bzr))\+)|(git:\/\/))\w.*?\w.*\#egg=([\w-]+) - with match-group 9 to catch all.
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
	extractor.skipRegexp, err = utils.GetRegExp(`.*\\$`)
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
