package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
)

func SetProps(spec *utils.SpecFiles, flags utils.CommonFlags, props string) error {
	return utils.SetProps(spec, flags, props)
}