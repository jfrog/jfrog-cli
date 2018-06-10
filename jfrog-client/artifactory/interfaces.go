package artifactory

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/auth"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
)

type Config interface {
	GetUrl() string
	GetPassword() string
	GetApiKey() string
	GetCertifactesPath() string
	GetThreads() int
	GetMinSplitSize() int64
	GetSplitCount() int
	GetMinChecksumDeploy() int64
	IsDryRun() bool
	GetArtDetails() auth.ArtifactoryDetails
	GetLogger() log.Log
}

type ArtifactoryServicesSetter interface {
	SetThread(threads int)
	SetArtDetails(artDetails auth.ArtifactoryDetails)
	SetDryRun(isDryRun bool)
}
