package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
	"github.com/jfrogdev/jfrog-cli-go/utils/io/httputils"
	"github.com/jfrogdev/jfrog-cli-go/utils/io/fileutils"
	"encoding/json"
	"errors"
	"path/filepath"
)

func BuildAddArtifact(buildName, buildNumber, artifact string, flags *BuildAddArtifactFlags) (err error) {
	log.Info("Adding artifact '" + artifact + "' to " + buildName + "#" + buildNumber)
	if err = utils.SaveBuildGeneralDetails(buildName, buildNumber); err != nil {
		return
	}

	fileName, details, err := getFileInfo(artifact, flags)
	if err != nil {
		return err
	}

	populateFunc := func(wrapper *utils.ArtifactBuildInfoWrapper) {
		wrapper.Artifacts = []utils.ArtifactsBuildInfo{utils.CreateArtifactsBuildInfo(fileName, details)}
	}
	err = utils.SavePartialBuildInfo(buildName, buildNumber, populateFunc)

	return
}

func getFileInfo(artifact string, flags *BuildAddArtifactFlags) (fileName string, details *fileutils.FileDetails, err error) {
	apiUrl := flags.ArtDetails.Url + "api/storage/" + artifact
	log.Debug("Retrieving file info through URL: " + apiUrl)
	httpClientsDetails := utils.GetArtifactoryHttpClientDetails(flags.ArtDetails)
	resp, body, _, err := httputils.SendGet(apiUrl, true, httpClientsDetails)
	if err != nil {
		return "", nil, err
	}
	if resp.StatusCode != 200 {
		return "", nil, cliutils.CheckError(errors.New("Error retrieving file info; Artifactory response: " + resp.Status + "\n" + cliutils.IndentJson(body)))
	}

	var result fileInfoResult
	err = json.Unmarshal(body, &result)
	if cliutils.CheckError(err) != nil {
		return "", nil, err
	}

	if len(result.Children) > 0 {
		return "", nil, cliutils.CheckError(errors.New("Artifact has children; this is currently not supported"))
	}

	fileName = filepath.Base(result.Path)
	details = new(fileutils.FileDetails)
	details.Checksum = result.Checksums
	details.Size = int64(0)

	log.Debug("Successfully retrieved file info:\n" +
						"  fileName      = " + fileName + "\n" +
						"  Sha1 checksum = " + details.Checksum.Sha1 + "\n" +
						"  Md5 checksum  = " + details.Checksum.Md5)
	return
}

type fileInfoResult struct {
	Path        string
	Checksums   fileutils.ChecksumDetails
	Children    []map[string]interface{}
}

type BuildAddArtifactFlags struct {
	ArtDetails  *config.ArtifactoryDetails
}
