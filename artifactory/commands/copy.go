package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
)

// Copies the artifacts using the specified move pattern.
func Copy(sourcePattern, destPath string, flags *utils.MoveFlags) {
	utils.MoveFilesWrapper(sourcePattern, destPath, flags, utils.COPY)
}
