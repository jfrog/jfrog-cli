package commands

import (
	"github.com/jfrog/jfrog-client-go/bintray"
	"github.com/jfrog/jfrog-client-go/bintray/services/mavensync"
	"github.com/jfrog/jfrog-client-go/bintray/services/versions"
)

func MavenCentralSync(config bintray.Config, params *mavensync.Params, path *versions.Path) error {
	sm, err := bintray.New(config)
	if err != nil {
		return err
	}
	return sm.MavenCentralContentSync(params, path)
}
