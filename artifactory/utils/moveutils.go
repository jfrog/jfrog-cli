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

func moveFiles(regexpPath string, resultItems []AqlSearchResultItem, destPath string, flags *MoveFlags, moveType MoveType) error {
	movedCount := 0

	for _, v := range resultItems {
		destPathLocal := destPath
		if !flags.Flat {
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

func MoveFilesWrapper(sourcePattern, destPath string, flags *MoveFlags, moveType MoveType) (err error) {
	PreCommandSetup(flags)
	if IsWildcardPattern(sourcePattern) || flags.Props != "" {
	    var resultItems []AqlSearchResultItem
		resultItems, err = AqlSearchDefaultReturnFields(sourcePattern, flags)
		if err != nil {
		    return
		}
		regexpPath := cliutils.PathToRegExp(sourcePattern)
		err = moveFiles(regexpPath, resultItems, destPath, flags, moveType)
	} else {
		_, err = moveFile(sourcePattern, destPath, flags, moveType)
	}
	return
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
	Recursive    bool
	Flat         bool
	DryRun       bool
	Props        string
	ArtDetails   *config.ArtifactoryDetails
}

func (flags *MoveFlags) GetArtifactoryDetails() *config.ArtifactoryDetails {
	return flags.ArtDetails
}

func (flags *MoveFlags) IsRecursive() bool {
	return flags.Recursive
}

func (flags *MoveFlags) GetProps() string {
	return flags.Props
}

func (flags *MoveFlags) IsDryRun() bool {
	return flags.DryRun
}