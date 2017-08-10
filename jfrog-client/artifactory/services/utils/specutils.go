package utils

import (
	"strings"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
)

const (
	WILDCARD SpecType = "wildcard"
	SIMPLE   SpecType = "simple"
	AQL      SpecType = "aql"
)

type SpecType string

type Aql struct {
	ItemsFind string `json:"items.find"`
}

type ArtifactoryCommonParams struct {
	Aql         Aql
	Pattern     string
	Target      string
	Props       string
	Build       string
	Recursive   bool
	IncludeDirs bool
	Regexp      bool
}

type FileGetter interface {
	GetAql() Aql
	GetPattern() string
	SetPattern(pattern string)
	GetTarget() string
	SetTarget(target string)
	GetProps() string
	GetBuild() string
	GetSpecType() (specType SpecType)
	IsRegexp() bool
	IsRecursive() bool
	IsIncludeDirs() bool
}

func (params *ArtifactoryCommonParams) GetPattern() string {
	return params.Pattern
}

func (params *ArtifactoryCommonParams) SetPattern(pattern string) {
	params.Pattern = pattern
}

func (params *ArtifactoryCommonParams) SetTarget(target string) {
	params.Target = target
}

func (params *ArtifactoryCommonParams) GetTarget() string {
	return params.Target
}

func (params *ArtifactoryCommonParams) GetProps() string {
	return params.Props
}

func (params *ArtifactoryCommonParams) IsRecursive() bool {
	return params.Recursive
}

func (params *ArtifactoryCommonParams) IsRegexp() bool {
	return params.Regexp
}

func (params *ArtifactoryCommonParams) GetAql() Aql {
	return params.Aql
}

func (params *ArtifactoryCommonParams) GetBuild() string {
	return params.Build
}

func (params ArtifactoryCommonParams) IsIncludeDirs() bool {
	return params.IncludeDirs
}

func (params *ArtifactoryCommonParams) SetProps(props string) {
	params.Props = props
}

func (aql *Aql) UnmarshalJSON(value []byte) error {
	str := string(value)
	first := strings.Index(str[strings.Index(str, "{")+1:], "{")
	last := strings.LastIndex(str, "}")

	aql.ItemsFind = str[first:last]
	return nil
}

func (params ArtifactoryCommonParams) GetSpecType() (specType SpecType) {
	switch {
	case params.Pattern != "" && (IsWildcardPattern(params.Pattern) || params.Build != ""):
		specType = WILDCARD
	case params.Pattern != "":
		specType = SIMPLE
	case params.Aql.ItemsFind != "":
		specType = AQL
	}
	return specType
}
