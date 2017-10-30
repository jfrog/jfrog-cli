package artifactory

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils/auth"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
)

type ArtifactoryServicesConfigBuilder struct {
	*auth.ArtifactoryDetails
	certifactesPath   string
	threads           int
	minSplitSize      int64
	splitCount        int
	minChecksumDeploy int64
	isDryRun          bool
	logger            log.Log
}

func (builder *ArtifactoryServicesConfigBuilder) SetArtDetails(artDetails *auth.ArtifactoryDetails) *ArtifactoryServicesConfigBuilder {
	builder.ArtifactoryDetails = artDetails
	return builder
}

func (builder *ArtifactoryServicesConfigBuilder) SetCertificatesPath(certificatesPath string) *ArtifactoryServicesConfigBuilder {
	builder.certifactesPath = certificatesPath
	return builder
}

func (builder *ArtifactoryServicesConfigBuilder) SetNumOfThreadPerOperation(threads int) *ArtifactoryServicesConfigBuilder {
	builder.threads = threads
	return builder
}

func (builder *ArtifactoryServicesConfigBuilder) SetMinSplitSize(splitSize int64) *ArtifactoryServicesConfigBuilder {
	builder.minSplitSize = splitSize
	return builder
}

func (builder *ArtifactoryServicesConfigBuilder) SetSplitCount(splitCount int) *ArtifactoryServicesConfigBuilder {
	builder.splitCount = splitCount
	return builder
}

func (builder *ArtifactoryServicesConfigBuilder) SetMinChecksumDeploy(minChecksumDeploy int64) *ArtifactoryServicesConfigBuilder {
	builder.minChecksumDeploy = minChecksumDeploy
	return builder
}

func (builder *ArtifactoryServicesConfigBuilder) SetDryRun(dryRun bool) *ArtifactoryServicesConfigBuilder {
	builder.isDryRun = dryRun
	return builder
}

func (builder *ArtifactoryServicesConfigBuilder) Build() (ArtifactoryConfig, error) {
	c := &artifactoryServicesConfig{}
	c.ArtifactoryDetails = builder.ArtifactoryDetails

	if builder.threads == 0 {
		c.threads = 3
	} else {
		c.threads = builder.threads
	}

	if builder.minChecksumDeploy == 0 {
		c.minChecksumDeploy = 10240
	} else {
		c.minChecksumDeploy = builder.minChecksumDeploy
	}

	c.minSplitSize = builder.minSplitSize
	c.splitCount = builder.splitCount
	c.logger = builder.logger
	c.certifactesPath = builder.certifactesPath
	c.dryRun = builder.isDryRun
	return c, nil
}

func (builder *ArtifactoryServicesConfigBuilder) SetLogger(logger log.Log) *ArtifactoryServicesConfigBuilder {
	builder.logger = logger
	return builder
}
