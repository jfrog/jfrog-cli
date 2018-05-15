package buildinfo

import (
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/spec"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/buildinfo"
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
)

func AddArtifact(addArtifactSpec *spec.SpecFiles, flags *BuildAddArtifactConfiguration) (err error) {
	log.Info("Adding artifacts to build info " + flags.BuildName + " #" + flags.BuildNumber + "...")

	servicesManager, err := utils.CreateServiceManager(flags.ArtDetails, false); if err != nil {
		return err
	}

	buildArtifacts, err := getBuildArtifacts(servicesManager, addArtifactSpec); if err != nil {
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

func getBuildArtifacts(servicesManager *artifactory.ArtifactoryServicesManager, addArtifactSpec *spec.SpecFiles) ([]buildinfo.InternalArtifact, error) {
	var buildArtifacts []buildinfo.InternalArtifact
	for i := 0; i < len(addArtifactSpec.Files); i++ {
		params, err := addArtifactSpec.Get(i).ToArtifatorySearchParams(); if err != nil {
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

type BuildAddArtifactConfiguration struct {
	ArtDetails  *config.ArtifactoryDetails
	BuildName   string
	BuildNumber string
}
