package utils

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"encoding/json"
	"strconv"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"fmt"
	"bytes"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"errors"
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
	cliutils.CliLogger.Debug("Replacing variables in the provided File Spec: \n" + string(content))
	for key, val := range specVars {
		key = "${" + key + "}"
		cliutils.CliLogger.Debug(fmt.Sprintf("Replacing '%s' with '%s'", key, val))
		content = bytes.Replace(content, []byte(key), []byte(val), -1)
	}
	cliutils.CliLogger.Debug("The reformatted File Spec is: \n" + string(content))
	return content
}

func CreateSpec(pattern string, excludePatterns []string, target, props, build string, recursive, flat, regexp, includeDirs bool) (spec *SpecFiles) {
	spec = &SpecFiles{
		Files: []File{
			{
				Pattern:         pattern,
				ExcludePatterns: excludePatterns,
				Target:          target,
				Props:           props,
				Build:           build,
				Recursive:       strconv.FormatBool(recursive),
				Flat:            strconv.FormatBool(flat),
				Regexp:          strconv.FormatBool(regexp),
				IncludeDirs:     strconv.FormatBool(includeDirs),
			},
		},
	}
	return spec
}

type File struct {
	Aql             utils.Aql
	Pattern         string
	ExcludePatterns []string
	Target          string
	Props           string
	Build           string
	Recursive       string
	Flat            string
	Regexp          string
	IncludeDirs     string
}

func (f File) IsFlat(defaultValue bool) (bool, error) {
	return cliutils.StringToBool(f.Flat, defaultValue)
}

func (f *File) ToArtifatoryUploadParams() (*utils.ArtifactoryCommonParams, error) {
	params := f.toArtifactoryCommonParams()

	recursive, err := cliutils.StringToBool(f.Recursive, true)
	if err != nil {
		return nil, err
	}
	params.Recursive = recursive

	regexp, err := cliutils.StringToBool(f.Regexp, false)
	if err != nil {
		return nil, err
	}
	params.Regexp = regexp

	includeDirs, err := cliutils.StringToBool(f.IncludeDirs, false)
	if err != nil {
		return nil, err
	}
	params.IncludeDirs = includeDirs
	return params, nil
}

func (f *File) ToArtifatoryDownloadParams() (*utils.ArtifactoryCommonParams, error) {
	params := f.toArtifactoryCommonParams()
	recursive, err := cliutils.StringToBool(f.Recursive, true)
	if err != nil {
		return nil, err
	}
	params.Recursive = recursive

	includeDirs, err := cliutils.StringToBool(f.IncludeDirs, false)
	if err != nil {
		return nil, err
	}
	params.IncludeDirs = includeDirs
	return params, nil
}

func (f *File) ToArtifatoryDeleteParams() (*utils.ArtifactoryCommonParams, error) {
	params := f.toArtifactoryCommonParams()
	recursive, err := cliutils.StringToBool(f.Recursive, true)
	if err != nil {
		return nil, err
	}
	params.Recursive = recursive
	return params, nil
}

func (f *File) ToArtifatorySearchParams() (*utils.ArtifactoryCommonParams, error) {
	params := f.toArtifactoryCommonParams()
	recursive, err := cliutils.StringToBool(f.Recursive, true)
	if err != nil {
		return nil, err
	}
	params.Recursive = recursive

	return params, nil
}

func (f *File) ToArtifatoryMoveCopyParams() (*utils.ArtifactoryCommonParams, error) {
	params := f.toArtifactoryCommonParams()
	recursive, err := cliutils.StringToBool(f.Recursive, true)
	if err != nil {
		return nil, err
	}
	params.Recursive = recursive
	return params, nil
}

func (f *File) ToArtifatorySetPropsParams() (*utils.ArtifactoryCommonParams, error) {
	params := f.toArtifactoryCommonParams()
	recursive, err := cliutils.StringToBool(f.Recursive, false)
	if err != nil {
		return nil, err
	}
	params.Recursive = recursive
	params.IncludeDirs, err = cliutils.StringToBool(f.IncludeDirs, false)
	if err != nil {
		return nil, err
	}
	return params, nil
}

func (f *File) toArtifactoryCommonParams() *utils.ArtifactoryCommonParams {
	params := new(utils.ArtifactoryCommonParams)
	params.Aql = f.Aql
	params.Pattern = f.Pattern
	params.ExcludePatterns = f.ExcludePatterns
	params.Target = f.Target
	params.Props = f.Props
	params.Build = f.Build
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

		if isTargetMandatory && !isTarget {
			return errors.New("Spec must include the target values")
		}
		if !isAql && !isPattern {
			return errors.New("Spec must include either the aql or pattern values")
		}
		if isAql && isPattern {
			return errors.New("Spec cannot include both the aql and pattern values")
		}
		if isAql && isExcludePattern {
			return errors.New("Spec cannot include both the aql and exclude-patterns values")
		}
	}
	return nil
}