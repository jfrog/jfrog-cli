package utils

import (
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

func CreateDownloadServiceManager(artDetails *config.ArtifactoryDetails, flags *DownloadConfiguration,
	progressBar log.ProgressBar) (*artifactory.ArtifactoryServicesManager, error) {
	certPath, err := GetJfrogSecurityDir()
	if err != nil {
		return nil, err
	}
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
	if err != nil {
		return nil, err
	}
	return artifactory.NewWithProgress(&artAuth, servicesConfig, progressBar)
}

type DownloadConfiguration struct {
	Threads         int
	SplitCount      int
	MinSplitSize    int64
	BuildName       string
	BuildNumber     string
	DryRun          bool
	Symlink         bool
	ValidateSymlink bool
	ArtDetails      *config.ArtifactoryDetails
	Retries         int
}
