package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
)

// Moves the artifacts using the specified move pattern.
func Move(moveSpec *utils.SpecFiles, flags *utils.MoveFlags) error {
	return utils.MoveFilesWrapper(moveSpec, flags, utils.MOVE)
}
