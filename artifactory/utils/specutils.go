package utils

import (
	"encoding/json"
	"github.com/jfrogdev/jfrog-cli-go/utils/io/fileutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"strings"
	"strconv"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
	"bytes"
	"fmt"
)

const (
	WILDCARD SpecType = "wildcard"
	SIMPLE SpecType = "simple"
	AQL SpecType = "aql"
)

type Aql struct {
	ItemsFind string `json:"items.find"`
}

type File struct {
	Pattern     string
	Target      string
	Props       string
	Recursive   string
	Flat        string
	Regexp      string
	Aql         Aql
	Build       string
	IncludeDirs string
}

type SpecFiles struct {
	Files []File
}

func (spec *SpecFiles) Get(index int) *File {
	if index < len(spec.Files) {
		return &spec.Files[index]
	}
	return new(File)
}

func (aql *Aql) UnmarshalJSON(value []byte) error {
	str := string(value)
	first := strings.Index(str[strings.Index(str, "{") + 1 :], "{")
	last := strings.LastIndex(str, "}")

	aql.ItemsFind = cliutils.StripChars(str[first:last], "\n\t ")
	return nil
}

func CreateSpecFromFile(specFilePath string, specVars map[string]string) (spec *SpecFiles, err error) {
	spec = new(SpecFiles)
	content, err := fileutils.ReadFile(specFilePath)
	if cliutils.CheckError(err) != nil {
		return
	}

	if len(specVars) > 0 {
		content = replaceSpecVars(content, specVars)
	}

	err = json.Unmarshal(content, spec)
	if cliutils.CheckError(err) != nil {
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

func CreateSpec(pattern, target, props, build string, recursive, flat, regexp, includeDirs bool) (spec *SpecFiles) {
	spec = &SpecFiles{
		Files: []File{
			{
				Pattern:     pattern,
				Target:      target,
				Props:       props,
				Build:       build,
				Recursive:   strconv.FormatBool(recursive),
				Flat:        strconv.FormatBool(flat),
				Regexp:      strconv.FormatBool(regexp),
				IncludeDirs: strconv.FormatBool(includeDirs),
			},
		},
	}
	return spec
}

func (file File) GetSpecType() (specType SpecType) {
	switch {
	case file.Pattern != "" && (IsWildcardPattern(file.Pattern) || file.Build != ""):
		specType = WILDCARD
	case file.Pattern != "":
		specType = SIMPLE
	case file.Aql.ItemsFind != "" :
		specType = AQL
	}
	return specType
}

func (file File) IsIncludeDirs() bool {
	return file.IncludeDirs == "true"
}

type SpecType string