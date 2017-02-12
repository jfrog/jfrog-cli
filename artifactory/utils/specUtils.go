package utils

import (
	"encoding/json"
	"github.com/jfrogdev/jfrog-cli-go/utils/io/fileutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"strings"
	"strconv"
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
	Pattern   string
	Target    string
	Props     string
	Recursive string
	Flat      string
	Regexp    string
	Aql       Aql
	Build	  string
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

func CreateSpecFromFile(specFilePath string) (spec *SpecFiles, err error) {
	spec = new(SpecFiles)
	content, err := fileutils.ReadFile(specFilePath)
	if cliutils.CheckError(err) != nil {
		return
	}

	err = json.Unmarshal(content, spec)
	if cliutils.CheckError(err) != nil {
		return
	}
	return
}

func CreateSpec(pattern, target, props, build string, recursive, flat, regexp bool) (spec *SpecFiles) {
	spec = &SpecFiles{
		Files: []File{
			{
				Pattern:   pattern,
				Target:    target,
				Props:     props,
				Build:	   build,
				Recursive: strconv.FormatBool(recursive),
				Flat:      strconv.FormatBool(flat),
				Regexp:    strconv.FormatBool(regexp),
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

type SpecType string