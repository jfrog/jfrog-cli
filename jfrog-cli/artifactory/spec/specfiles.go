package spec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	clientutils "github.com/jfrog/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
)

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
	SortOrder       string
	SortBy          []string
	Offset          int
	Limit           int
	Build           string
	Recursive       string
	Flat            string
	Regexp          string
	IncludeDirs     string
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

func (f *File) ToArtifatoryUploadParams() (*utils.ArtifactoryCommonParams, error) {
	params := f.ToArtifactoryCommonParams()

	recursive, err := clientutils.StringToBool(f.Recursive, true)
	if err != nil {
		return nil, err
	}
	params.Recursive = recursive

	regexp, err := clientutils.StringToBool(f.Regexp, false)
	if err != nil {
		return nil, err
	}
	params.Regexp = regexp

	includeDirs, err := clientutils.StringToBool(f.IncludeDirs, false)
	if err != nil {
		return nil, err
	}
	params.IncludeDirs = includeDirs
	return params, nil
}

func (f *File) ToArtifatoryDownloadParams() (*utils.ArtifactoryCommonParams, error) {
	params := f.ToArtifactoryCommonParams()
	recursive, err := clientutils.StringToBool(f.Recursive, true)
	if err != nil {
		return nil, err
	}
	params.Recursive = recursive

	includeDirs, err := clientutils.StringToBool(f.IncludeDirs, false)
	if err != nil {
		return nil, err
	}
	params.IncludeDirs = includeDirs
	return params, nil
}

func (f *File) ToArtifatoryDeleteParams() (*utils.ArtifactoryCommonParams, error) {
	params := f.ToArtifactoryCommonParams()
	recursive, err := clientutils.StringToBool(f.Recursive, true)
	if err != nil {
		return nil, err
	}
	params.Recursive = recursive
	return params, nil
}

func (f *File) ToArtifatorySearchParams() (*utils.ArtifactoryCommonParams, error) {
	params := f.ToArtifactoryCommonParams()
	recursive, err := clientutils.StringToBool(f.Recursive, true)
	if err != nil {
		return nil, err
	}
	params.Recursive = recursive

	return params, nil
}

func (f *File) ToArtifatoryMoveCopyParams() (*utils.ArtifactoryCommonParams, error) {
	params := f.ToArtifactoryCommonParams()
	recursive, err := clientutils.StringToBool(f.Recursive, true)
	if err != nil {
		return nil, err
	}
	params.Recursive = recursive
	return params, nil
}

func (f *File) ToArtifatorySetPropsParams() (*utils.ArtifactoryCommonParams, error) {
	params := f.ToArtifactoryCommonParams()
	recursive, err := clientutils.StringToBool(f.Recursive, false)
	if err != nil {
		return nil, err
	}
	params.Recursive = recursive
	params.IncludeDirs, err = clientutils.StringToBool(f.IncludeDirs, false)
	if err != nil {
		return nil, err
	}
	return params, nil
}

func (f *File) ToArtifactoryCommonParams() *utils.ArtifactoryCommonParams {
	params := new(utils.ArtifactoryCommonParams)
	params.Aql = f.Aql
	params.Pattern = f.Pattern
	params.ExcludePatterns = f.ExcludePatterns
	params.Target = f.Target
	params.Props = f.Props
	params.Build = f.Build
	params.SortOrder = f.SortOrder
	params.SortBy = f.SortBy
	params.Offset = f.Offset
	params.Limit = f.Limit
	return params
}

func ValidateSpec(files []File, isTargetMandatory bool) error {
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
		isOffset := file.Offset > 0
		isLimit := file.Limit > 0
		isBuild := len(file.Build) > 0
		isValidSortOrder := file.SortOrder == "asc" || file.SortOrder == "desc"

		if isTargetMandatory && !isTarget {
			return errors.New("Spec must include the target properties")
		}
		if !isAql && !isPattern {
			return errors.New("Spec must include either the aql or pattern properties")
		}
		if isAql && isPattern {
			return errors.New("Spec cannot include both the aql and pattern properties")
		}
		if isAql && isExcludePattern {
			return errors.New("Spec cannot include both the aql and exclude-patterns properties")
		}
		if isBuild && isOffset {
			return errors.New("Spec cannot include both the 'build' and 'offset' properties")
		}
		if isBuild && isLimit {
			return errors.New("Spec cannot include both the 'build' and 'limit' properties")
		}
		if !isSortBy && isSortOrder {
			return errors.New("Spec cannot include 'sort-order' if 'sort-by' is not included")
		}
		if isSortOrder && !isValidSortOrder {
			return errors.New("The value of 'sort-order'can only be 'asc' or 'desc'.")
		}
	}
	return nil
}
