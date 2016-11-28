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
	"strings"
)

func BuildDistribute(buildName, buildNumber, targetRepo string, flags *BuildDistributeFlags) (err error) {
	err = utils.PreCommandSetup(flags)
	if err != nil {
		return
	}

	dryRun := ""
	if flags.DryRun == true {
		dryRun = "[Dry run] "
	}
	message := "Destributing build..."
	log.Info(dryRun + message)

	distributeUrl := flags.ArtDetails.Url
	restApi := path.Join("api/build/distribute/", buildName, buildNumber)
	requestFullUrl, err := utils.BuildArtifactoryUrl(distributeUrl, restApi, make(map[string]string))
	if err != nil {
		return
	}

	data := DistributionConfigContent{
		SourceRepo:             strings.Split(flags.SourceRepo, ","),
		TargetRepo:             targetRepo,
		Publish:                flags.Publish,
		OverrideExistingFiles:  flags.OverrideExistingFiles,
		GpgPassphrase:          flags.GpgPassphrase,
		Async:                  flags.Async,
		DryRun:                 flags.DryRun}
	requestContent, err := json.Marshal(data)
	if err != nil {
		err = cliutils.CheckError(errors.New("Failed to execute request. " + cliutils.GetDocumentationMessage()))
		if err != nil {
			return
		}
	}

	httpClientsDetails := utils.GetArtifactoryHttpClientDetails(flags.ArtDetails)
	utils.SetContentType("application/json", &httpClientsDetails.Headers)

	resp, body, err := ioutils.SendPost(requestFullUrl, requestContent, httpClientsDetails)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = cliutils.CheckError(errors.New(string(body)))
		return
	}
	if flags.Async && !flags.DryRun {
		log.Info("Asynchronously distributed build", buildName, "#" + buildNumber, "to:", targetRepo, "repository, logs are avalable in Artifactory.")
		return
	}

	log.Info(dryRun + "Distributed build", buildName, "#" + buildNumber, "to:", targetRepo, "repository.")
	return
}

type BuildDistributeFlags struct {
	ArtDetails            *config.ArtifactoryDetails
	SourceRepo            string
	GpgPassphrase         string
	Publish               bool
	OverrideExistingFiles bool
	Async                 bool
	DryRun                bool
}

type DistributionConfigContent struct {
	SourceRepo            []string  `json:"sourceRepos,omitempty"`
	TargetRepo            string    `json:"targetRepo,omitempty"`
	GpgPassphrase         string    `json:"gpgPassphrase,omitempty"`
	Publish               bool      `json:"publish"`
	OverrideExistingFiles bool      `json:"overrideExistingFiles,omitempty"`
	Async                 bool      `json:"async,omitempty"`
	DryRun                bool      `json:"dryRun,omitempty"`
}

func (flags *BuildDistributeFlags) GetArtifactoryDetails() *config.ArtifactoryDetails {
	return flags.ArtDetails
}

func (flags *BuildDistributeFlags) IsDryRun() bool {
	return flags.DryRun
}
