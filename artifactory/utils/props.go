package utils

import (
  "github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/io/httputils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
	"github.com/jfrogdev/jfrog-cli-go/errors/httperrors"
)

func SetProps(artifactPath string, props string, artDetails *config.ArtifactoryDetails) error {
	encodedParam, err := EncodeParams(props)
	if err != nil {
		return err
	}
	setPropertiesUrl := artDetails.Url + "api/storage/" + artifactPath + "?properties=" + encodedParam
	httpClientsDetails := GetArtifactoryHttpClientDetails(artDetails)
	log.Debug("Sending set properties request:", setPropertiesUrl)
	resp, body, err := httputils.SendPut(setPropertiesUrl, nil, httpClientsDetails)
	if err != nil {
		return err
	}
	if err = httperrors.CheckResponseStatus(resp, body, 204); err != nil {
		return err
	}
	log.Debug("Successfully set properties")
	return nil
}