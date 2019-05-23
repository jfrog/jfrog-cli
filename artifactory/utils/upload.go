package utils

import (
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/utils/io"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

func CreateUploadServiceManager(artDetails *config.ArtifactoryDetails, flags *UploadConfiguration, certPath string, dryRun bool, progressBar io.Progress) (*artifactory.ArtifactoryServicesManager, error) {
	artAuth, err := artDetails.CreateArtAuthConfig()
	if err != nil {
		return nil, err
	}
	servicesConfig, err := artifactory.NewConfigBuilder().
		SetArtDetails(artAuth).
		SetDryRun(dryRun).
		SetCertificatesPath(certPath).
		SetInsecureTls(artDetails.InsecureTls).
		SetThreads(flags.Threads).
		SetLogger(log.Logger).
		Build()

	return artifactory.NewWithProgress(&artAuth, servicesConfig, progressBar)
}

type UploadConfiguration struct {
	Deb                   string
	Threads               int
	MinChecksumDeploySize int64
	Symlink               bool
	ExplodeArchive        bool
	Retries               int
}
