package buildinfo

import (
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/spec"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/buildinfo"
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
)

func AddArtifacts(addArtifactsSpec *spec.SpecFiles, flags *BuildAddArtifactsConfiguration) (err error) {
	log.Info("Adding artifacts to build info " + flags.BuildName + " #" + flags.BuildNumber + "...")

	buildArtifacts, err := getBuildArtifacts(flags.ArtDetails, addArtifactsSpec); if err != nil {
		return err
	}

	err = utils.SaveBuildGeneralDetails(flags.BuildName, flags.BuildNumber); if err != nil {
		return
	}

	populateFunc := func(partial *buildinfo.Partial) {
		partial.Artifacts = buildArtifacts
	}
	err = utils.SavePartialBuildInfo(flags.BuildName, flags.BuildNumber, populateFunc); if err != nil {
		return
	}

	log.Info("Successfully added artifact to build info")
	return
}

func getBuildArtifacts(artifactoryDetails *config.ArtifactoryDetails, addArtifactsSpec *spec.SpecFiles) ([]buildinfo.InternalArtifact, error) {
	var buildArtifacts []buildinfo.InternalArtifact

	servicesManager, err := utils.CreateServiceManager(artifactoryDetails, false); if err != nil {
		return nil, err
	}

	for i := 0; i < len(addArtifactsSpec.Files); i++ {
		params, err := addArtifactsSpec.Get(i).ToArtifatorySearchParams(); if err != nil {
			return nil, err
		}
		resultItems, err := servicesManager.Search(&clientutils.SearchParamsImpl{ArtifactoryCommonParams: params}); if err != nil {
			return nil, err
		}
		for _, resultItem := range resultItems {
			artifactPath := resultItem.Repo + "/" + resultItem.Path + "/" + resultItem.Name
			artifact := buildinfo.InternalArtifact{Path: artifactPath, Checksum: &buildinfo.Checksum{}}
			artifact.Sha1 = resultItem.Actual_Sha1
			artifact.Md5 = resultItem.Actual_Md5
			buildArtifacts = append(buildArtifacts, artifact)
		}
	}

	if len(buildArtifacts) == 0 {
		return nil, errorutils.CheckError(errors.New("No artifacts to add found."))
	}

	return buildArtifacts, nil
}

type BuildAddArtifactsConfiguration struct {
	ArtDetails  *config.ArtifactoryDetails
	BuildName   string
	BuildNumber string
}
