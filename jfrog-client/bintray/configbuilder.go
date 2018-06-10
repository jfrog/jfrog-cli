package bintray

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray/auth"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
)

func NewConfigBuilder() *bintrayServicesConfigBuilder {
	configBuilder := &bintrayServicesConfigBuilder{}
	configBuilder.threads = 3
	configBuilder.minChecksumDeploy = 10240
	configBuilder.splitCount = 3
	configBuilder.minSplitSize = 5120
	return configBuilder
}

type bintrayServicesConfigBuilder struct {
	auth.BintrayDetails
	threads           int
	minSplitSize      int64
	splitCount        int
	minChecksumDeploy int64
	isDryRun          bool
	logger            log.Log
}

func (builder *bintrayServicesConfigBuilder) SetBintrayDetails(artDetails auth.BintrayDetails) *bintrayServicesConfigBuilder {
	builder.BintrayDetails = artDetails
	return builder
}

func (builder *bintrayServicesConfigBuilder) SetThreads(threads int) *bintrayServicesConfigBuilder {
	builder.threads = threads
	return builder
}

func (builder *bintrayServicesConfigBuilder) SetMinSplitSize(splitSize int64) *bintrayServicesConfigBuilder {
	builder.minSplitSize = splitSize
	return builder
}

func (builder *bintrayServicesConfigBuilder) SetSplitCount(splitCount int) *bintrayServicesConfigBuilder {
	builder.splitCount = splitCount
	return builder
}

func (builder *bintrayServicesConfigBuilder) SetMinChecksumDeploy(minChecksumDeploy int64) *bintrayServicesConfigBuilder {
	builder.minChecksumDeploy = minChecksumDeploy
	return builder
}

func (builder *bintrayServicesConfigBuilder) SetDryRun(dryRun bool) *bintrayServicesConfigBuilder {
	builder.isDryRun = dryRun
	return builder
}

func (builder *bintrayServicesConfigBuilder) Build() Config {
	c := &bintrayServicesConfig{}
	c.BintrayDetails = builder.BintrayDetails
	c.threads = builder.threads
	c.minChecksumDeploy = builder.minChecksumDeploy
	c.minSplitSize = builder.minSplitSize
	c.splitCount = builder.splitCount
	c.logger = builder.logger
	return c
}

func (builder *bintrayServicesConfigBuilder) SetLogger(logger log.Log) *bintrayServicesConfigBuilder {
	builder.logger = logger
	return builder
}
