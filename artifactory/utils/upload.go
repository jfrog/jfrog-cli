package utils

import (
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

func CreateUploadServiceManager(artDetails *config.ArtifactoryDetails, flags *UploadConfiguration, certPath string, progressBar log.ProgressBar) (*artifactory.ArtifactoryServicesManager, error) {
	artAuth, err := artDetails.CreateArtAuthConfig()
	if err != nil {
		return nil, err
	}
	servicesConfig, err := artifactory.NewConfigBuilder().
		SetArtDetails(artAuth).
		SetDryRun(flags.DryRun).
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
	BuildName             string
	BuildNumber           string
	DryRun                bool
	Symlink               bool
	ExplodeArchive        bool
	ArtDetails            *config.ArtifactoryDetails
	Retries               int
}
