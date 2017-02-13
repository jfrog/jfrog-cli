package commands


import (
	"github.com/jfrogdev/jfrog-cli-go/utils/io/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"errors"
	"fmt"
	"net/url"
)

func GetConfig() (*config.MissionControlDetails, error) {
	return config.ReadMissionControlConf()
}

func ShowConfig() error {
	details, err := config.ReadMissionControlConf()
	if err != nil {
	    return err
	}
	if details.Url != "" {
		fmt.Println("Url: " + details.Url)
	}
	if details.User != "" {
		fmt.Println("User: " + details.User)
	}
	if details.Password != "" {
		fmt.Println("Password: ***")
	}
	return nil
}

func ClearConfig() {
	config.SaveMissionControlConf(new(config.MissionControlDetails))
}

func Config(details, defaultDetails *config.MissionControlDetails, interactive bool) (conf *config.MissionControlDetails, err error) {
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
			u, err = url.Parse(conf.Url);
			err = cliutils.CheckError(err)
			if err != nil {
			    return
			}
			if u.Scheme != "http" && u.Scheme != "https" {
				err = cliutils.CheckError(errors.New("URL scheme is not valid " + u.Scheme))
                if err != nil {
                    return
                }
			}
		}
		ioutils.ReadCredentialsFromConsole(conf, defaultDetails)
	}
	conf.Url = cliutils.AddTrailingSlashIfNeeded(conf.Url)
	config.SaveMissionControlConf(conf)
	return
}

type ConfigFlags struct {
	MissionControlDetails *config.MissionControlDetails
	Interactive           bool
}