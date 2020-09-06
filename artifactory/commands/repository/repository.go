package repository

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/jfrog/jfrog-cli/artifactory/commands/utils"
	rtUtils "github.com/jfrog/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

type RepoCommand struct {
	rtDetails    *config.ArtifactoryDetails
	templatePath string
	vars         string
}

func (rc *RepoCommand) Vars() string {
	return rc.vars
}

func (rc *RepoCommand) TemplatePath() string {
	return rc.templatePath
}

func (rc *RepoCommand) PerformRepoCmd(isUpdate bool) (err error) {
	repoConfigMap, err := utils.ConvertTemplateToMap(rc)
	if err != nil {
		return err
	}
	// All the values in the template are strings
	// Go over the the confMap and write the values with the correct type using the writersMap
	for key, value := range repoConfigMap {
		if err = utils.ValidateMapEntry(key, value, writersMap); err != nil {
			return
		}
		writersMap[key](&repoConfigMap, key, value.(string))
	}
	// Write a JSON with the correct values
	content, err := json.Marshal(repoConfigMap)

	servicesManager, err := rtUtils.CreateServiceManager(rc.rtDetails, false)
	if err != nil {
		return err
	}
	// Rclass and packgeType are mandatory keys in our templates
	// Using their values we'll pick the suitable handler from one of the handler maps to create/update a repository
	switch repoConfigMap[Rclass] {
	case Local:
		err = localRepoHandlers[repoConfigMap[PackageType].(string)](servicesManager, content, isUpdate)
	case Remote:
		err = remoteRepoHandlers[repoConfigMap[PackageType].(string)](servicesManager, content, isUpdate)
	case Virtual:
		err = virtualRepoHandlers[repoConfigMap[PackageType].(string)](servicesManager, content, isUpdate)
	default:
		return errorutils.CheckError(errors.New("unsupported rclass"))
	}
	return err
}

var writersMap = map[string]utils.AnswerWriter{
	Key:                               utils.WriteStringAnswer,
	Rclass:                            utils.WriteStringAnswer,
	PackageType:                       utils.WriteStringAnswer,
	MandatoryUrl:                      utils.WriteStringAnswer,
	Url:                               utils.WriteStringAnswer,
	Description:                       utils.WriteStringAnswer,
	Notes:                             utils.WriteStringAnswer,
	IncludePatterns:                   utils.WriteStringAnswer,
	ExcludePatterns:                   utils.WriteStringAnswer,
	RepoLayoutRef:                     utils.WriteStringAnswer,
	HandleReleases:                    utils.WriteBoolAnswer,
	HandleSnapshots:                   utils.WriteBoolAnswer,
	MaxUniqueSnapshots:                utils.WriteIntAnswer,
	SuppressPomConsistencyChecks:      utils.WriteBoolAnswer,
	BlackedOut:                        utils.WriteBoolAnswer,
	DownloadRedirect:                  utils.WriteBoolAnswer,
	BlockPushingSchema1:               utils.WriteBoolAnswer,
	DebianTrivialLayout:               utils.WriteBoolAnswer,
	ExternalDependenciesEnabled:       utils.WriteBoolAnswer,
	ExternalDependenciesPatterns:      utils.WriteStringArrayAnswer,
	ChecksumPolicyType:                utils.WriteStringAnswer,
	MaxUniqueTags:                     utils.WriteIntAnswer,
	SnapshotVersionBehavior:           utils.WriteStringAnswer,
	XrayIndex:                         utils.WriteBoolAnswer,
	PropertySets:                      utils.WriteStringArrayAnswer,
	ArchiveBrowsingEnabled:            utils.WriteBoolAnswer,
	CalculateYumMetadata:              utils.WriteBoolAnswer,
	YumRootDepth:                      utils.WriteIntAnswer,
	DockerApiVersion:                  utils.WriteStringAnswer,
	EnableFileListsIndexing:           utils.WriteBoolAnswer,
	OptionalIndexCompressionFormats:   utils.WriteStringAnswer,
	Username:                          utils.WriteStringAnswer,
	Password:                          utils.WriteStringAnswer,
	Proxy:                             utils.WriteStringAnswer,
	RemoteRepoChecksumPolicyType:      utils.WriteStringAnswer,
	HardFail:                          utils.WriteBoolAnswer,
	Offline:                           utils.WriteBoolAnswer,
	StoreArtifactsLocally:             utils.WriteBoolAnswer,
	SocketTimeoutMillis:               utils.WriteIntAnswer,
	LocalAddress:                      utils.WriteStringAnswer,
	RetrievalCachePeriodSecs:          utils.WriteIntAnswer,
	FailedRetrievalCachePeriodSecs:    utils.WriteIntAnswer,
	MissedRetrievalCachePeriodSecs:    utils.WriteIntAnswer,
	UnusedArtifactsCleanupEnabled:     utils.WriteBoolAnswer,
	UnusedArtifactsCleanupPeriodHours: utils.WriteIntAnswer,
	AssumedOfflinePeriodSecs:          utils.WriteIntAnswer,
	FetchJarsEagerly:                  utils.WriteBoolAnswer,
	FetchSourcesEagerly:               utils.WriteBoolAnswer,
	ShareConfiguration:                utils.WriteBoolAnswer,
	SynchronizeProperties:             utils.WriteBoolAnswer,
	BlockMismatchingMimeTypes:         utils.WriteBoolAnswer,
	AllowAnyHostAuth:                  utils.WriteBoolAnswer,
	EnableCookieManagement:            utils.WriteBoolAnswer,
	BowerRegistryUrl:                  utils.WriteStringAnswer,
	ComposerRegistryUrl:               utils.WriteStringAnswer,
	PyPIRegistryUrl:                   utils.WriteStringAnswer,
	VcsType:                           utils.WriteStringAnswer,
	VcsGitProvider:                    utils.WriteStringAnswer,
	VcsGitDownloadUrl:                 utils.WriteStringAnswer,
	BypassHeadRequests:                utils.WriteBoolAnswer,
	ClientTlsCertificate:              utils.WriteStringAnswer,
	FeedContextPath:                   utils.WriteStringAnswer,
	DownloadContextPath:               utils.WriteStringAnswer,
	V3FeedUrl:                         utils.WriteStringAnswer,
	ContentSynchronisation:            writeContentSynchronisation,
	ListRemoteFolderItems:             utils.WriteBoolAnswer,
	RejectInvalidJars:                 utils.WriteBoolAnswer,
	PodsSpecsRepoUrl:                  utils.WriteStringAnswer,
	EnableTokenAuthentication:         utils.WriteBoolAnswer,
	Repositories:                      utils.WriteStringArrayAnswer,
	ArtifactoryRequestsCanRetrieveRemoteArtifacts: utils.WriteBoolAnswer,
	KeyPair:                              utils.WriteStringAnswer,
	PomRepositoryReferencesCleanupPolicy: utils.WriteStringAnswer,
	DefaultDeploymentRepo:                utils.WriteStringAnswer,
	ForceMavenAuthentication:             utils.WriteBoolAnswer,
	ExternalDependenciesRemoteRepo:       utils.WriteStringAnswer,
}

func writeContentSynchronisation(resultMap *map[string]interface{}, key, value string) error {
	answerArray := strings.Split(value, ",")
	if len(answerArray) != 4 {
		return errors.New("invalid value for Content Synchronisation")
	}
	var cs services.ContentSynchronisation
	cs.Enabled, _ = strconv.ParseBool(answerArray[0])
	cs.Statistics.Enabled, _ = strconv.ParseBool(answerArray[1])
	cs.Properties.Enabled, _ = strconv.ParseBool(answerArray[2])
	cs.Source.OriginAbsenceDetection, _ = strconv.ParseBool(answerArray[3])

	(*resultMap)[key] = cs
	return nil

}

// repoHandler is a function that gets serviceManager, JSON configuration content and a flag indicates is the operation in an update operation
// Each handler unmarshal the JSOn content into the jfrog-client's unique rclass-pkgType param struct, and run the operation service
type repoHandler func(artifactory.ArtifactoryServicesManager, []byte, bool) error

var localRepoHandlers = map[string]repoHandler{
	Maven:     localMavenHandler,
	Gradle:    localGradleHandler,
	Ivy:       localIvyHandles,
	Sbt:       localSbtHandler,
	Helm:      localHelmHandler,
	Cocoapods: localCocoapodsHandler,
	Opkg:      localOpkgHandler,
	Rpm:       localRpmHandler,
	Nuget:     localNugetHandler,
	Cran:      localCranHandler,
	Gems:      localGemsHandler,
	Npm:       localNpmHandler,
	Bower:     localBowerHandler,
	Debian:    localDebianHandler,
	Composer:  localComposerHandler,
	Pypi:      localPypiHandler,
	Docker:    localDockerHandler,
	Vagrant:   localVagrantHandler,
	Gitlfs:    localGitlfsHandler,
	Go:        localGoHandler,
	Yum:       localYumHandler,
	Conan:     localConanHandler,
	Chef:      localChefHandler,
	Puppet:    localPuppetHandler,
	Generic:   localGenericHandler,
}

func localMavenHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewMavenLocalRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateLocalRepository().Maven(params)
	} else {
		err = servicesManager.CreateLocalRepository().Maven(params)
	}
	return err
}

func localGradleHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewGradleLocalRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateLocalRepository().Gradle(params)
	} else {
		err = servicesManager.CreateLocalRepository().Gradle(params)
	}
	return err
}

func localIvyHandles(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewIvyLocalRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateLocalRepository().Ivy(params)
	} else {
		err = servicesManager.CreateLocalRepository().Ivy(params)
	}
	return err
}

func localSbtHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewSbtLocalRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateLocalRepository().Sbt(params)
	} else {
		err = servicesManager.CreateLocalRepository().Sbt(params)
	}
	return err
}

func localHelmHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewHelmLocalRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateLocalRepository().Helm(params)
	} else {
		err = servicesManager.CreateLocalRepository().Helm(params)
	}
	return err
}

func localCocoapodsHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewCocoapodsLocalRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateLocalRepository().Cocoapods(params)
	} else {
		err = servicesManager.CreateLocalRepository().Cocoapods(params)
	}
	return err
}

func localOpkgHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewOpkgLocalRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateLocalRepository().Opkg(params)
	} else {
		err = servicesManager.CreateLocalRepository().Opkg(params)
	}
	return err
}

func localRpmHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewRpmLocalRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateLocalRepository().Rpm(params)
	} else {
		err = servicesManager.CreateLocalRepository().Rpm(params)
	}
	return err
}

func localNugetHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewNugetLocalRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateLocalRepository().Nuget(params)
	} else {
		err = servicesManager.CreateLocalRepository().Nuget(params)
	}
	return err
}

func localCranHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewCranLocalRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateLocalRepository().Cran(params)
	} else {
		err = servicesManager.CreateLocalRepository().Cran(params)
	}
	return err
}

func localGemsHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewGemsLocalRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateLocalRepository().Gems(params)
	} else {
		err = servicesManager.CreateLocalRepository().Gems(params)
	}
	return err
}

func localNpmHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewNpmLocalRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateLocalRepository().Npm(params)
	} else {
		err = servicesManager.CreateLocalRepository().Npm(params)
	}
	return err
}

func localBowerHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewBowerLocalRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateLocalRepository().Bower(params)
	} else {
		err = servicesManager.CreateLocalRepository().Bower(params)
	}
	return err
}

func localDebianHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewDebianLocalRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateLocalRepository().Debian(params)
	} else {
		err = servicesManager.CreateLocalRepository().Debian(params)
	}
	return err
}

func localComposerHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewComposerLocalRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateLocalRepository().Composer(params)
	} else {
		err = servicesManager.CreateLocalRepository().Composer(params)
	}
	return err
}

func localPypiHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewPypiLocalRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateLocalRepository().Pypi(params)
	} else {
		err = servicesManager.CreateLocalRepository().Pypi(params)
	}
	return err
}

func localDockerHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewDockerLocalRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateLocalRepository().Docker(params)
	} else {
		err = servicesManager.CreateLocalRepository().Docker(params)
	}
	return err
}

func localVagrantHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewVagrantLocalRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateLocalRepository().Vagrant(params)
	} else {
		err = servicesManager.CreateLocalRepository().Vagrant(params)
	}
	return err
}

func localGitlfsHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewGitlfsLocalRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateLocalRepository().Gitlfs(params)
	} else {
		err = servicesManager.CreateLocalRepository().Gitlfs(params)
	}
	return err
}

func localGoHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewGoLocalRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateLocalRepository().Go(params)
	} else {
		err = servicesManager.CreateLocalRepository().Go(params)
	}
	return err
}

func localYumHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewYumLocalRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateLocalRepository().Yum(params)
	} else {
		err = servicesManager.CreateLocalRepository().Yum(params)
	}
	return err
}

func localConanHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewConanLocalRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateLocalRepository().Conan(params)
	} else {
		err = servicesManager.CreateLocalRepository().Conan(params)
	}
	return err
}

func localChefHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewChefLocalRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateLocalRepository().Chef(params)
	} else {
		err = servicesManager.CreateLocalRepository().Chef(params)
	}
	return err
}

func localPuppetHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewPuppetLocalRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateLocalRepository().Puppet(params)
	} else {
		err = servicesManager.CreateLocalRepository().Puppet(params)
	}
	return err
}

func localGenericHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewGenericLocalRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}

	if isUpdate {
		err = servicesManager.UpdateLocalRepository().Generic(params)
	} else {
		err = servicesManager.CreateLocalRepository().Generic(params)
	}
	return err
}

var remoteRepoHandlers = map[string]repoHandler{
	Maven:     remoteMavenHandler,
	Gradle:    remoteGradleHandler,
	Ivy:       remoteIvyHandler,
	Sbt:       remoteSbtHandler,
	Helm:      remoteHelmHandler,
	Cocoapods: remoteCocoapodsHandler,
	Opkg:      remoteOpkgHandler,
	Rpm:       remoteRpmHandler,
	Nuget:     remoteNugetHandler,
	Cran:      remoteCranHandler,
	Gems:      remoteGemsHandler,
	Npm:       remoteNpmHandler,
	Bower:     remoteBowerHandler,
	Debian:    remoteDebianHandler,
	Composer:  remoteComposerHandler,
	Pypi:      remotelPypiHandler,
	Docker:    remoteDockerHandler,
	Gitlfs:    remoteGitlfsHandler,
	Go:        remoteGoHandler,
	Yum:       remoteYumHandler,
	Conan:     remoteConanHandler,
	Chef:      remoteChefHandler,
	Puppet:    remotePuppetHandler,
	Conda:     remoteCondaHandler,
	P2:        remoteP2Handler,
	Vcs:       remoteVcsHandler,
	Generic:   remoteGenericHandler,
}

func remoteMavenHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewMavenRemoteRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateRemoteRepository().Maven(params)
	} else {
		err = servicesManager.CreateRemoteRepository().Maven(params)
	}
	return err
}

func remoteGradleHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewGradleRemoteRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateRemoteRepository().Gradle(params)
	} else {
		err = servicesManager.CreateRemoteRepository().Gradle(params)
	}
	return err
}

func remoteIvyHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewIvyRemoteRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateRemoteRepository().Ivy(params)
	} else {
		err = servicesManager.CreateRemoteRepository().Ivy(params)
	}
	return err
}

func remoteSbtHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewSbtRemoteRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateRemoteRepository().Sbt(params)
	} else {
		err = servicesManager.CreateRemoteRepository().Sbt(params)
	}
	return err
}

func remoteHelmHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewHelmRemoteRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateRemoteRepository().Helm(params)
	} else {
		err = servicesManager.CreateRemoteRepository().Helm(params)
	}
	return err
}

func remoteCocoapodsHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewCocoapodsRemoteRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateRemoteRepository().Cocoapods(params)
	} else {
		err = servicesManager.CreateRemoteRepository().Cocoapods(params)
	}
	return err
}

func remoteOpkgHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewOpkgRemoteRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateRemoteRepository().Opkg(params)
	} else {
		err = servicesManager.CreateRemoteRepository().Opkg(params)
	}
	return err
}

func remoteRpmHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewRpmRemoteRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateRemoteRepository().Rpm(params)
	} else {
		err = servicesManager.CreateRemoteRepository().Rpm(params)
	}
	return err
}

func remoteNugetHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewNugetRemoteRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateRemoteRepository().Nuget(params)
	} else {
		err = servicesManager.CreateRemoteRepository().Nuget(params)
	}
	return err
}

func remoteCranHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewCranRemoteRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateRemoteRepository().Cran(params)
	} else {
		err = servicesManager.CreateRemoteRepository().Cran(params)
	}
	return err
}

func remoteGemsHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewGemsRemoteRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateRemoteRepository().Gems(params)
	} else {
		err = servicesManager.CreateRemoteRepository().Gems(params)
	}
	return err
}

func remoteNpmHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewNpmRemoteRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateRemoteRepository().Npm(params)
	} else {
		err = servicesManager.CreateRemoteRepository().Npm(params)
	}
	return err
}

func remoteBowerHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewBowerRemoteRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateRemoteRepository().Bower(params)
	} else {
		err = servicesManager.CreateRemoteRepository().Bower(params)
	}
	return err
}

func remoteDebianHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewDebianRemoteRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateRemoteRepository().Debian(params)
	} else {
		err = servicesManager.CreateRemoteRepository().Debian(params)
	}
	return err
}

func remoteComposerHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewComposerRemoteRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateRemoteRepository().Composer(params)
	} else {
		err = servicesManager.CreateRemoteRepository().Composer(params)
	}
	return err
}

func remotelPypiHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewPypiRemoteRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateRemoteRepository().Pypi(params)
	} else {
		err = servicesManager.CreateRemoteRepository().Pypi(params)
	}
	return err
}

func remoteDockerHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewDockerRemoteRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateRemoteRepository().Docker(params)
	} else {
		err = servicesManager.CreateRemoteRepository().Docker(params)
	}
	return err
}

func remoteGitlfsHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewGitlfsRemoteRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateRemoteRepository().Gitlfs(params)
	} else {
		err = servicesManager.CreateRemoteRepository().Gitlfs(params)
	}
	return err
}

func remoteGoHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewGoRemoteRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateRemoteRepository().Go(params)
	} else {
		err = servicesManager.CreateRemoteRepository().Go(params)
	}
	return err
}

func remoteConanHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewConanRemoteRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateRemoteRepository().Conan(params)
	} else {
		err = servicesManager.CreateRemoteRepository().Conan(params)
	}
	return err
}

func remoteChefHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewChefRemoteRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateRemoteRepository().Chef(params)
	} else {
		err = servicesManager.CreateRemoteRepository().Chef(params)
	}
	return err
}

func remotePuppetHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewPuppetRemoteRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateRemoteRepository().Puppet(params)
	} else {
		err = servicesManager.CreateRemoteRepository().Puppet(params)
	}
	return err
}

func remoteVcsHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewVcsRemoteRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateRemoteRepository().Vcs(params)
	} else {
		err = servicesManager.CreateRemoteRepository().Vcs(params)
	}
	return err
}

func remoteP2Handler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewP2RemoteRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateRemoteRepository().P2(params)
	} else {
		err = servicesManager.CreateRemoteRepository().P2(params)
	}
	return err
}

func remoteCondaHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewCondaRemoteRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateRemoteRepository().Conda(params)
	} else {
		err = servicesManager.CreateRemoteRepository().Conda(params)
	}
	return err
}

func remoteYumHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewYumRemoteRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateRemoteRepository().Yum(params)
	} else {
		err = servicesManager.CreateRemoteRepository().Yum(params)
	}
	return err
}

func remoteGenericHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewGenericRemoteRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateRemoteRepository().Generic(params)
	} else {
		err = servicesManager.CreateRemoteRepository().Generic(params)
	}
	return err
}

var virtualRepoHandlers = map[string]repoHandler{
	Maven:   virtualMavenHandler,
	Gradle:  virtualGradleHandler,
	Ivy:     virtualIvyHandler,
	Sbt:     virtualSbtHandler,
	Helm:    virtualHelmHandler,
	Rpm:     virtualRpmHandler,
	Nuget:   virtualNugetHandler,
	Cran:    virtualCranHandler,
	Gems:    virtualGemsHandler,
	Npm:     virtualNpmHandler,
	Bower:   virtualBowerHandler,
	Debian:  virtualDebianHandler,
	Pypi:    virtualPypiHandler,
	Docker:  virtualDockerHandler,
	Gitlfs:  virtualGitlfsHandler,
	Go:      virtualGoHandler,
	Yum:     virtualYumHandler,
	Conan:   virtualConanHandler,
	Chef:    virtualChefHandler,
	Puppet:  virtualPuppetHandler,
	Conda:   virtualCondaHandler,
	P2:      virtualP2Handler,
	Generic: virtualGenericHandler,
}

func virtualMavenHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewMavenVirtualRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateVirtualRepository().Maven(params)
	} else {
		err = servicesManager.CreateVirtualRepository().Maven(params)
	}
	return err
}

func virtualGradleHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewGradleVirtualRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateVirtualRepository().Gradle(params)
	} else {
		err = servicesManager.CreateVirtualRepository().Gradle(params)
	}
	return err
}

func virtualIvyHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewIvyVirtualRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateVirtualRepository().Ivy(params)
	} else {
		err = servicesManager.CreateVirtualRepository().Ivy(params)
	}
	return err
}

func virtualSbtHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewSbtVirtualRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateVirtualRepository().Sbt(params)
	} else {
		err = servicesManager.CreateVirtualRepository().Sbt(params)
	}
	return err
}

func virtualHelmHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewHelmVirtualRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateVirtualRepository().Helm(params)
	} else {
		err = servicesManager.CreateVirtualRepository().Helm(params)
	}
	return err
}

func virtualRpmHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewRpmVirtualRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateVirtualRepository().Rpm(params)
	} else {
		err = servicesManager.CreateVirtualRepository().Rpm(params)
	}
	return err
}

func virtualNugetHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewNugetVirtualRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateVirtualRepository().Nuget(params)
	} else {
		err = servicesManager.CreateVirtualRepository().Nuget(params)
	}
	return err
}

func virtualCranHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewCranVirtualRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateVirtualRepository().Cran(params)
	} else {
		err = servicesManager.CreateVirtualRepository().Cran(params)
	}
	return err
}

func virtualGemsHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewGemsVirtualRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateVirtualRepository().Gems(params)
	} else {
		err = servicesManager.CreateVirtualRepository().Gems(params)
	}
	return err
}

func virtualNpmHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewNpmVirtualRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateVirtualRepository().Npm(params)
	} else {
		err = servicesManager.CreateVirtualRepository().Npm(params)
	}
	return err
}

func virtualBowerHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewBowerVirtualRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateVirtualRepository().Bower(params)
	} else {
		err = servicesManager.CreateVirtualRepository().Bower(params)
	}
	return err
}

func virtualDebianHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewDebianVirtualRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateVirtualRepository().Debian(params)
	} else {
		err = servicesManager.CreateVirtualRepository().Debian(params)
	}
	return err
}

func virtualPypiHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewPypiVirtualRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateVirtualRepository().Pypi(params)
	} else {
		err = servicesManager.CreateVirtualRepository().Pypi(params)
	}
	return err
}

func virtualDockerHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewDockerVirtualRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateVirtualRepository().Docker(params)
	} else {
		err = servicesManager.CreateVirtualRepository().Docker(params)
	}
	return err
}

func virtualGitlfsHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewGitlfsVirtualRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateVirtualRepository().Gitlfs(params)
	} else {
		err = servicesManager.CreateVirtualRepository().Gitlfs(params)
	}
	return err
}

func virtualGoHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewGoVirtualRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateVirtualRepository().Go(params)
	} else {
		err = servicesManager.CreateVirtualRepository().Go(params)
	}
	return err
}

func virtualConanHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewConanVirtualRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateVirtualRepository().Conan(params)
	} else {
		err = servicesManager.CreateVirtualRepository().Conan(params)
	}
	return err
}

func virtualChefHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewChefVirtualRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateVirtualRepository().Chef(params)
	} else {
		err = servicesManager.CreateVirtualRepository().Chef(params)
	}
	return err
}

func virtualPuppetHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewPuppetVirtualRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateVirtualRepository().Puppet(params)
	} else {
		err = servicesManager.CreateVirtualRepository().Puppet(params)
	}
	return err
}

func virtualYumHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewYumVirtualRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateVirtualRepository().Yum(params)
	} else {
		err = servicesManager.CreateVirtualRepository().Yum(params)
	}
	return err
}

func virtualP2Handler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewP2VirtualRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateVirtualRepository().P2(params)
	} else {
		err = servicesManager.CreateVirtualRepository().P2(params)
	}
	return err
}

func virtualCondaHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewCondaVirtualRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateVirtualRepository().Conda(params)
	} else {
		err = servicesManager.CreateVirtualRepository().Conda(params)
	}
	return err
}

func virtualGenericHandler(servicesManager artifactory.ArtifactoryServicesManager, jsonConfig []byte, isUpdate bool) error {
	params := services.NewGenericVirtualRepositoryParams()
	err := json.Unmarshal(jsonConfig, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if isUpdate {
		err = servicesManager.UpdateVirtualRepository().Generic(params)
	} else {
		err = servicesManager.CreateVirtualRepository().Generic(params)
	}
	return err
}
