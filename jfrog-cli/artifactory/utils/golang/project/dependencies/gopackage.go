package dependencies

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils/golang"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"regexp"
)

type GoPackage interface {
	PopulateModIfNeededAndPublish(targetRepo string, cache *golang.DynamicCache,details *config.ArtifactoryDetails) error
	Init() error
	prepareAndPublish(targetRepo string, cache *golang.DynamicCache, details *config.ArtifactoryDetails) error
	New(cachePath string, dependency Package) GoPackage
}

type RegExp struct {
	notEmptyModRegex *regexp.Regexp
	indirectRegex    *regexp.Regexp
}

func (reg *RegExp) GetNotEmptyModRegex() *regexp.Regexp {
	return reg.notEmptyModRegex
}

func (reg *RegExp) GetIndirectRegex() *regexp.Regexp {
	return reg.indirectRegex
}
