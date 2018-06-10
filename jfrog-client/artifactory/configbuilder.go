package artifactory

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/auth"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
)

func NewConfigBuilder() *artifactoryServicesConfigBuilder {
	configBuilder := &artifactoryServicesConfigBuilder{}
	configBuilder.threads = 3
	configBuilder.minChecksumDeploy = 10240
	configBuilder.splitCount = 3
	configBuilder.minSplitSize = 5120
	return configBuilder
}

type artifactoryServicesConfigBuilder struct {
	auth.ArtifactoryDetails
	certifactesPath   string
	threads           int
	minSplitSize      int64
	splitCount        int
	minChecksumDeploy int64
	isDryRun          bool
	logger            log.Log
}

func (builder *artifactoryServicesConfigBuilder) SetArtDetails(artDetails auth.ArtifactoryDetails) *artifactoryServicesConfigBuilder {
	builder.ArtifactoryDetails = artDetails
	return builder
}

func (builder *artifactoryServicesConfigBuilder) SetCertificatesPath(certificatesPath string) *artifactoryServicesConfigBuilder {
	builder.certifactesPath = certificatesPath
	return builder
}

func (builder *artifactoryServicesConfigBuilder) SetThreads(threads int) *artifactoryServicesConfigBuilder {
	builder.threads = threads
	return builder
}

func (builder *artifactoryServicesConfigBuilder) SetMinSplitSize(splitSize int64) *artifactoryServicesConfigBuilder {
	builder.minSplitSize = splitSize
	return builder
}

func (builder *artifactoryServicesConfigBuilder) SetSplitCount(splitCount int) *artifactoryServicesConfigBuilder {
	builder.splitCount = splitCount
	return builder
}

func (builder *artifactoryServicesConfigBuilder) SetMinChecksumDeploy(minChecksumDeploy int64) *artifactoryServicesConfigBuilder {
	builder.minChecksumDeploy = minChecksumDeploy
	return builder
}

func (builder *artifactoryServicesConfigBuilder) SetDryRun(dryRun bool) *artifactoryServicesConfigBuilder {
	builder.isDryRun = dryRun
	return builder
}

func (builder *artifactoryServicesConfigBuilder) Build() (Config, error) {
	c := &artifactoryServicesConfig{}
	c.ArtifactoryDetails = builder.ArtifactoryDetails
	c.threads = builder.threads
	c.minChecksumDeploy = builder.minChecksumDeploy
	c.minSplitSize = builder.minSplitSize
	c.splitCount = builder.splitCount
	c.logger = builder.logger
	c.certifactesPath = builder.certifactesPath
	c.dryRun = builder.isDryRun
	return c, nil
}

func (builder *artifactoryServicesConfigBuilder) SetLogger(logger log.Log) *artifactoryServicesConfigBuilder {
	builder.logger = logger
	return builder
}
