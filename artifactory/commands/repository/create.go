package repository

import (
	"encoding/json"
	"errors"
	"github.com/jfrog/jfrog-cli-go/artifactory/commands/utils"
	rtUtils "github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"strconv"
	"strings"
)

type RepoCreateCommand struct {
	rtDetails    *config.ArtifactoryDetails
	templatePath string
	vars         string
}

func NewRepoCreateCommand() *RepoCreateCommand {
	return &RepoCreateCommand{}
}

func (rcc *RepoCreateCommand) SetTemplatePath(path string) *RepoCreateCommand {
	rcc.templatePath = path
	return rcc
}

func (rcc *RepoCreateCommand) SetVars(vars string) *RepoCreateCommand {
	rcc.vars = vars
	return rcc
}

func (rcc *RepoCreateCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *RepoCreateCommand {
	rcc.rtDetails = rtDetails
	return rcc
}

func (rcc *RepoCreateCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return rcc.rtDetails, nil
}

func (rcc *RepoCreateCommand) CommandName() string {
	return "rt_repo_create"
}

func (rcc *RepoCreateCommand) Run() (err error) {
	content, err := fileutils.ReadFile(rcc.templatePath)
	if errorutils.CheckError(err) != nil {
		return
	}
	if len(rcc.vars) > 0 {
		templateVars := cliutils.SpecVarsStringToMap(rcc.vars)
		content = cliutils.ReplaceSpecVars(content, templateVars)
	}
	var repoConfigMap map[string]interface{}
	err = json.Unmarshal(content, &repoConfigMap)
	if errorutils.CheckError(err) != nil {
		return
	}

	for key, value := range repoConfigMap {
		writertsMap[key](&repoConfigMap, key, value.(string))
	}
	content, err = json.Marshal(repoConfigMap)
	switch repoConfigMap[Rclass] {
	case Local:
		err = localRepoCreators[repoConfigMap[PackageType].(string)](rcc.rtDetails, content)
	case Remote:
		err = remoteRepoCreators[repoConfigMap[PackageType].(string)](rcc.rtDetails, content)
	default:
		return errors.New("unsupported rclass")
	}
	return err
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

var writertsMap = map[string]utils.AnswerWriter{
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

type repoCreator func(*config.ArtifactoryDetails, []byte) error

var localRepoCreators = map[string]repoCreator{
	Maven:     createLocalMaven,
	Gradle:    createLocalGradle,
	Ivy:       createLocalIvy,
	Sbt:       createLocalSbt,
	Helm:      createLocalHelm,
	Cocoapods: createLocalCocapods,
	Opkg:      createLocalOpkg,
	Rpm:       createLocalRpm,
	Nuget:     createLocalNuget,
	Cran:      createLocalCran,
	Gems:      createLocalGems,
	Npm:       createLocalNpm,
	Bower:     createLocalBower,
	Debian:    createLocalDebian,
	Composer:  createLocalComposer,
	Pypi:      createLocalPypi,
	Docker:    createLocalDocker,
	Vagrant:   createLocalVagrant,
	Gitlfs:    createLocalGitlfs,
	Go:        createLocalGo,
	Yum:       createLocalYum,
	Conan:     createLocalConan,
	Chef:      createLocalChef,
	Puppet:    createLocalPuppet,
	Generic:   createLocalGeneric,
}

func createLocalMaven(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.MavenGradleLocalRepositoryParams
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateLocalRepository().Maven(params)
}

func createLocalGradle(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.MavenGradleLocalRepositoryParams
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateLocalRepository().Maven(params)
}

func createLocalIvy(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.IvyLocalRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateLocalRepository().Ivy(params)
}

func createLocalSbt(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.SbtLocalRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateLocalRepository().Sbt(params)
}

func createLocalHelm(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.HelmLocalRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateLocalRepository().Helm(params)
}

func createLocalCocapods(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.CocapodsLocalRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateLocalRepository().Cocapods(params)
}

func createLocalOpkg(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.OpkgLocalRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateLocalRepository().Opkg(params)
}

func createLocalRpm(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.RpmLocalRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateLocalRepository().Rpm(params)
}

func createLocalNuget(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.NugetLocalRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateLocalRepository().Nuget(params)
}

func createLocalCran(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.CranLocalRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateLocalRepository().Cran(params)
}

func createLocalGems(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.GemsLocalRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateLocalRepository().Gems(params)
}

func createLocalNpm(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.NpmLocalRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateLocalRepository().Npm(params)
}

func createLocalBower(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.BowerLocalRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateLocalRepository().Bower(params)
}

func createLocalDebian(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.DebianLocalRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateLocalRepository().Debian(params)
}

func createLocalComposer(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.ComposerLocalRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateLocalRepository().Composer(params)
}

func createLocalPypi(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.PypiLocalRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateLocalRepository().Pypi(params)
}

func createLocalDocker(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.DockerLocalRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateLocalRepository().Docker(params)
}

func createLocalVagrant(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.VagrantLocalRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateLocalRepository().Vagrant(params)
}

func createLocalGitlfs(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.GitlfsLocalRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateLocalRepository().Gitlfs(params)
}

func createLocalGo(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.GoLocalRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateLocalRepository().Go(params)
}

func createLocalYum(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.YumLocalRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateLocalRepository().Yum(params)
}

func createLocalConan(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.ConanLocalRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateLocalRepository().Conan(params)
}

func createLocalChef(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.ChefLocalRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateLocalRepository().Chef(params)
}

func createLocalPuppet(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.PuppetLocalRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateLocalRepository().Puppet(params)
}

func createLocalGeneric(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.GenericLocalRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateLocalRepository().Generic(params)
}

var remoteRepoCreators = map[string]repoCreator{
	Maven:     createRemoteMaven,
	Gradle:    createRemoteGradle,
	Ivy:       createRemoteIvy,
	Sbt:       createRemoteSbt,
	Helm:      createRemoteHelm,
	Cocoapods: createRemoteCocapods,
	Opkg:      createRemoteOpkg,
	Rpm:       createRemoteRpm,
	Nuget:     createRemoteNuget,
	Cran:      createRemoteCran,
	Gems:      createRemoteGems,
	Npm:       createRemoteNpm,
	Bower:     createRemoteBower,
	Debian:    createRemoteDebian,
	Composer:  createRemoteComposer,
	Pypi:      createRemotelPypi,
	Docker:    createRemoteDocker,
	Gitlfs:    createRemoteGitlfs,
	Go:        createRemoteGo,
	Yum:       createRemoteYum,
	Conan:     createRemoteConan,
	Chef:      createRemoteChef,
	Puppet:    createRemotePuppet,
	Conda:     createRemoteConda,
	P2:        createRemoteP2,
	Vcs:       createRemoteVcs,
	Generic:   createRemoteGeneric,
}

func createRemoteMaven(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.MavenGradleRemoteRepositoryParams
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateRemoteRepository().Maven(params)
}

func createRemoteGradle(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.MavenGradleRemoteRepositoryParams
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateRemoteRepository().Gradle(params)
}

func createRemoteIvy(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.IvyRemoteRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateRemoteRepository().Ivy(params)
}

func createRemoteSbt(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.SbtRemoteRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateRemoteRepository().Sbt(params)
}

func createRemoteHelm(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.HelmRemoteRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateRemoteRepository().Helm(params)
}

func createRemoteCocapods(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.CocapodsRemoteRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateRemoteRepository().Cocapods(params)
}

func createRemoteOpkg(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.OpkgRemoteRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateRemoteRepository().Opkg(params)
}

func createRemoteRpm(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.RpmRemoteRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateRemoteRepository().Rpm(params)
}

func createRemoteNuget(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.NugetRemoteRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateRemoteRepository().Nuget(params)
}

func createRemoteCran(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.CranRemoteRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateRemoteRepository().Cran(params)
}

func createRemoteGems(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.GemsRemoteRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateRemoteRepository().Gems(params)
}

func createRemoteNpm(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.NpmRemoteRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateRemoteRepository().Npm(params)
}

func createRemoteBower(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.BowerRemoteRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateRemoteRepository().Bower(params)
}

func createRemoteDebian(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.DebianRemoteRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateRemoteRepository().Debian(params)
}

func createRemoteComposer(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.ComposerRemoteRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateRemoteRepository().Composer(params)
}

func createRemotelPypi(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.PypiRemoteRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateRemoteRepository().Pypi(params)
}

func createRemoteDocker(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.DockerRemoteRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateRemoteRepository().Docker(params)
}

func createRemoteGitlfs(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.GitlfsRemoteRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateRemoteRepository().Gitlfs(params)
}

func createRemoteGo(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.GoRemoteRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateRemoteRepository().Go(params)
}

func createRemoteConan(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.ConanRemoteRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateRemoteRepository().Conan(params)
}

func createRemoteChef(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.ChefRemoteRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateRemoteRepository().Chef(params)
}

func createRemotePuppet(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.PuppetRemoteRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateRemoteRepository().Puppet(params)
}

func createRemoteVcs(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.VcsRemoteRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateRemoteRepository().Vcs(params)
}

func createRemoteP2(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.P2RemoteRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateRemoteRepository().P2(params)
}

func createRemoteConda(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.CondaRemoteRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateRemoteRepository().Conda(params)
}

func createRemoteYum(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.YumRemoteRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateRemoteRepository().Yum(params)
}

func createRemoteGeneric(rtDetails *config.ArtifactoryDetails, jsonConfig []byte) error {
	var params services.GenericRemoteRepositoryParam
	err := json.Unmarshal(jsonConfig, &params)
	if err != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	return servicesManager.CreateRemoteRepository().Generic(params)
}
