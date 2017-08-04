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

	// artifactsBuildInfo := utils.CreateArtifactsBuildInfo(fileName, details)


	// apiUrl := flags.ArtDetails.Url + "api/storage/" + artifact
	// httpClientsDetails := utils.GetArtifactoryHttpClientDetails(flags.ArtDetails)
	// resp, body, _, err := httputils.SendGet(apiUrl, true, httpClientsDetails)
	// if err != nil {
	// 	return err
	// }
	// if resp.StatusCode != 200 {
	// 	return cliutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + cliutils.IndentJson(body)))
	// }
	
	// log.Info("RETURN BODY: " + string(body))

	// var result FileInfoResult
	// err = json.Unmarshal(body, &result)
	// if cliutils.CheckError(err) != nil {
	// 	return err
	// }

	// if len(result.Children) > 0 {
	// 	return cliutils.CheckError(errors.New("Artifact has children: this is currently not supported"))
	// }

	// ext := filepath.Ext(result.Path)
	// log.Info("Type: " + ext[1:len(ext)])
	// log.Info("Name: " + filepath.Base(result.Path))
	// log.Info("Sha1: " + result.Checksums.Sha1)
	// log.Info("Md5: " + result.Checksums.Md5)

	

// "artifacts" : [ {
//       "type" : "jar",
//       "sha1" : "2ed52ad1d864aec00d7a2394e99b3abca6d16639",
//       "md5" : "844920070d81271b69579e17ddc6715c",
//       "name" : "multi2-4.2-SNAPSHOT.jar"
//     }

	// log.Info("Deploying build info...")
	// resp, body, err := utils.PublishBuildInfo(flags.ArtDetails.Url, marshaledBuildInfo, httpClientsDetails)

	// 	err := initTransport(artifactoryDetails)
	// if err != nil {
	// 	return nil, "", err
	// }
	// apiUrl := artifactoryDetails.Url + "api/security/encryptedPassword"
	// httpClientsDetails := GetArtifactoryHttpClientDetails(artifactoryDetails)
	// resp, body, _, err := httputils.SendGet(apiUrl, true, httpClientsDetails)
	// return resp, string(body), err

	// TODO:
	//
	// Decision to make
	// A)
	// Every artifact will be considered a Module (in terms of the build-upload API; see
	// https://www.jfrog.com/confluence/display/RTF/Artifactory+REST+API#ArtifactoryRESTAPI-BuildUpload).
	//
	// Depending on whether the artifact refers to exactly one file or a folder containing file, the
	// module will have either one or multiple artifacts.
	//
	//  If artifact refers to one file:
	//    Get file info for artifact
	//  Else in case artifact refers to a folder with only non-folder children:
	//    Get file info for all child artifacts
	//  Else:
	//    Exit with error
	// 
	// B)
	// For starters, just add all individual artifacts as the upload command does. This will
	// result in all artifacts being part of the "default" module and thus would not be entirely
	// correct. However, it will allow us to make progress faster without having to modify
	// buildpublish.go. Splitting things up in modules can be added in a next step.
	//
	// For now we go with B:
	//
	// 1) Artifact should refer to a single file.
	// 2) Set item properties (i.a. build.number and build.name) recursively on artifact
	// 3) Save all artifacts as partial build info to add them to local BOM.

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
