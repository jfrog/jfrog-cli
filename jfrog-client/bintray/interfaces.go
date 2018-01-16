package bintray

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/bintray/auth"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
)

type Config interface {
	GetUrl() string
	GetKey() string
	GetThreads() int
	GetMinSplitSize() int64
	GetSplitCount() int
	GetMinChecksumDeploy() int64
	IsDryRun() bool
	GetBintrayDetails() auth.BintrayDetails
	GetLogger() log.Log
}
