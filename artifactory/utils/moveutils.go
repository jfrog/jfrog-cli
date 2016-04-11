package utils

import (
	"strings"
	"fmt"
	"strconv"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
)

const (
	MOVE MoveType = "move"
	COPY MoveType = "copy"
)

func MoveFiles(regexpPath string, resultItems []AqlSearchResultItem, destPath string, flags *Flags, moveType MoveType) {
	movedCount := 0

	for _, v := range resultItems {
		destPathLocal := destPath
		if !flags.Flat {
			file, dir := ioutils.GetFileAndDirFromPath(destPathLocal)
			destPathLocal = cliutils.TrimPath(dir + "/" + v.Path + "/" + file)
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

func moveFile(sourcePath, destPath string, flags *Flags, moveType MoveType) bool {
	fmt.Println(moveMsgs[moveType].MovingMsg + " artifact: " + sourcePath + " to " + destPath)

	moveUrl := flags.ArtDetails.Url
	restApi := "api/" + string(moveType) + "/" + sourcePath + "?to=" + destPath
	httpClientsDetails := GetArtifactoryHttpClientDetails(flags.ArtDetails)
	resp, _ := ioutils.SendPost(moveUrl + restApi, nil, httpClientsDetails)

	fmt.Println("Artifactory response:", resp.Status)
	return resp.StatusCode == 200
}

var moveMsgs = map[MoveType]MoveOptions{
	MOVE: MoveOptions{MovingMsg: "moving", MovedMsg: "moved"},
	COPY: MoveOptions{MovingMsg: "copying", MovedMsg: "copied"},
}

type MoveOptions struct {
	MovingMsg string
	MovedMsg  string
}

type MoveType string