package utils

import (
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/utils/io"
)

func CreateDiagnosisServiceManager(artDetails *config.ArtifactoryDetails, threadCount int, dryRun bool, progressBar io.Progress) (*artifactory.ArtifactoryServicesManager, error) {
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
		SetDryRun(dryRun).
		SetCertificatesPath(certPath).
		SetInsecureTls(artDetails.InsecureTls).
		SetThreads(threadCount).
		Build()
	if err != nil {
		return nil, err
	}
	return artifactory.NewWithProgress(&artAuth, servicesConfig, progressBar)
}

