package services

import (
	"encoding/json"
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/auth"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/httpclient"
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
	rtclientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
)

type getFileInfoService struct {
	client     *httpclient.HttpClient
	ArtDetails auth.ArtifactoryDetails
}

func NewFileInfoService(client *httpclient.HttpClient) *getFileInfoService {
	return &getFileInfoService{client: client}
}

func (gfi *getFileInfoService) GetArtifactoryDetails() auth.ArtifactoryDetails {
	return gfi.ArtDetails
}

func (gfi *getFileInfoService) SetArtifactoryDetails(rt auth.ArtifactoryDetails) {
	gfi.ArtDetails = rt
}

func (gfi *getFileInfoService) GetFileInfo(path string) (fileInfo *rtclientutils.FileInfo, err error) {
	getFileInfoUrl := gfi.ArtDetails.GetUrl() + "api/storage/" + path
	log.Debug("Retrieving file info through URL: " + getFileInfoUrl)
	httpClientsDetails := gfi.GetArtifactoryDetails().CreateHttpClientDetails()
	resp, body, _, err := gfi.client.SendGet(getFileInfoUrl, true, httpClientsDetails); if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errorutils.CheckError(errors.New("Error retrieving file info; Artifactory response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Artifactory response:", resp.Status)

	var result fileInfoResult
	err = json.Unmarshal(body, &result)
	if errorutils.CheckError(err) != nil {
		return nil, err
	}

	if len(result.Children) > 0 {
		return nil, errorutils.CheckError(errors.New("Artifact has children; this is currently not supported"))
	}

	fileInfo = new(rtclientutils.FileInfo)
	fileInfo.ArtifactoryPath = path
	fileInfo.FileHashes = new(rtclientutils.FileHashes)
	*fileInfo.FileHashes = result.Checksums

	log.Debug("Successfully retrieved file info:\n" +
	          "  path            = " + fileInfo.ArtifactoryPath + "\n" +
	          "  Sha1 checksum   = " + fileInfo.FileHashes.Sha1 + "\n" +
	          "  Sha256 checksum = " + fileInfo.FileHashes.Sha256 + "\n" +
	          "  Md5 checksum    = " + fileInfo.FileHashes.Md5)
	return
}

type fileInfoResult struct {
	Path        string
	Checksums   rtclientutils.FileHashes
	Children    []map[string]interface{}
}
