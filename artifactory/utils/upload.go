package utils

import (
	"github.com/jfrog/jfrog-cli-go/utils/config"
	clientConfig "github.com/jfrog/jfrog-client-go/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/utils/io"
)

func CreateUploadServiceManager(artDetails *config.ArtifactoryDetails, flags *UploadConfiguration, certPath string, dryRun bool, progressBar io.Progress) (*artifactory.ArtifactoryServicesManager, error) {
	artAuth, err := artDetails.CreateArtAuthConfig()
	if err != nil {
		return nil, err
	}
	servicesConfig, err := clientConfig.NewConfigBuilder().
		SetArtDetails(artAuth).
		SetDryRun(dryRun).
		SetCertificatesPath(certPath).
		SetInsecureTls(artDetails.InsecureTls).
		SetThreads(flags.Threads).
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
