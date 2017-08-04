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

func BuildAddArtifact(buildName, buildNumber, artifactPath string, flags *BuildAddArtifactFlags) (err error) {
	log.Info("Adding artifact '" + artifactPath + "' to build info " + buildName + " #" + buildNumber)
	if err = utils.SaveBuildGeneralDetails(buildName, buildNumber); err != nil {
		return
	}

	fileName, details, err := getFileInfo(artifactPath, flags); if err != nil {
		return
	}

	err = setBuildNameAndNumber(buildName, buildNumber, artifactPath, flags); if err != nil {
		return
	}

	populateFunc := func(wrapper *utils.ArtifactBuildInfoWrapper) {
		wrapper.Artifacts = []utils.ArtifactsBuildInfo{utils.CreateArtifactsBuildInfo(fileName, details)}
	}

	err = utils.SavePartialBuildInfo(buildName, buildNumber, populateFunc); if err != nil {
		return
	}

	log.Info("Successfully added artifact to build info")

	return
}

func getFileInfo(artifactPath string, flags *BuildAddArtifactFlags) (fileName string, details *fileutils.FileDetails, err error) {
	getFileInfoUrl := flags.ArtDetails.Url + "api/storage/" + artifactPath
	log.Debug("Retrieving file info through URL: " + getFileInfoUrl)
	httpClientsDetails := utils.GetArtifactoryHttpClientDetails(flags.ArtDetails)
	resp, body, _, err := httputils.SendGet(getFileInfoUrl, true, httpClientsDetails)
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

func setBuildNameAndNumber(buildName, buildNumber, artifactPath string, flags *BuildAddArtifactFlags) error {
	log.Info("Setting build.name and build.number properties")

	// Not all Artifactory versions (e.g. 4.7.5 rev 40176) seem to support setting multiple
	// properties at once. Hence setting build.number and build.name separately.
	err := utils.SetProps(artifactPath, "build.name=" + buildName, flags.ArtDetails); if err != nil {
		return err
	}
	err = utils.SetProps(artifactPath, "build.number=" + buildNumber, flags.ArtDetails); if err != nil {
		return err
	}
	return nil
}

type fileInfoResult struct {
	Path        string
	Checksums   fileutils.ChecksumDetails
	Children    []map[string]interface{}
}

type BuildAddArtifactFlags struct {
	ArtDetails  *config.ArtifactoryDetails
}
