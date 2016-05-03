package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
)

// Moves the artifacts using the specified move pattern.
func Move(sourcePattern, destPath string, flags *utils.Flags) {
	utils.MoveFilesWrapper(sourcePattern, destPath, flags, utils.MOVE)
}
