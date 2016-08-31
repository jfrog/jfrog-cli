package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
)

// Copies the artifacts using the specified move pattern.
func Copy(copySpec *utils.SpecFiles, flags *utils.MoveFlags) error {
	return utils.MoveFilesWrapper(copySpec, flags, utils.COPY)
}
