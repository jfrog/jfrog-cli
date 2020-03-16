package utils

import (
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/utils/io"
)

func CreateUploadServiceManager(artDetails *config.ArtifactoryDetails, threads int, dryRun bool, progressBar io.Progress) (*artifactory.ArtifactoryServicesManager, error) {
	return CreateServiceManagerWithProgressBar(artDetails, threads, dryRun, progressBar)
}

type UploadConfiguration struct {
	Deb                   string
	Threads               int
	MinChecksumDeploySize int64
	Symlink               bool
	ExplodeArchive        bool
	Retries               int
}
