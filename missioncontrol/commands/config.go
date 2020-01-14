package commands

import (
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-cli-go/utils/ioutils"
	"github.com/jfrog/jfrog-cli-go/utils/lock"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"golang.org/x/crypto/ssh/terminal"
	"net/url"
	"sync"
	"syscall"
)

// Internal golang locking for the same process.
var mutex sync.Mutex

func GetConfig() (*config.MissionControlDetails, error) {
	return config.ReadMissionControlConf()
}

func ShowConfig() error {
	details, err := config.ReadMissionControlConf()
	if err != nil {
		return err
	}
	if details.Url != "" {
		log.Output("Url: " + details.Url)
	}
	if details.AccessToken != "" {
		log.Output("AccessToken: ***")
	}
	return nil
}

func ClearConfig() error {
	return config.SaveMissionControlConf(new(config.MissionControlDetails))
}

func Config(details, defaultDetails *config.MissionControlDetails, interactive bool) (conf *config.MissionControlDetails, err error) {
	mutex.Lock()
	lockFile, err := lock.CreateLock()
	defer mutex.Unlock()
	defer lockFile.Unlock()

	if err != nil {
		return nil, err
	}
	conf = details
	if conf == nil {
		conf = new(config.MissionControlDetails)
	}
	if interactive {
		if defaultDetails == nil {
			defaultDetails, err = config.ReadMissionControlConf()
			if err != nil {
				return
			}
		}
		if conf.Url == "" {
			ioutils.ScanFromConsole("Mission Control URL", &conf.Url, defaultDetails.Url)
			var u *url.URL
			u, err = url.Parse(conf.Url)
			err = errorutils.CheckError(err)
			if err != nil {
				return
			}
			if u.Scheme != "http" && u.Scheme != "https" {
				err = errorutils.CheckError(errors.New("URL scheme is not valid " + u.Scheme))
				if err != nil {
					return
				}
			}
		}
		if conf.AccessToken == "" {
			print("Access token: ")
			byteToken, err := terminal.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return nil, errorutils.CheckError(err)
			}
			// New-line required after the access token input:
			fmt.Println()
			conf.SetAccessToken(string(byteToken))
		}
	}
	conf.Url = utils.AddTrailingSlashIfNeeded(conf.Url)
	err = config.SaveMissionControlConf(conf)
	return
}

type ConfigFlags struct {
	MissionControlDetails *config.MissionControlDetails
	Interactive           bool
}
