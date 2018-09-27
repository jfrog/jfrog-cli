package services

import (
	"errors"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/auth"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/httpclient"
	clientutils "github.com/jfrog/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
	"net/http"
	"path"
	"strconv"
	"strings"
)

const (
	MOVE MoveType = "move"
	COPY MoveType = "copy"
)

type MoveCopyService struct {
	moveType   MoveType
	client     *httpclient.HttpClient
	DryRun     bool
	ArtDetails auth.ArtifactoryDetails
}

func NewMoveCopyService(client *httpclient.HttpClient, moveType MoveType) *MoveCopyService {
	return &MoveCopyService{moveType: moveType, client: client}
}

func (mc *MoveCopyService) GetArtifactoryDetails() auth.ArtifactoryDetails {
	return mc.ArtDetails
}

func (mc *MoveCopyService) SetArtifactoryDetails(rt auth.ArtifactoryDetails) {
	mc.ArtDetails = rt
}

func (mc *MoveCopyService) IsDryRun() bool {
	return mc.DryRun
}

func (mc *MoveCopyService) GetJfrogHttpClient() *httpclient.HttpClient {
	return mc.client
}

func (mc *MoveCopyService) MoveCopyServiceMoveFilesWrapper(moveSpec MoveCopyParams) (successCount, failedCount int, err error) {
	switch moveSpec.GetSpecType() {
	case utils.WILDCARD, utils.SIMPLE:
		successCount, failedCount, err = mc.moveWildcard(moveSpec)
	case utils.AQL:
		successCount, failedCount, err = mc.moveAql(moveSpec)
	}
	if err != nil {
		return 0, 0, err
	}

	log.Debug(moveMsgs[mc.moveType].MovedMsg, strconv.Itoa(successCount), "artifacts.")
	if failedCount > 0 {
		err = errorutils.CheckError(errors.New("Failed " + moveMsgs[mc.moveType].MovingMsg + " " + strconv.Itoa(failedCount) + " artifacts."))
	}
	return
}

func (mc *MoveCopyService) moveAql(params MoveCopyParams) (successCount, failedCount int, err error) {
	log.Info("Searching artifacts...")
	resultItems, err := utils.AqlSearchBySpec(params.GetFile(), mc, utils.NONE)
	if err != nil {
		return
	}
	successCount, failedCount, err = mc.moveFiles("", resultItems, params)
	return
}

func (mc *MoveCopyService) moveWildcard(params MoveCopyParams) (successCount, failedCount int, err error) {
	log.Info("Searching artifacts...")
	params.SetIncludeDir(true)
	resultItems, err := utils.AqlSearchDefaultReturnFields(params.GetFile(), mc, utils.NONE)
	if err != nil {
		return
	}
	regexpPath := clientutils.PathToRegExp(params.GetFile().Pattern)
	successCount, failedCount, err = mc.moveFiles(regexpPath, resultItems, params)
	return
}

func reduceMovePaths(resultItems []utils.ResultItem, params MoveCopyParams) []utils.ResultItem {
	if params.IsFlat() {
		return utils.ReduceDirResult(resultItems, utils.FilterBottomChainResults)
	}
	return utils.ReduceDirResult(resultItems, utils.FilterTopChainResults)
}

func (mc *MoveCopyService) moveFiles(regexpPath string, resultItems []utils.ResultItem, params MoveCopyParams) (successCount, failedCount int, err error) {
	successCount = 0
	failedCount = 0
	resultItems = reduceMovePaths(resultItems, params)
	utils.LogSearchResults(len(resultItems))
	for _, v := range resultItems {
		destPathLocal := params.GetFile().Target
		if !params.IsFlat() {
			if strings.Contains(destPathLocal, "/") {
				file, dir := fileutils.GetFileAndDirFromPath(destPathLocal)
				destPathLocal = clientutils.TrimPath(dir + "/" + v.Path + "/" + file)
			} else {
				destPathLocal = clientutils.TrimPath(destPathLocal + "/" + v.Path + "/")
			}
		}
		destFile, err := clientutils.ReformatRegexp(regexpPath, v.GetItemRelativePath(), destPathLocal)
		if err != nil {
			log.Error(err)
			continue
		}
		if strings.HasSuffix(destFile, "/") {
			if v.Type != "folder" {
				destFile += v.Name
			} else {
				mc.createPathForMoveAction(destFile)
			}
		}
		success, err := mc.moveFile(v.GetItemRelativePath(), destFile)
		if err != nil {
			log.Error(err)
			continue
		}

		successCount += clientutils.Bool2Int(success)
		failedCount += clientutils.Bool2Int(!success)
	}
	return
}

func (mc *MoveCopyService) moveFile(sourcePath, destPath string) (bool, error) {
	message := moveMsgs[mc.moveType].MovingMsg + " artifact: " + sourcePath + " to: " + destPath
	moveUrl := mc.GetArtifactoryDetails().GetUrl()
	restApi := path.Join("api", string(mc.moveType), sourcePath)
	params := map[string]string{"to": destPath}
	if mc.IsDryRun() {
		log.Info("[Dry run]", message)
		params["dry"] = "1"
	} else {
		log.Info(message)
	}
	requestFullUrl, err := utils.BuildArtifactoryUrl(moveUrl, restApi, params)
	if err != nil {
		return false, err
	}
	httpClientsDetails := mc.GetArtifactoryDetails().CreateHttpClientDetails()

	resp, body, err := mc.client.SendPost(requestFullUrl, nil, httpClientsDetails)
	if err != nil {
		return false, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Error("Artifactory response: " + resp.Status + "\n" + clientutils.IndentJson(body))
	}

	log.Debug("Artifactory response:", resp.Status)
	return resp.StatusCode == http.StatusOK, nil
}

// Create destPath in Artifactory
func (mc *MoveCopyService) createPathForMoveAction(destPath string) (bool, error) {
	if mc.IsDryRun() == true {
		log.Info("[Dry run]", "Create path:", destPath)
		return true, nil
	}

	return mc.createPathInArtifactory(destPath, mc)
}

func (mc *MoveCopyService) createPathInArtifactory(destPath string, conf utils.CommonConf) (bool, error) {
	rtUrl := conf.GetArtifactoryDetails().GetUrl()
	requestFullUrl, err := utils.BuildArtifactoryUrl(rtUrl, destPath, map[string]string{})
	if err != nil {
		return false, err
	}
	httpClientsDetails := conf.GetArtifactoryDetails().CreateHttpClientDetails()
	resp, body, err := mc.client.SendPut(requestFullUrl, nil, httpClientsDetails)
	if err != nil {
		return false, err
	}

	if resp.StatusCode != http.StatusCreated {
		log.Error("Artifactory response: " + resp.Status + "\n" + clientutils.IndentJson(body))
	}

	log.Debug("Artifactory response:", resp.Status)
	return resp.StatusCode == http.StatusOK, nil
}

var moveMsgs = map[MoveType]MoveOptions{
	MOVE: {MovingMsg: "Moving", MovedMsg: "Moved"},
	COPY: {MovingMsg: "Copying", MovedMsg: "Copied"},
}

type MoveOptions struct {
	MovingMsg string
	MovedMsg  string
}

type MoveType string

type MoveCopyParams interface {
	utils.FileGetter
	GetFile() *utils.ArtifactoryCommonParams
	SetIncludeDir(bool)
	IsFlat() bool
}

type MoveCopyParamsImpl struct {
	*utils.ArtifactoryCommonParams
	Flat bool
}

func (mc *MoveCopyParamsImpl) GetFile() *utils.ArtifactoryCommonParams {
	return mc.ArtifactoryCommonParams
}

func (mc *MoveCopyParamsImpl) SetIncludeDir(isIncludeDir bool) {
	mc.GetFile().IncludeDirs = isIncludeDir
}

func (mc *MoveCopyParamsImpl) IsFlat() bool {
	return mc.Flat
}
