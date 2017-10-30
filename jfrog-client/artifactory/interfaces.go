package artifactory

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils/auth"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
)

type ArtifactoryConfig interface {
	GetUrl() string
	GetPassword() string
	GetApiKey() string
	GetCertifactesPath() string
	GetNumOfThreadPerOperation() int
	GetMinSplitSize() int64
	GetSplitCount() int
	GetMinChecksumDeploy() int64
	IsDryRun() bool
	GetArtDetails() *auth.ArtifactoryDetails
	GetLogger() log.Log
}

type ArtifactoryServicesSetter interface {
	SetThread(threads int)
	SetArtDetails(artDetails *auth.ArtifactoryDetails)
	SetDryRun(isDryRun bool)
}
