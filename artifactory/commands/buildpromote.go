package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"encoding/json"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"path"
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
)

func BuildPromote(buildName, buildNumber, targetRepo string, flags *BuildPromoteFlags) (err error) {
	err = utils.PreCommandSetup(flags)
	if err != nil {
		return
	}

	message := "Promoting build..."
	if flags.DryRun == true {
		message = "[Dry run] " + message
	}
	log.Info(message)

	promoteUrl := flags.ArtDetails.Url
	restApi := path.Join("api/build/promote/", buildName, buildNumber)
	requestFullUrl, err := utils.BuildArtifactoryUrl(promoteUrl, restApi, make(map[string]string))
	if err != nil {
		return
	}

	data := PromotionConfigContent{
		Status:                 flags.Status,
		Comment :               flags.Comment,
		Copy:                   flags.Copy,
		IncludeDependencies:    flags.IncludeDependencies,
		SourceRepo:             flags.SourceRepo,
		TargetRepo:             targetRepo,
		DryRun:                 flags.DryRun}
	requestContent, err := json.Marshal(data)
	if err != nil {
		err = cliutils.CheckError(errors.New("Failed to execute request. " + cliutils.GetDocumentationMessage()))
		if err != nil {
			return
		}
	}

	httpClientsDetails := utils.GetArtifactoryHttpClientDetails(flags.ArtDetails)
	utils.SetContentType("application/vnd.org.jfrog.artifactory.build.PromotionRequest+json", &httpClientsDetails.Headers)

	resp, body, err := ioutils.SendPost(requestFullUrl, requestContent, httpClientsDetails)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = cliutils.CheckError(errors.New(string(body)))
		return
	}
	log.Info("Promoted build", buildName , "#" + buildNumber, "to:", targetRepo, "repository.")
	return
}

type BuildPromoteFlags struct {
	ArtDetails          *config.ArtifactoryDetails
	Comment             string
	SourceRepo          string
	Status              string
	IncludeDependencies bool
	Copy                bool
	DryRun              bool
}

type PromotionConfigContent struct {
	Comment             string `json:"comment,omitempty"`
	SourceRepo          string `json:"sourceRepo,omitempty"`
	TargetRepo          string `json:"targetRepo,omitempty"`
	Status              string `json:"status,omitempty"`
	IncludeDependencies bool   `json:"dependencies,omitempty"`
	Copy                bool   `json:"copy,omitempty"`
	DryRun              bool   `json:"dryRun,omitempty"`
}

func (flags *BuildPromoteFlags) GetArtifactoryDetails() *config.ArtifactoryDetails {
	return flags.ArtDetails
}

func (flags *BuildPromoteFlags) IsDryRun() bool {
	return flags.DryRun
}
