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

const fileSpecWithBuildNoRepoValidationMessage = "Spec cannot include both 'build' and '%s', if 'pattern' is empty or '*'."
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
	ExcludePatterns []string
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
	for _, file := range files {
		isAql := len(file.Aql.ItemsFind) > 0
		isPattern := len(file.Pattern) > 0
		isExcludePattern := len(file.ExcludePatterns) > 0 && len(file.ExcludePatterns[0]) > 0
		isTarget := len(file.Target) > 0
		isSortOrder := len(file.SortOrder) > 0
		isSortBy := len(file.SortBy) > 0
		isBuild := len(file.Build) > 0
		isValidSortOrder := file.SortOrder == "asc" || file.SortOrder == "desc"

		if isTargetMandatory && !isTarget {
			return errors.New("Spec must include target.")
		}
		if !isSearchBasedSpec && !isAql && !isPattern {
			return errors.New("Spec must include either aql or pattern.")
		}
		if isSearchBasedSpec && !isAql && !isPattern && !isBuild {
			return errors.New("Spec must include either aql, pattern or build.")
		}
		if isAql && isPattern {
			return errors.New(fmt.Sprintf(fileSpecCannotIncludeBothPropertiesValidationMessage, "aql", "pattern"))
		}
		if isAql && isExcludePattern {
			return errors.New(fmt.Sprintf(fileSpecCannotIncludeBothPropertiesValidationMessage, "aql", "exclude-patterns"))
		}
		if !isSortBy && isSortOrder {
			return errors.New("Spec cannot include 'sort-order' if 'sort-by' is not included")
		}
		if isSortOrder && !isValidSortOrder {
			return errors.New("The value of 'sort-order'can only be 'asc' or 'desc'.")
		}
		if isBuild && isSearchBasedSpec {
			err := validateFileSpecWithBuild(file, isExcludePattern)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func validateFileSpecWithBuild(file File, isExcludePattern bool) error {
	isOffset := file.Offset > 0
	isLimit := file.Limit > 0
	isArchiveEntries := len(file.ArchiveEntries) > 0
	isIncludeDirs := len(file.IncludeDirs) > 0
	isRecursive := len(file.Recursive) > 0
	isProps := len(file.Props) > 0

	if isOffset {
		return errors.New(fmt.Sprintf(fileSpecCannotIncludeBothPropertiesValidationMessage, "build", "offset"))
	}
	if isLimit {
		return errors.New(fmt.Sprintf(fileSpecCannotIncludeBothPropertiesValidationMessage, "build", "limit"))
	}

	if file.Pattern == "*" || file.Pattern == "" {
		if isExcludePattern {
			return errors.New(fmt.Sprintf(fileSpecWithBuildNoRepoValidationMessage, "exclude-patterns"))
		}
		if isArchiveEntries {
			return errors.New(fmt.Sprintf(fileSpecWithBuildNoRepoValidationMessage, "archive-entries"))
		}
		if isRecursive {
			return errors.New(fmt.Sprintf(fileSpecWithBuildNoRepoValidationMessage, "recursive"))
		}
		if isIncludeDirs {
			return errors.New(fmt.Sprintf(fileSpecWithBuildNoRepoValidationMessage, "include-dirs"))
		}
		if isProps {
			return errors.New(fmt.Sprintf(fileSpecWithBuildNoRepoValidationMessage, "props"))
		}
	}

	return nil
}
