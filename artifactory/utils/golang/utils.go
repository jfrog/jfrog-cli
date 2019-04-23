package golang

import (
	"github.com/jfrog/gocmd/cmd"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

func LogGoVersion() error {
	output, err := cmd.GetGoVersion()
	if err != nil {
		return errorutils.CheckError(err)
	}
	log.Info("Using go:", output)
	return nil
}
