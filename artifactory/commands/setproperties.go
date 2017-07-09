package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/io/httputils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
	"fmt"
)

func SetProps(spec *utils.SpecFiles, flags utils.CommonFlags, props string) error {
	err := utils.PreCommandSetup(flags)
	if err != nil {
		return err
	}
	resultItems, err := utils.SearchBySpecFiles(spec, flags)
	if err != nil {
		return err
	}
	updatePropertiesBaseUrl := flags.GetArtifactoryDetails().Url + "api/storage"
	log.Info("Setting properties...")
	for _, item := range resultItems {
		log.Info("Setting properties to:", item.GetFullUrl())
		httpClientsDetails := utils.GetArtifactoryHttpClientDetails(flags.GetArtifactoryDetails())
		encodedParam, err := utils.EncodeParams(props)
		if err != nil {
			return err
		}
		setPropertiesUrl := updatePropertiesBaseUrl + "/" + item.GetFullUrl() + "?properties=" + encodedParam
		log.Debug("Sending set properties request:", setPropertiesUrl)
		resp, _, err := httputils.SendPut(setPropertiesUrl, nil, httpClientsDetails)
		if err != nil {
			return err
		}
		if resp.StatusCode != 204 {
			return fmt.Errorf("Coldn't set properties for: %v", item.GetFullUrl())
		}
	}

	log.Info("Done setting properties...")
	return err
}