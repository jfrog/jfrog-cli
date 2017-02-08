package utils

import (
	"strings"
	"strconv"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
	"errors"
)

const (
	MOVE MoveType = "move"
	COPY MoveType = "copy"
)

func MoveFilesWrapper(moveSpec *SpecFiles, flags *MoveFlags, moveType MoveType) (err error) {
	err = PreCommandSetup(flags)
	if err != nil {
		return
	}

	var successCount int
	var failedCount int

	for i := 0; i < len(moveSpec.Files); i++ {
		var successPartial, failedPartial int
		switch moveSpec.Get(i).GetSpecType() {
		case WILDCARD:
			successPartial, failedPartial, err = moveWildcard(moveSpec.Get(i), flags, moveType)
		case SIMPLE:
			successPartial, failedPartial, err = moveSimple(moveSpec.Get(i), flags, moveType)
		case AQL:
			successPartial, failedPartial, err = moveAql(moveSpec.Get(i), flags, moveType)
		}
		successCount += successPartial
		failedCount += failedPartial
		if err != nil {
			return
		}
	}

	log.Info(moveMsgs[moveType].MovedMsg, strconv.Itoa(successCount), "artifacts.")
	if failedCount > 0 {
		err = cliutils.CheckError(errors.New("Failed " + moveMsgs[moveType].MovingMsg + " " +strconv.Itoa(failedCount) + " artifacts."))
	}

	return
}

func moveAql(fileSpec *File, flags *MoveFlags, moveType MoveType) (successCount, failedCount int, err error) {
	log.Info("Searching artifacts...")
	resultItems, err := AqlSearchBySpec(fileSpec, flags)
	if err != nil {
		return
	}
	LogSearchResults(len(resultItems))
	successCount, failedCount, err = moveFiles("", resultItems, fileSpec, flags, moveType)
	return
}

func moveWildcard(fileSpec *File, flags *MoveFlags, moveType MoveType) (successCount, failedCount int, err error) {
	log.Info("Searching artifacts...")
	resultItems, err := AqlSearchDefaultReturnFields(fileSpec, flags)
	if err != nil {
		return
	}
	LogSearchResults(len(resultItems))
	regexpPath := cliutils.PathToRegExp(fileSpec.Pattern)
	successCount, failedCount, err = moveFiles(regexpPath, resultItems, fileSpec, flags, moveType)
	return
}

func moveSimple(fileSpec *File, flags *MoveFlags, moveType MoveType) (successCount, failedCount int, err error) {

	cleanPattern := cliutils.StripChars(fileSpec.Pattern, "()")
	patternFileName, _ := ioutils.GetFileAndDirFromPath(fileSpec.Pattern)

	regexpPattern := cliutils.PathToRegExp(fileSpec.Pattern)
	placeHolderTarget, err := cliutils.ReformatRegexp(regexpPattern, cleanPattern, fileSpec.Target)
	if err != nil {
		return
	}

	if strings.HasSuffix(placeHolderTarget, "/") {
		placeHolderTarget += patternFileName
	}
	success, err := moveFile(cleanPattern, placeHolderTarget, flags, moveType)
	successCount = cliutils.Bool2Int(success)
	failedCount = cliutils.Bool2Int(!success)
	return
}

func moveFiles(regexpPath string, resultItems []AqlSearchResultItem, fileSpec *File, flags *MoveFlags, moveType MoveType) (successCount, failedCount int, err error) {
	successCount = 0
	failedCount = 0

	for _, v := range resultItems {
		destPathLocal := fileSpec.Target
		isFlat, e := cliutils.StringToBool(fileSpec.Flat, false)
		if e != nil {
			err = e
			return
		}
		if !isFlat {
			if strings.Contains(destPathLocal, "/") {
				file, dir := ioutils.GetFileAndDirFromPath(destPathLocal)
				destPathLocal = cliutils.TrimPath(dir + "/" + v.Path + "/" + file)
			} else {
				destPathLocal = cliutils.TrimPath(destPathLocal + "/" + v.Path + "/")
			}
		}
		destFile, e := cliutils.ReformatRegexp(regexpPath, v.GetFullUrl(), destPathLocal)
		if e != nil {
			err = e
			return
		}
		if strings.HasSuffix(destFile, "/") {
			destFile += v.Name
		}
		success, e := moveFile(v.GetFullUrl(), destFile, flags, moveType)
		if e != nil {
			err = e
			return
		}

		successCount += cliutils.Bool2Int(success)
		failedCount += cliutils.Bool2Int(!success)
	}
	return
}

func moveFile(sourcePath, destPath string, flags *MoveFlags, moveType MoveType) (bool, error) {
	message := moveMsgs[moveType].MovingMsg + " artifact: " + sourcePath + " to: " + destPath
	if flags.DryRun == true {
		log.Info("[Dry run] ", message)
		return true, nil
	}

	log.Info(message)

	moveUrl := flags.ArtDetails.Url
	restApi := "api/" + string(moveType) + "/" + sourcePath
	requestFullUrl, err := BuildArtifactoryUrl(moveUrl, restApi, map[string]string{"to": destPath})
	if err != nil {
		return false, err
	}
	httpClientsDetails := GetArtifactoryHttpClientDetails(flags.ArtDetails)
	resp, body, err := ioutils.SendPost(requestFullUrl, nil, httpClientsDetails)
	if err != nil {
		return false, err
	}

	if resp.StatusCode != 200 {
		log.Error("Artifactory response: " + resp.Status + "\n" + cliutils.IndentJson(body))
	}

	log.Debug("Artifactory response:", resp.Status)
	return resp.StatusCode == 200, nil
}

var moveMsgs = map[MoveType]MoveOptions{
	MOVE: MoveOptions{MovingMsg: "Moving", MovedMsg: "Moved"},
	COPY: MoveOptions{MovingMsg: "Copying", MovedMsg: "Copied"},
}

type MoveOptions struct {
	MovingMsg string
	MovedMsg  string
}

type MoveType string

type MoveFlags struct {
	DryRun     bool
	ArtDetails *config.ArtifactoryDetails
}

func (flags *MoveFlags) GetArtifactoryDetails() *config.ArtifactoryDetails {
	return flags.ArtDetails
}

func (flags *MoveFlags) IsDryRun() bool {
	return flags.DryRun
}