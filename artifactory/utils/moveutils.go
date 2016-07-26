package utils

import (
	"strings"
	"fmt"
	"strconv"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
)

const (
	MOVE MoveType = "move"
	COPY MoveType = "copy"
)

func moveFiles(regexpPath string, resultItems []AqlSearchResultItem, destPath string, flags *MoveFlags, moveType MoveType) {
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
		destFile := cliutils.ReformatRegexp(regexpPath, v.GetFullUrl(), destPathLocal)
		if strings.HasSuffix(destFile, "/") {
			destFile += v.Name
		}
		success := moveFile(v.GetFullUrl(), destFile, flags, moveType)
		movedCount += cliutils.Bool2Int(success)
	}

	fmt.Println(moveMsgs[moveType].MovedMsg + " " + strconv.Itoa(movedCount) + " artifacts in Artifactory")
}

func moveFile(sourcePath, destPath string, flags *MoveFlags, moveType MoveType) bool {
	message := moveMsgs[moveType].MovingMsg + " artifact: " + sourcePath + " to " + destPath
	if flags.DryRun == true {
		fmt.Println("[Dry run] " + message)
		return true
	}

	fmt.Println(message)

	moveUrl := flags.ArtDetails.Url
	restApi := "api/" + string(moveType) + "/" + sourcePath
	requestFullUrl := BuildArtifactoryUrl(moveUrl, restApi, map[string]string{"to": destPath})
	httpClientsDetails := GetArtifactoryHttpClientDetails(flags.ArtDetails)
	resp, _ := ioutils.SendPost(requestFullUrl, nil, httpClientsDetails)

	fmt.Println("Artifactory response:", resp.Status)
	return resp.StatusCode == 200
}

func MoveFilesWrapper(sourcePattern, destPath string, flags *MoveFlags, moveType MoveType) {
	PreCommandSetup(flags)
	if IsWildcardPattern(sourcePattern) || flags.Props != "" {
		resultItems := AqlSearchDefaultReturnFields(sourcePattern, flags)
		regexpPath := cliutils.PathToRegExp(sourcePattern)
		moveFiles(regexpPath, resultItems, destPath, flags, moveType)
	} else {
		moveFile(sourcePattern, destPath, flags, moveType)
	}
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