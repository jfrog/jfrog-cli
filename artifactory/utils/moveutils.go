package utils

import (
	"strings"
	"fmt"
	"strconv"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/logger"
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

	for i := 0; i < len(moveSpec.Files); i++ {
		switch moveSpec.Get(i).GetSpecType() {
		case WILDCARD:
			err = moveWildcard(moveSpec.Get(i), flags, moveType)
		case SIMPLE:
			err = moveSimple(moveSpec.Get(i), flags, moveType)
		case AQL:
			err = moveAql(moveSpec.Get(i), flags, moveType)
		}
		if err != nil {
			return
		}
	}
	return
}

func moveAql(fileSpec *Files, flags *MoveFlags, moveType MoveType) error {
	resultItems, err := AqlSearchBySpec(fileSpec.Aql, flags)
	if err != nil {
		return err
	}
	return moveFiles("", resultItems, fileSpec, flags, moveType)
}

func moveWildcard(fileSpec *Files, flags *MoveFlags, moveType MoveType) error {
	isRecursive, err := cliutils.StringToBool(fileSpec.Recursive, true)
	if err != nil {
		return err
	}
	resultItems, err := AqlSearchDefaultReturnFields(fileSpec.Pattern, isRecursive, fileSpec.Props, flags)
	if err != nil {
		return err
	}
	regexpPath := cliutils.PathToRegExp(fileSpec.Pattern)
	return moveFiles(regexpPath, resultItems, fileSpec, flags, moveType)
}

func moveSimple(fileSpec *Files, flags *MoveFlags, moveType MoveType) error {

	cleanPattern := cliutils.StripChars(fileSpec.Pattern, "()")
	patternFileName, _ := ioutils.GetFileAndDirFromPath(fileSpec.Pattern)

	regexpPattern := cliutils.PathToRegExp(fileSpec.Pattern)
	placeHolderTarget, err := cliutils.ReformatRegexp(regexpPattern, cleanPattern, fileSpec.Target)
	if err != nil {
		return err
	}

	if strings.HasSuffix(placeHolderTarget, "/") {
		placeHolderTarget += patternFileName
	}
	_, err = moveFile(cleanPattern, placeHolderTarget, flags, moveType)
	return err
}

func moveFiles(regexpPath string, resultItems []AqlSearchResultItem, fileSpec *Files, flags *MoveFlags, moveType MoveType) error {
	movedCount := 0

	for _, v := range resultItems {
		destPathLocal := fileSpec.Target
		isFlat, err := cliutils.StringToBool(fileSpec.Flat, false)
		if err != nil {
			return err
		}
		if !isFlat {
			if strings.Contains(destPathLocal, "/") {
				file, dir := ioutils.GetFileAndDirFromPath(destPathLocal)
				destPathLocal = cliutils.TrimPath(dir + "/" + v.Path + "/" + file)
			} else {
				destPathLocal = cliutils.TrimPath(destPathLocal + "/" + v.Path + "/")
			}
		}
		destFile, err := cliutils.ReformatRegexp(regexpPath, v.GetFullUrl(), destPathLocal)
		if err != nil {
			return err
		}
		if strings.HasSuffix(destFile, "/") {
			destFile += v.Name
		}
		success, err := moveFile(v.GetFullUrl(), destFile, flags, moveType)
		if err != nil {
			return err
		}
		movedCount += cliutils.Bool2Int(success)
	}

	logger.Logger.Info(moveMsgs[moveType].MovedMsg + " " + strconv.Itoa(movedCount) + " artifacts in Artifactory")
	return nil
}

func moveFile(sourcePath, destPath string, flags *MoveFlags, moveType MoveType) (bool, error) {
	message := moveMsgs[moveType].MovingMsg + " artifact: " + sourcePath + " to " + destPath
	if flags.DryRun == true {
		fmt.Println("[Dry run] " + message)
		return true, nil
	}

	logger.Logger.Info(message)

	moveUrl := flags.ArtDetails.Url
	restApi := "api/" + string(moveType) + "/" + sourcePath
	requestFullUrl, err := BuildArtifactoryUrl(moveUrl, restApi, map[string]string{"to": destPath})
	if err != nil {
		return false, err
	}
	httpClientsDetails := GetArtifactoryHttpClientDetails(flags.ArtDetails)
	resp, _, err := ioutils.SendPost(requestFullUrl, nil, httpClientsDetails)
	if err != nil {
		return false, err
	}

	logger.Logger.Info("Artifactory response:", resp.Status)
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