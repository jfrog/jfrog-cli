package utils

import (
	"encoding/json"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"strings"
)

const (
	WILDCARD SpecType = "wildcard"
	SIMPLE SpecType = "simple"
	AQL SpecType = "aql"
)

type Aql struct {
	ItemsFind string `json:"items.find"`
}

type Files struct {
	Pattern   string
	Target    string
	Props     string
	Recursive bool `json:"recursive,string"`
	Flat      bool `json:"flat,string"`
	Regexp    bool `json:"regexp,string"`
	Aql       Aql
}

type SpecFiles struct {
	Files []Files
}

func (spec *SpecFiles) Get(index int) *Files {
	if index < len(spec.Files) {
		return &spec.Files[index]
	}
	return new(Files)
}

func (aql *Aql) UnmarshalJSON(value []byte) error {
	str := string(value)
	first := strings.Index(str[strings.Index(str, "{") + 1 :], "{")
	last := strings.LastIndex(str, "}")

	aql.ItemsFind = cliutils.StripChars(str[first:last], "\n\t ")
	return nil
}

func CreateSpecFromFile(specFilePath string) (spec *SpecFiles, e error) {
	spec = new(SpecFiles)
	content, e := ioutils.ReadFile(specFilePath)
	if e != nil {
		return
	}
	e = json.Unmarshal(content, spec)
	if e != nil {
		return
	}
	return
}

func CreateSpec(pattern, target, props string, recursive, flat, regexp bool) (spec *SpecFiles) {
	spec = &SpecFiles{
		Files: []Files{
			{
				Pattern:   pattern,
				Target:    target,
				Props:     props,
				Recursive: recursive,
				Flat:      flat,
				Regexp:    regexp,
			},
		},
	}
	return spec
}

func (files Files) GetSpecType() (specType SpecType) {
	switch {
	case files.Pattern != "" && IsWildcardPattern(files.Pattern):
		specType = WILDCARD
	case files.Pattern != "":
		specType = SIMPLE
	case files.Aql.ItemsFind != "" :
		specType = AQL
	}
	return specType
}

type SpecType string