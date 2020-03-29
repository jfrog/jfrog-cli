package utils

import (
	"github.com/jfrog/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/utils/io"
)

func CreateDownloadServiceManager(artDetails *config.ArtifactoryDetails, threads int, dryRun bool, progressBar io.Progress) (*artifactory.ArtifactoryServicesManager, error) {
	return CreateServiceManagerWithProgressBar(artDetails, threads, dryRun, progressBar)
}

type DownloadConfiguration struct {
	Threads         int
	SplitCount      int
	MinSplitSize    int64
	Symlink         bool
	ValidateSymlink bool
	Retries         int
}
