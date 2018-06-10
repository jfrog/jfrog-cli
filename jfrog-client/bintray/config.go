package bintray

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray/auth"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
)

type bintrayServicesConfig struct {
	auth.BintrayDetails
	dryRun            bool
	threads           int
	minSplitSize      int64
	splitCount        int
	minChecksumDeploy int64
	isDryRun          bool
	logger            log.Log
}

func (config *bintrayServicesConfig) GetUrl() string {
	return config.GetApiUrl()
}

func (config *bintrayServicesConfig) IsDryRun() bool {
	return config.isDryRun
}

func (config *bintrayServicesConfig) GetUser() string {
	return config.GetUser()
}

func (config *bintrayServicesConfig) GetKey() string {
	return config.GetKey()
}

func (config *bintrayServicesConfig) GetThreads() int {
	return config.threads
}

func (config *bintrayServicesConfig) GetMinSplitSize() int64 {
	return config.minSplitSize
}

func (config *bintrayServicesConfig) GetSplitCount() int {
	return config.splitCount
}
func (config *bintrayServicesConfig) GetMinChecksumDeploy() int64 {
	return config.minChecksumDeploy
}

func (config *bintrayServicesConfig) GetBintrayDetails() auth.BintrayDetails {
	return config.BintrayDetails
}

func (config *bintrayServicesConfig) GetLogger() log.Log {
	return config.logger
}
