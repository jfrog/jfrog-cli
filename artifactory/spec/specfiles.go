package spec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

const fileSpecCannotIncludeBothPropertiesValidationMessage = "Spec cannot include both '%s' and '%s.'"

type SpecFiles struct {
	Files []File
}

func (spec *SpecFiles) Get(index int) *File {
	if index < len(spec.Files) {
		return &spec.Files[index]
	}
	return new(File)
}

func CreateSpecFromFile(specFilePath string, specVars map[string]string) (spec *SpecFiles, err error) {
	spec = new(SpecFiles)
	content, err := fileutils.ReadFile(specFilePath)
	if errorutils.CheckError(err) != nil {
		return
	}

	if len(specVars) > 0 {
		content = replaceSpecVars(content, specVars)
	}

	err = json.Unmarshal(content, spec)
	if errorutils.CheckError(err) != nil {
		return
	}
	return
}

func replaceSpecVars(content []byte, specVars map[string]string) []byte {
	log.Debug("Replacing variables in the provided File Spec: \n" + string(content))
	for key, val := range specVars {
		key = "${" + key + "}"
		log.Debug(fmt.Sprintf("Replacing '%s' with '%s'", key, val))
		content = bytes.Replace(content, []byte(key), []byte(val), -1)
	}
	log.Debug("The reformatted File Spec is: \n" + string(content))
	return content
}

type File struct {
	Aql             utils.Aql
	Pattern         string
	// Deprecated, use Exclusions instead
	ExcludePatterns []string
	Exclusions      []string
	Target          string
	Explode         string
	Props           string
	ExcludeProps    string
	SortOrder       string
	SortBy          []string
	Offset          int
	Limit           int
	Build           string
	Recursive       string
	Flat            string
	Regexp          string
	IncludeDirs     string
	ArchiveEntries  string
}

func (f File) IsFlat(defaultValue bool) (bool, error) {
	return clientutils.StringToBool(f.Flat, defaultValue)
}

func (f File) IsExplode(defaultValue bool) (bool, error) {
	return clientutils.StringToBool(f.Explode, defaultValue)
}

func (f File) IsRegexp(defaultValue bool) (bool, error) {
	return clientutils.StringToBool(f.Regexp, defaultValue)
}

func (f File) IsRecursive(defaultValue bool) (bool, error) {
	return clientutils.StringToBool(f.Recursive, defaultValue)
}

func (f File) IsIncludeDirs(defaultValue bool) (bool, error) {
	return clientutils.StringToBool(f.IncludeDirs, defaultValue)
}

func (f *File) ToArtifactoryCommonParams() *utils.ArtifactoryCommonParams {
	params := new(utils.ArtifactoryCommonParams)
	params.Aql = f.Aql
	params.Pattern = f.Pattern
	params.ExcludePatterns = f.ExcludePatterns
	params.Exclusions = f.Exclusions
	params.Target = f.Target
	params.Props = f.Props
	params.ExcludeProps = f.ExcludeProps
	params.Build = f.Build
	params.SortOrder = f.SortOrder
	params.SortBy = f.SortBy
	params.Offset = f.Offset
	params.Limit = f.Limit
	params.ArchiveEntries = f.ArchiveEntries
	return params
}

func ValidateSpec(files []File, isTargetMandatory, isSearchBasedSpec bool) error {
	if len(files) == 0 {
		return errors.New("Spec must include at least one file group")
	}
	excludePatternsUsed := false
	for _, file := range files {
		isAql := len(file.Aql.ItemsFind) > 0
		isPattern := len(file.Pattern) > 0
		isExcludePatterns := len(file.ExcludePatterns) > 0 && len(file.ExcludePatterns[0]) > 0
		excludePatternsUsed = excludePatternsUsed || isExcludePatterns
		isExclusions := len(file.Exclusions) > 0 && len(file.Exclusions[0]) > 0
		isTarget := len(file.Target) > 0
		isSortOrder := len(file.SortOrder) > 0
		isSortBy := len(file.SortBy) > 0
		isBuild := len(file.Build) > 0
		isValidSortOrder := file.SortOrder == "asc" || file.SortOrder == "desc"

		if isTargetMandatory && !isTarget {
			return errors.New("Spec must include target.")
		}
		if !isSearchBasedSpec && !isPattern {
			return errors.New("Spec must include a pattern.")
		}
		if isSearchBasedSpec && !isAql && !isPattern && !isBuild {
			return errors.New("Spec must include either aql, pattern or build.")
		}
		if isAql && isPattern {
			return errors.New(fmt.Sprintf(fileSpecCannotIncludeBothPropertiesValidationMessage, "aql", "pattern"))
		}
		if isAql && isExcludePatterns {
			return errors.New(fmt.Sprintf(fileSpecCannotIncludeBothPropertiesValidationMessage, "aql", "exclude-patterns"))
		}
		if isAql && isExclusions {
			return errors.New(fmt.Sprintf(fileSpecCannotIncludeBothPropertiesValidationMessage, "aql", "exclusions"))
		}
		if isExclusions && isExcludePatterns {
			return errors.New(fmt.Sprintf(fileSpecCannotIncludeBothPropertiesValidationMessage, "exclusions", "exclude-patterns"))
		}
		if !isSortBy && isSortOrder {
			return errors.New("Spec cannot include 'sort-order' if 'sort-by' is not included")
		}
		if isSortOrder && !isValidSortOrder {
			return errors.New("The value of 'sort-order' can only be 'asc' or 'desc'.")
		}
		if isBuild && isSearchBasedSpec {
			if err := validateFileSpecWithBuild(file); err != nil {
				return err
			}
		}
	}
	if excludePatternsUsed {
		showDeprecationOnExcludePatterns()
	}
	return nil
}

func validateFileSpecWithBuild(file File) error {
	isOffset := file.Offset > 0
	isLimit := file.Limit > 0

	if isOffset {
		return errors.New(fmt.Sprintf(fileSpecCannotIncludeBothPropertiesValidationMessage, "build", "offset"))
	}
	if isLimit {
		return errors.New(fmt.Sprintf(fileSpecCannotIncludeBothPropertiesValidationMessage, "build", "limit"))
	}
	return nil
}

func showDeprecationOnExcludePatterns() {
	log.Warn(`exclude-patterns is deprecated. Please use exclusions instead.
	Unlike exclude-patterns, the exclusions take into account the repository.
	For example: 
	exclude-patterns = 'a.zip' 
	can be translated to
	exclusions = 'repo-name/a.zip' or '*/a.zip'`)
}