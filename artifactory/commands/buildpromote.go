package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/io/httputils"
	"encoding/json"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"path"
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
)

func BuildPromote(buildName, buildNumber, targetRepo string, flags *BuildPromotionFlags) error {
	err := utils.PreCommandSetup(flags)
	if err != nil {
		return err
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
		return err
	}

	data := BuildPromotionConfig{
		Status:                 flags.Status,
		Comment :               flags.Comment,
		Copy:                   flags.Copy,
		IncludeDependencies:    flags.IncludeDependencies,
		SourceRepo:             flags.SourceRepo,
		TargetRepo:             targetRepo,
		DryRun:                 flags.DryRun}
	requestContent, err := json.Marshal(data)
	if err != nil {
		return cliutils.CheckError(errors.New("Failed to execute request. " + cliutils.GetDocumentationMessage()))
	}

	httpClientsDetails := utils.GetArtifactoryHttpClientDetails(flags.ArtDetails)
	utils.SetContentType("application/vnd.org.jfrog.artifactory.build.PromotionRequest+json", &httpClientsDetails.Headers)

	resp, body, err := httputils.SendPost(requestFullUrl, requestContent, httpClientsDetails)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return cliutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + cliutils.IndentJson(body)))
	}

	log.Debug("Artifactory response:", resp.Status)
	log.Info("Promoted build", buildName , "#" + buildNumber, "to:", targetRepo, "repository.")
	return nil
}

type BuildPromotionFlags struct {
	ArtDetails          *config.ArtifactoryDetails
	Comment             string
	SourceRepo          string
	Status              string
	IncludeDependencies bool
	Copy                bool
	DryRun              bool
}

type BuildPromotionConfig struct {
	Comment             string `json:"comment,omitempty"`
	SourceRepo          string `json:"sourceRepo,omitempty"`
	TargetRepo          string `json:"targetRepo,omitempty"`
	Status              string `json:"status,omitempty"`
	IncludeDependencies bool   `json:"dependencies,omitempty"`
	Copy                bool   `json:"copy,omitempty"`
	DryRun              bool   `json:"dryRun,omitempty"`
}

func (flags *BuildPromotionFlags) GetArtifactoryDetails() *config.ArtifactoryDetails {
	return flags.ArtDetails
}

func (flags *BuildPromotionFlags) IsDryRun() bool {
	return flags.DryRun
}
