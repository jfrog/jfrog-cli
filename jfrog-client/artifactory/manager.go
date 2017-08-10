package artifactory

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"net/http"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils/auth/cert"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/types/httpclient"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
)

type ArtifactoryServicesManager struct {
	client *httpclient.HttpClient
	config ArtifactoryConfig
}

func NewArtifactoryService(config ArtifactoryConfig) (*ArtifactoryServicesManager, error) {
	var err error
	manager := &ArtifactoryServicesManager{config: config}
	if config.GetCertifactesPath() == "" {
		manager.client = httpclient.NewDefaultJforgHttpClient()
	} else {
		transport, err := cert.GetTransportWithLoadedCert(config.GetCertifactesPath())
		if err != nil {
			return nil, err
		}
		manager.client = httpclient.NewJforgHttpClient(&http.Client{Transport: transport})
	}
	if config.GetLogger() != nil {
		log.SetLogger(config.GetLogger())
	}
	return manager, err
}

func (sm *ArtifactoryServicesManager) DistributeBuild(params services.BuildDistributionParams) error {
	distributionService := services.NewDistributionService(sm.client)
	distributionService.DryRun = sm.config.IsDryRun()
	distributionService.ArtDetails = sm.config.GetArtDetails()
	return distributionService.BuildDistribute(params)
}

func (sm *ArtifactoryServicesManager) PromoteBuild(params services.PromotionParams) error {
	promotionService := services.NewPromotionService(sm.client)
	promotionService.DryRun = sm.config.IsDryRun()
	promotionService.ArtDetails = sm.config.GetArtDetails()
	return promotionService.BuildPromote(params)
}

func (sm *ArtifactoryServicesManager) GetPathsToDelete(params services.DeleteParams) ([]utils.ResultItem, error) {
	deleteService := services.NewDeleteService(sm.client)
	deleteService.DryRun = sm.config.IsDryRun()
	deleteService.ArtDetails = sm.config.GetArtDetails()
	return deleteService.GetPathsToDelete(params)
}

func (sm *ArtifactoryServicesManager) DeleteFiles(resultItems []services.DeleteItem) error {
	deleteService := services.NewDeleteService(sm.client)
	deleteService.DryRun = sm.config.IsDryRun()
	deleteService.ArtDetails = sm.config.GetArtDetails()
	return deleteService.DeleteFiles(resultItems, deleteService)
}

func (sm *ArtifactoryServicesManager) DownloadFiles(params services.DownloadParams) ([]utils.FileInfo, error) {
	downloadService := services.NewDownloadService(sm.client)
	downloadService.DryRun = sm.config.IsDryRun()
	downloadService.ArtDetails = sm.config.GetArtDetails()
	downloadService.Threads = sm.config.GetNumOfThreadPerOperation()
	downloadService.SplitCount = sm.config.GetSplitCount()
	downloadService.MinSplitSize = sm.config.GetMinSplitSize()
	return downloadService.DownloadFiles(params)
}

func (sm *ArtifactoryServicesManager) GetUnreferencedGitLfsFiles(params services.GitLfsCleanParams) ([]utils.ResultItem, error) {
	gitLfsCleanService := services.NewGitLfsCleanService(sm.client)
	gitLfsCleanService.DryRun = sm.config.IsDryRun()
	gitLfsCleanService.ArtDetails = sm.config.GetArtDetails()
	return gitLfsCleanService.GetUnreferencedGitLfsFiles(params)
}

func (sm *ArtifactoryServicesManager) Search(params utils.SearchParams) ([]utils.ResultItem, error) {
	searchService := services.NewSearchService(sm.client)
	searchService.ArtDetails = sm.config.GetArtDetails()
	return searchService.Search(params)
}

func (sm *ArtifactoryServicesManager) SetProps(params services.SetPropsParams) error {
	setPropsService := services.NewSetPropsService(sm.client)
	setPropsService.ArtDetails = sm.config.GetArtDetails()
	return setPropsService.SetProps(params)
}

func (sm *ArtifactoryServicesManager) UploadFiles(params services.UploadParams) ([]utils.FileInfo, int, int, error) {
	uploadService := services.NewUploadService(sm.client)
	sm.setCommonServiceConfig(uploadService)
	uploadService.MinChecksumDeploy = sm.config.GetMinChecksumDeploy()
	return uploadService.UploadFiles(params)
}

func (sm *ArtifactoryServicesManager) Copy(params services.MoveCopyParams) error {
	copyService := services.NewMoveCopyService(sm.client, services.COPY)
	copyService.ArtDetails = sm.config.GetArtDetails()
	return copyService.MoveCopyServiceMoveFilesWrapper(params)
}

func (sm *ArtifactoryServicesManager) Move(params services.MoveCopyParams) error {
	moveService := services.NewMoveCopyService(sm.client, services.MOVE)
	moveService.ArtDetails = sm.config.GetArtDetails()
	return moveService.MoveCopyServiceMoveFilesWrapper(params)
}

func (sm *ArtifactoryServicesManager) setCommonServiceConfig(commonConfig ArtifactoryServicesSetter) {
	commonConfig.SetThread(sm.config.GetNumOfThreadPerOperation())
	commonConfig.SetArtDetails(sm.config.GetArtDetails())
	commonConfig.SetDryRun(sm.config.IsDryRun())
}