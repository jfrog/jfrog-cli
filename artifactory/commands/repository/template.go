package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/jfrog/jfrog-cli-go/artifactory/commands/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io/ioutil"
	"strings"
)

type RepoTemplateCommand struct {
	path string
}

const (
	// Strings for prompt questions
	SelectConfigKeyMsg = "Select the next configuration key"
	InsertValueMsg     = "Insert value for %s: "

	// Template types
	TemplateType = "templateType"
	Create       = "create"
	Update       = "update"

	MandatoryUrl = "mandatoryUrl"

	// Common repository configuration JSON keys
	Key             = "key"
	Rclass          = "rclass"
	PackageType     = "packageType"
	Description     = "description"
	Notes           = "notes"
	IncludePatterns = "includesPattern"
	ExcludePatterns = "excludesPattern"
	RepoLayoutRef   = "repoLayoutRef"

	// Mutual local and virtual repository configuration JSON keys
	DebianTrivialLayout = "debianTrivialLayout"

	// Mutual local and remote repository configuration JSON keys
	HandleReleases               = "handleReleases"
	HandleSnapshots              = "handleSnapshots"
	MaxUniqueSnapshots           = "maxUniqueSnapshots"
	SuppressPomConsistencyChecks = "suppressPomConsistencyChecks"
	BlackedOut                   = "blackedOut"
	PropertySets                 = "propertySets"
	DownloadRedirect             = "downloadRedirect"
	BlockPushingSchema1          = "blockPushingSchema1"

	// Mutual remote and virtual repository configuration JSON keys
	ExternalDependenciesEnabled  = "externalDependenciesEnabled"
	ExternalDependenciesPatterns = "externalDependenciesPatterns"

	// Unique local repository configuration JSON keys
	ChecksumPolicyType              = "checksumPolicyType"
	MaxUniqueTags                   = "maxUniqueTags"
	SnapshotVersionBehavior         = "snapshotVersionBehavior"
	XrayIndex                       = "xrayIndex"
	ArchiveBrowsingEnabled          = "archiveBrowsingEnabled"
	CalculateYumMetadata            = "calculateYumMetadata"
	YumRootDepth                    = "yumRootDepth"
	DockerApiVersion                = "dockerApiVersion"
	EnableFileListsIndexing         = "enableFileListsIndexing"
	OptionalIndexCompressionFormats = "optionalIndexCompressionFormats"
	ForceNugetAuthentication        = "forceNugetAuthentication"

	// Unique remote repository configuration JSON keys
	Url                               = "url"
	Username                          = "username"
	Password                          = "password"
	Proxy                             = "proxy"
	RemoteRepoChecksumPolicyType      = "remoteRepoChecksumPolicyType"
	HardFail                          = "hardFail"
	Offline                           = "offline"
	StoreArtifactsLocally             = "storeArtifactsLocally"
	SocketTimeoutMillis               = "socketTimeoutMillis"
	LocalAddress                      = "localAddress"
	RetrievalCachePeriodSecs          = "retrievalCachePeriodSecs"
	FailedRetrievalCachePeriodSecs    = "failedRetrievalCachePeriodSecs"
	MissedRetrievalCachePeriodSecs    = "missedRetrievalCachePeriodSecs"
	UnusedArtifactsCleanupEnabled     = "unusedArtifactsCleanupEnabled"
	UnusedArtifactsCleanupPeriodHours = "unusedArtifactsCleanupPeriodHours"
	AssumedOfflinePeriodSecs          = "assumedOfflinePeriodSecs"
	FetchJarsEagerly                  = "fetchJarsEagerly"
	FetchSourcesEagerly               = "fetchSourcesEagerly"
	RejectInvalidJars                 = "rejectInvalidJars"
	ShareConfiguration                = "shareConfiguration"
	SynchronizeProperties             = "synchronizeProperties"
	BlockMismatchingMimeTypes         = "blockMismatchingMimeTypes"
	AllowAnyHostAuth                  = "allowAnyHostAuth"
	EnableCookieManagement            = "enableCookieManagement"
	BowerRegistryUrl                  = "bowerRegistryUrl"
	ComposerRegistryUrl               = "composerRegistryUrl"
	PyPIRegistryUrl                   = "pyPIRegistryUrl"
	VcsType                           = "vcsType"
	VcsGitProvider                    = "vcsGitProvider"
	VcsGitDownloadUrl                 = "vcsGitDownloadUrl"
	BypassHeadRequests                = "bypassHeadRequests"
	ClientTlsCertificate              = "clientTlsCertificate"
	FeedContextPath                   = "feedContextPath"
	DownloadContextPath               = "downloadContextPath"
	V3FeedUrl                         = "v3FeedUrl"
	ContentSynchronisation            = "contentSynchronisation"
	ListRemoteFolderItems             = "listRemoteFolderItems"
	EnableTokenAuthentication         = "enableTokenAuthentication"
	PodsSpecsRepoUrl                  = "podsSpecsRepoUrl"

	// Unique virtual repository configuration JSON keys
	Repositories                                  = "repositories"
	ArtifactoryRequestsCanRetrieveRemoteArtifacts = "artifactoryRequestsCanRetrieveRemoteArtifacts"
	KeyPair                                       = "keyPair"
	PomRepositoryReferencesCleanupPolicy          = "pomRepositoryReferencesCleanupPolicy"
	DefaultDeploymentRepo                         = "defaultDeploymentRepo"
	ForceMavenAuthentication                      = "forceMavenAuthentication"
	ExternalDependenciesRemoteRepo                = "externalDependenciesRemoteRepo"

	// rclasses
	Local   = "local"
	Remote  = "remote"
	Virtual = "virtual"

	// PackageTypes
	Generic   = "generic"
	Maven     = "maven"
	Gradle    = "gradle"
	Ivy       = "ivy"
	Sbt       = "sbt"
	Helm      = "helm"
	Cocoapods = "cocoapods"
	Opkg      = "opkg"
	Rpm       = "rpm"
	Nuget     = "nuget"
	Cran      = "cran"
	Gems      = "gems"
	Npm       = "npm"
	Bower     = "bower"
	Debian    = "debian"
	Composer  = "composer"
	Pypi      = "pypi"
	Docker    = "docker"
	Vagrant   = "vagrant"
	Gitlfs    = "gitlfs"
	Go        = "go"
	Yum       = "yum"
	Conan     = "conan"
	Chef      = "chef"
	Puppet    = "puppet"
	Vcs       = "vcs"
	Conda     = "conda"
	P2        = "p2"

	// Repo layout Refs
	BowerDefault    = "bower-default"
	buildDefault    = "build-default"
	ComposerDefault = "composer-default"
	ConanDefault    = "conan-default"
	GoDefault       = "go-default"
	GradleDefault   = "gradle-default"
	IvyDefault      = "ivy-default"
	Maven1Default   = "maven-1-default"
	Maven2Default   = "maven-2-default"
	NpmDefault      = "npm-default"
	NugetDefault    = "nuget-default"
	puppetDefault   = "puppet-default"
	SbtDefault      = "sbt-default"
	SimpleDefault   = "simple-default"
	VcsDefault      = "vcs-default"

	// Checksum Policy Types
	ClientChecksum           = "client-checksums"
	ServerGeneratedChecksums = "server-generated-checksums"

	// Snapshot version behaviors
	Unique    = "unique"
	NonUnique = "non-unique"
	Deployer  = "deployer"

	// Optional index compression formats
	Bz2  = "bz2"
	Lzma = "lzma"
	Xz   = "xz"

	// Docker api versions
	V1 = "V1"
	V2 = "V2"

	// Remote repo checksum policy types
	GenerateIfAbsent  = "generate-if-absent"
	Fail              = "fail"
	IgnoreAndGenerate = "ignore-and-generate"
	PassThru          = "pass-thru"

	// Vcs Types
	Git = "GIT"

	// Vcs git provider
	Github      = "GITHUB"
	Bitbucket   = "BITBUCKET"
	Oldstash    = "OLDSTASH"
	Stash       = "STASH"
	Artifactory = "ARTIFACTORY"
	Custom      = "CUSTOM"

	// POM repository references cleanup policies
	DiscardActiveRefrence = "discard_active_reference"
	DiscardAnyReference   = "discard_any_reference"
	Nothing               = "nothing"
)

var optionalSuggestsMap = map[string]prompt.Suggest{
	utils.WriteAndExist:               {Text: utils.WriteAndExist},
	Description:                       {Text: Description},
	Notes:                             {Text: Notes},
	IncludePatterns:                   {Text: IncludePatterns},
	ExcludePatterns:                   {Text: ExcludePatterns},
	RepoLayoutRef:                     {Text: RepoLayoutRef},
	HandleReleases:                    {Text: HandleReleases},
	HandleSnapshots:                   {Text: HandleSnapshots},
	MaxUniqueSnapshots:                {Text: MaxUniqueSnapshots},
	SuppressPomConsistencyChecks:      {Text: SuppressPomConsistencyChecks},
	BlackedOut:                        {Text: BlackedOut},
	DownloadRedirect:                  {Text: DownloadRedirect},
	BlockPushingSchema1:               {Text: BlockPushingSchema1},
	DebianTrivialLayout:               {Text: DebianTrivialLayout},
	ExternalDependenciesEnabled:       {Text: ExternalDependenciesEnabled},
	ExternalDependenciesPatterns:      {Text: ExternalDependenciesPatterns},
	ChecksumPolicyType:                {Text: ChecksumPolicyType},
	MaxUniqueTags:                     {Text: MaxUniqueTags},
	SnapshotVersionBehavior:           {Text: SnapshotVersionBehavior},
	XrayIndex:                         {Text: XrayIndex},
	PropertySets:                      {Text: PropertySets},
	ArchiveBrowsingEnabled:            {Text: ArchiveBrowsingEnabled},
	CalculateYumMetadata:              {Text: CalculateYumMetadata},
	YumRootDepth:                      {Text: YumRootDepth},
	DockerApiVersion:                  {Text: DockerApiVersion},
	EnableFileListsIndexing:           {Text: EnableFileListsIndexing},
	OptionalIndexCompressionFormats:   {Text: OptionalIndexCompressionFormats},
	Username:                          {Text: Username},
	Password:                          {Text: Password},
	Proxy:                             {Text: Proxy},
	RemoteRepoChecksumPolicyType:      {Text: RemoteRepoChecksumPolicyType},
	HardFail:                          {Text: HardFail},
	Offline:                           {Text: Offline},
	StoreArtifactsLocally:             {Text: StoreArtifactsLocally},
	SocketTimeoutMillis:               {Text: SocketTimeoutMillis},
	LocalAddress:                      {Text: LocalAddress},
	RetrievalCachePeriodSecs:          {Text: RetrievalCachePeriodSecs},
	FailedRetrievalCachePeriodSecs:    {Text: FailedRetrievalCachePeriodSecs},
	MissedRetrievalCachePeriodSecs:    {Text: MissedRetrievalCachePeriodSecs},
	UnusedArtifactsCleanupEnabled:     {Text: UnusedArtifactsCleanupEnabled},
	UnusedArtifactsCleanupPeriodHours: {Text: UnusedArtifactsCleanupPeriodHours},
	AssumedOfflinePeriodSecs:          {Text: AssumedOfflinePeriodSecs},
	FetchJarsEagerly:                  {Text: FetchJarsEagerly},
	FetchSourcesEagerly:               {Text: FetchSourcesEagerly},
	RejectInvalidJars:                 {Text: RejectInvalidJars},
	ShareConfiguration:                {Text: ShareConfiguration},
	SynchronizeProperties:             {Text: SynchronizeProperties},
	BlockMismatchingMimeTypes:         {Text: BlockMismatchingMimeTypes},
	AllowAnyHostAuth:                  {Text: AllowAnyHostAuth},
	EnableCookieManagement:            {Text: EnableCookieManagement},
	BowerRegistryUrl:                  {Text: BowerRegistryUrl},
	ComposerRegistryUrl:               {Text: ComposerRegistryUrl},
	PyPIRegistryUrl:                   {Text: PyPIRegistryUrl},
	VcsType:                           {Text: VcsType},
	VcsGitProvider:                    {Text: VcsGitProvider},
	VcsGitDownloadUrl:                 {Text: VcsGitDownloadUrl},
	BypassHeadRequests:                {Text: BypassHeadRequests},
	ClientTlsCertificate:              {Text: ClientTlsCertificate},
	FeedContextPath:                   {Text: FeedContextPath},
	DownloadContextPath:               {Text: DownloadContextPath},
	V3FeedUrl:                         {Text: V3FeedUrl},
	ContentSynchronisation:            {Text: ContentSynchronisation},
	ListRemoteFolderItems:             {Text: ListRemoteFolderItems},
	PodsSpecsRepoUrl:                  {Text: PodsSpecsRepoUrl},
	EnableTokenAuthentication:         {Text: EnableTokenAuthentication},
	Repositories:                      {Text: Repositories},
	ArtifactoryRequestsCanRetrieveRemoteArtifacts: {Text: ArtifactoryRequestsCanRetrieveRemoteArtifacts},
	KeyPair:                              {Text: KeyPair},
	PomRepositoryReferencesCleanupPolicy: {Text: PomRepositoryReferencesCleanupPolicy},
	DefaultDeploymentRepo:                {Text: DefaultDeploymentRepo},
	ForceMavenAuthentication:             {Text: ForceMavenAuthentication},
	ExternalDependenciesRemoteRepo:       {Text: ExternalDependenciesRemoteRepo},
}

var baseLocalRepoConfKeys = []string{
	Description, Notes, IncludePatterns, ExcludePatterns, RepoLayoutRef, BlackedOut, XrayIndex,
	PropertySets, ArchiveBrowsingEnabled, OptionalIndexCompressionFormats, DownloadRedirect, BlockPushingSchema1,
}

var mavenGradleLocalRepoConfKeys = []string{
	MaxUniqueSnapshots, HandleReleases, HandleSnapshots, SuppressPomConsistencyChecks, SnapshotVersionBehavior, ChecksumPolicyType,
}

var rpmLocalRepoConfKeys = []string{
	YumRootDepth, CalculateYumMetadata, EnableFileListsIndexing,
}

var nugetLocalRepoConfKeys = []string{
	MaxUniqueSnapshots, ForceNugetAuthentication,
}

var debianLocalRepoConfKeys = []string{
	DebianTrivialLayout,
}

var dockerLocalRepoConfKeys = []string{
	DockerApiVersion, MaxUniqueTags,
}

var baseRemoteRepoConfKeys = []string{
	Username, Password, Proxy, Description, Notes, IncludePatterns, ExcludePatterns, RepoLayoutRef, HardFail, Offline,
	BlackedOut, StoreArtifactsLocally, SocketTimeoutMillis, LocalAddress, RetrievalCachePeriodSecs, FailedRetrievalCachePeriodSecs,
	MissedRetrievalCachePeriodSecs, UnusedArtifactsCleanupEnabled, UnusedArtifactsCleanupPeriodHours, AssumedOfflinePeriodSecs,
	ShareConfiguration, SynchronizeProperties, BlockMismatchingMimeTypes, PropertySets, AllowAnyHostAuth, EnableCookieManagement,
	BypassHeadRequests, ClientTlsCertificate, DownloadRedirect, BlockPushingSchema1, ContentSynchronisation,
}

var mavenGradleRemoteRepoConfKeys = []string{
	FetchJarsEagerly, FetchSourcesEagerly, RemoteRepoChecksumPolicyType, HandleReleases, HandleSnapshots,
	SuppressPomConsistencyChecks, RejectInvalidJars,
}

var cocapodsRemoteRepoConfKeys = []string{
	PodsSpecsRepoUrl,
}

var opkgRemoteRepoConfKeys = []string{
	ListRemoteFolderItems,
}

var rpmRemoteRepoConfKeys = []string{
	ListRemoteFolderItems,
}

var nugetRemoteRepoConfKeys = []string{
	FeedContextPath, DownloadContextPath, V3FeedUrl, ForceNugetAuthentication,
}

var gemsRemoteRepoConfKeys = []string{
	ListRemoteFolderItems,
}

var npmRemoteRepoConfKeys = []string{
	ListRemoteFolderItems,
}

var bowerRemoteRepoConfKeys = []string{
	BowerRegistryUrl,
}

var debianRemoteRepoConfKeys = []string{
	ListRemoteFolderItems,
}

var composerRemoteRepoConfKeys = []string{
	ComposerRegistryUrl,
}

var pypiRemoteRepoConfKeys = []string{
	PyPIRegistryUrl, ListRemoteFolderItems,
}

var dockerRemoteRepoConfKeys = []string{
	ExternalDependenciesEnabled, ExternalDependenciesPatterns, EnableTokenAuthentication,
}

var gitlfsRemoteRepoConfKeys = []string{
	ListRemoteFolderItems,
}

var vcsRemoteRepoConfKeys = []string{
	VcsGitProvider, VcsType, MaxUniqueSnapshots, VcsGitDownloadUrl, ListRemoteFolderItems,
}

var genericRemoteRepoConfKeys = []string{
	ListRemoteFolderItems,
}

var baseVirtualRepoConfKeys = []string{
	Repositories, Description, Notes, IncludePatterns, ExcludePatterns, RepoLayoutRef, ArtifactoryRequestsCanRetrieveRemoteArtifacts,
	DefaultDeploymentRepo,
}

var mavenGradleVirtualRepoConfKeys = []string{
	ForceMavenAuthentication, PomRepositoryReferencesCleanupPolicy, KeyPair,
}

var nugetVirtualRepoConfKeys = []string{
	ForceNugetAuthentication,
}

var npmVirtualRepoConfKeys = []string{
	ExternalDependenciesEnabled, ExternalDependenciesPatterns, ExternalDependenciesRemoteRepo,
}

var bowerVirtualRepoConfKeys = []string{
	ExternalDependenciesEnabled, ExternalDependenciesPatterns, ExternalDependenciesRemoteRepo,
}

var debianVirtualRepoConfKeys = []string{
	DebianTrivialLayout,
}

var goVirtualRepoConfKeys = []string{
	ExternalDependenciesEnabled, ExternalDependenciesPatterns,
}

func getAllPossibleOptionalRepoConfKeys() []prompt.Suggest {
	//allKeys := append(commonConfKeys, prompt.Suggest{Text: PackageType})
	//allKeys = append(allKeys, localRemoteConfKeys...)
	//allKeys = append(allKeys, localVirtualConfKeys...)
	//allKeys = append(allKeys, remoteVirtualConfKeys...)
	//allKeys = append(allKeys, uniqueLocalConfKeys...)
	//allKeys = append(allKeys, uniqueRemoteConfKeys...)
	//allKeys = append(allKeys, prompt.Suggest{Text: Url})
	//return append(allKeys, uniqueVirtualConfKeys...)
	return nil
}

var localRepoPkgTypes = []string{
	Maven, Gradle, Ivy, Sbt, Helm, Cocoapods, Opkg, Rpm, Nuget, Cran, Gems, Npm, Bower, Debian, Composer, Pypi, Docker,
	Vagrant, Gitlfs, Go, Yum, Conan, Chef, Puppet, Generic,
}

var remoteRepoPkgTypes = []string{
	Maven, Gradle, Ivy, Sbt, Helm, Cocoapods, Opkg, Rpm, Nuget, Cran, Gems, Npm, Bower, Debian, Composer, Pypi, Docker,
	Gitlfs, Go, Yum, Conan, Chef, Puppet, Conda, P2, Vcs, Generic,
}

var virtualRepoPkgTypes = []string{
	Maven, Gradle, Ivy, Sbt, Helm, Rpm, Nuget, Cran, Gems, Npm, Bower, Debian, Pypi, Docker, Gitlfs, Go, Yum, Conan,
	Chef, Puppet, Conda, P2, Generic,
}

var pkgTypeSuggestsMap = map[string]prompt.Suggest{
	Generic:   {Text: Generic},
	Maven:     {Text: Maven},
	Gradle:    {Text: Gradle},
	Ivy:       {Text: Ivy},
	Sbt:       {Text: Sbt},
	Helm:      {Text: Helm},
	Cocoapods: {Text: Cocoapods},
	Opkg:      {Text: Opkg},
	Rpm:       {Text: Rpm},
	Nuget:     {Text: Nuget},
	Cran:      {Text: Cran},
	Gems:      {Text: Gems},
	Npm:       {Text: Npm},
	Bower:     {Text: Bower},
	Debian:    {Text: Debian},
	Composer:  {Text: Composer},
	Pypi:      {Text: Pypi},
	Docker:    {Text: Docker},
	Vagrant:   {Text: Vagrant},
	Gitlfs:    {Text: Gitlfs},
	Go:        {Text: Go},
	Yum:       {Text: Yum},
	Conan:     {Text: Conan},
	Chef:      {Text: Chef},
	Puppet:    {Text: Puppet},
	Vcs:       {Text: Vcs},
	Conda:     {Text: Conda},
	P2:        {Text: P2},
}

func NewRepoTemplateCommand() *RepoTemplateCommand {
	return &RepoTemplateCommand{}
}

func (rtc *RepoTemplateCommand) SetTemplatePath(path string) *RepoTemplateCommand {
	rtc.path = path
	return rtc
}

func (rtc *RepoTemplateCommand) Run() (err error) {
	repoTemplateQuestionnaire := &utils.InteractiveQuestionnaire{
		MandatoryQuestionsKeys: []string{Key, Rclass},
		QuestionsMap:           questionMap,
	}
	err = repoTemplateQuestionnaire.Perform()
	if err != nil {
		return err
	}
	resBytes, err := json.Marshal(repoTemplateQuestionnaire.ConfigMap)
	if err != nil {
		return errorutils.CheckError(err)
	}
	if err = ioutil.WriteFile(rtc.path, resBytes, 0644); err != nil {
		return errorutils.CheckError(err)
	}
	log.Info(fmt.Sprintf("Repository creation config template successfully created at %s.", rtc.path))

	return nil
}

func (rtc *RepoTemplateCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	// Since it's a local command, usage won't be reported.
	return nil, nil
}

func (rtc *RepoTemplateCommand) CommandName() string {
	return "rt_repo_template"
}

func rclassCallback(iq *utils.InteractiveQuestionnaire, rclass string) (string, error) {
	var pkgTypes []string
	switch rclass {
	case Remote:
		iq.AskQuestion(iq.QuestionsMap[MandatoryUrl])
		pkgTypes = remoteRepoPkgTypes
	case Local:
		pkgTypes = localRepoPkgTypes
	case Virtual:
		pkgTypes = virtualRepoPkgTypes
	default:
		return "", errors.New("unsupported rclass")
	}
	var pkgTypeQuestion = utils.QuestionInfo{
		Options:      utils.GetSuggestsFromKeys(pkgTypes, pkgTypeSuggestsMap),
		Msg:          "Select the repository's package type",
		PromptPrefix: ">",
		AllowVars:    false,
		Writer:       utils.WriteStringAnswer,
		MapKey:       PackageType,
		Callback:     pkgTypeCallback,
	}
	return iq.AskQuestion(pkgTypeQuestion)
}

func pkgTypeCallback(iq *utils.InteractiveQuestionnaire, pkgType string) (string, error) {
	rclass := iq.ConfigMap[Rclass]
	switch rclass {
	case Remote:
		iq.OptionalKeysSuggests = getRemoteRepoConfKeys(pkgType)
	case Local:
		iq.OptionalKeysSuggests = getLocalRepoConfKeys(pkgType)
	case Virtual:
		iq.OptionalKeysSuggests = getVirtualRepoConfKeys(pkgType)
	default:
		return "", errors.New("unsupported rclass was configured")
	}
	return "", nil
}

func getRemoteRepoConfKeys(pkgType string) []prompt.Suggest {
	optionalKeys := []string{utils.WriteAndExist}
	optionalKeys = append(optionalKeys, baseRemoteRepoConfKeys...)
	switch pkgType {
	case Gradle:
	case Maven:
		optionalKeys = append(optionalKeys, mavenGradleRemoteRepoConfKeys...)
	case Cocoapods:
		optionalKeys = append(optionalKeys, cocapodsRemoteRepoConfKeys...)
	case Opkg:
		optionalKeys = append(optionalKeys, opkgRemoteRepoConfKeys...)
	case Rpm:
		optionalKeys = append(optionalKeys, rpmRemoteRepoConfKeys...)
	case Nuget:
		optionalKeys = append(optionalKeys, nugetRemoteRepoConfKeys...)
	case Gems:
		optionalKeys = append(optionalKeys, gemsRemoteRepoConfKeys...)
	case Npm:
		optionalKeys = append(optionalKeys, npmRemoteRepoConfKeys...)
	case Bower:
		optionalKeys = append(optionalKeys, bowerRemoteRepoConfKeys...)
	case Debian:
		optionalKeys = append(optionalKeys, debianRemoteRepoConfKeys...)
	case Composer:
		optionalKeys = append(optionalKeys, composerRemoteRepoConfKeys...)
	case Pypi:
		optionalKeys = append(optionalKeys, pypiRemoteRepoConfKeys...)
	case Docker:
		optionalKeys = append(optionalKeys, dockerLocalRepoConfKeys...)
	case Gitlfs:
		optionalKeys = append(optionalKeys, gitlfsRemoteRepoConfKeys...)
	case Vcs:
		optionalKeys = append(optionalKeys, vcsRemoteRepoConfKeys...)
	}
	return utils.GetSuggestsFromKeys(optionalKeys, optionalSuggestsMap)
}

func getVirtualRepoConfKeys(pkgType string) []prompt.Suggest {
	optionalKeys := []string{utils.WriteAndExist}
	optionalKeys = append(optionalKeys, baseVirtualRepoConfKeys...)
	switch pkgType {
	case Gradle:
	case Maven:
		optionalKeys = append(optionalKeys, mavenGradleVirtualRepoConfKeys...)
	case Nuget:
		optionalKeys = append(optionalKeys, nugetVirtualRepoConfKeys...)
	case Npm:
		optionalKeys = append(optionalKeys, npmVirtualRepoConfKeys...)
	case Bower:
		optionalKeys = append(optionalKeys, bowerVirtualRepoConfKeys...)
	case Debian:
		optionalKeys = append(optionalKeys, debianVirtualRepoConfKeys...)
	case Go:
		optionalKeys = append(optionalKeys, goVirtualRepoConfKeys...)
	}
	return utils.GetSuggestsFromKeys(optionalKeys, optionalSuggestsMap)
}

func getLocalRepoConfKeys(pkgType string) []prompt.Suggest {
	optionalKeys := []string{utils.WriteAndExist}
	optionalKeys = append(optionalKeys, baseLocalRepoConfKeys...)
	switch pkgType {
	case Gradle:
	case Maven:
		optionalKeys = append(optionalKeys, mavenGradleLocalRepoConfKeys...)
	case Rpm:
		optionalKeys = append(optionalKeys, rpmLocalRepoConfKeys...)
	case Nuget:
		optionalKeys = append(optionalKeys, nugetLocalRepoConfKeys...)
	case Debian:
		optionalKeys = append(optionalKeys, debianLocalRepoConfKeys...)
	case Docker:
		optionalKeys = append(optionalKeys, dockerLocalRepoConfKeys...)
	}
	return utils.GetSuggestsFromKeys(optionalKeys, optionalSuggestsMap)
}

func contentSynchronisationCallBack(iq *utils.InteractiveQuestionnaire, answer string) (value string, err error) {
	//var cs contentSynchronisation
	//cs.Enabled, err = strconv.ParseBool(enabled)
	if err != nil {
		return "", nil
	}
	answer += "," + utils.AskFromList("", "Insert the value for statistic.enable >", false, utils.GetBoolSuggests())
	//cs.Statistics.Enabled, err = strconv.ParseBool(enabled)
	if err != nil {
		return "", nil
	}
	answer += "," + utils.AskFromList("", "Insert the value for properties.enable >", false, utils.GetBoolSuggests())
	//cs.Properties.Enabled, err = strconv.ParseBool(enabled)
	if err != nil {
		return "", nil
	}
	answer += "," + utils.AskFromList("", "Insert the value for source.originAbsenceDetection >", false, utils.GetBoolSuggests())
	//cs.Source.OriginAbsenceDetection, err = strconv.ParseBool(enabled)
	if err != nil {
		return "", nil
	}
	iq.ConfigMap[ContentSynchronisation] = answer
	return "", nil
}

func optionalKeyCallback(iq *utils.InteractiveQuestionnaire, key string) (value string, err error) {
	if key != utils.WriteAndExist {
		valueQuestion := iq.QuestionsMap[key]
		valueQuestion.MapKey = key
		valueQuestion.PromptPrefix = "Insert the value for " + key
		if valueQuestion.Options != nil {
			valueQuestion.PromptPrefix += utils.PressTabMsg
		}
		valueQuestion.PromptPrefix += " >"
		value, err = iq.AskQuestion(valueQuestion)
	}
	return value, err
}

var questionMap = map[string]utils.QuestionInfo{
	utils.OptionalKey: {
		Msg:          "Select the next property",
		PromptPrefix: ">",
		AllowVars:    false,
		Writer:       nil,
		MapKey:       "",
		Callback:     optionalKeyCallback,
	},
	Key: {
		Msg:          "Insert the repository key",
		PromptPrefix: ">",
		AllowVars:    true,
		Writer:       utils.WriteStringAnswer,
		MapKey:       Key,
		Callback:     nil,
	},
	Rclass: {
		Options: []prompt.Suggest{
			{Text: Local, Description: "A physical, locally-managed repositories into which you can deploy artifacts"},
			{Text: Remote, Description: "A caching proxy for a repository managed at a remote URL"},
			{Text: Virtual, Description: "An Aggregation of several repositories with the same package type under a common URL."},
		},
		Msg:          "Select the repository class",
		PromptPrefix: ">",
		AllowVars:    false,
		Writer:       utils.WriteStringAnswer,
		MapKey:       Rclass,
		Callback:     rclassCallback,
	},
	MandatoryUrl: {
		Msg:          "",
		PromptPrefix: "Insert the remote repository URL >",
		AllowVars:    true,
		Writer:       utils.WriteStringAnswer,
		MapKey:       Url,
		Callback:     nil,
	},
	Url:             utils.FreeStringQuestionInfo,
	Description:     utils.FreeStringQuestionInfo,
	Notes:           utils.FreeStringQuestionInfo,
	IncludePatterns: utils.StringListQuestionInfo,
	ExcludePatterns: utils.StringListQuestionInfo,
	RepoLayoutRef: {
		Options: []prompt.Suggest{
			{Text: BowerDefault},
			{Text: buildDefault},
			{Text: ComposerDefault},
			{Text: ConanDefault},
			{Text: GoDefault},
			{Text: GradleDefault},
			{Text: IvyDefault},
			{Text: Maven1Default},
			{Text: Maven2Default},
			{Text: NpmDefault},
			{Text: NugetDefault},
			{Text: puppetDefault},
			{Text: SbtDefault},
			{Text: SimpleDefault},
			{Text: VcsDefault},
		},
		AllowVars: true,
		Writer:    utils.WriteStringAnswer,
	},
	HandleReleases:               utils.BoolQuestionInfo,
	HandleSnapshots:              utils.BoolQuestionInfo,
	MaxUniqueSnapshots:           utils.IntQuestionInfo,
	SuppressPomConsistencyChecks: utils.BoolQuestionInfo,
	BlackedOut:                   utils.BoolQuestionInfo,
	DownloadRedirect:             utils.BoolQuestionInfo,
	BlockPushingSchema1:          utils.BoolQuestionInfo,
	DebianTrivialLayout:          utils.BoolQuestionInfo,
	ExternalDependenciesEnabled:  utils.BoolQuestionInfo,
	ExternalDependenciesPatterns: utils.StringListQuestionInfo,
	ChecksumPolicyType: {
		Options: []prompt.Suggest{
			{Text: ClientChecksum},
			{Text: ServerGeneratedChecksums},
		},
		AllowVars: true,
		Writer:    utils.WriteStringAnswer,
	},
	MaxUniqueTags: utils.IntQuestionInfo,
	SnapshotVersionBehavior: {
		Options: []prompt.Suggest{
			{Text: Unique},
			{Text: NonUnique},
			{Text: Deployer},
		},
		AllowVars: true,
		Writer:    utils.WriteStringAnswer,
	},
	XrayIndex:              utils.BoolQuestionInfo,
	PropertySets:           utils.StringListQuestionInfo,
	ArchiveBrowsingEnabled: utils.BoolQuestionInfo,
	CalculateYumMetadata:   utils.BoolQuestionInfo,
	YumRootDepth:           utils.IntQuestionInfo,
	DockerApiVersion: {
		Options: []prompt.Suggest{
			{Text: V1},
			{Text: V2},
		},
		AllowVars: true,
		Writer:    utils.WriteStringAnswer,
	},
	EnableFileListsIndexing: utils.BoolQuestionInfo,
	OptionalIndexCompressionFormats: {
		Msg:       "Enter a comma separated list of values from " + strings.Join([]string{Bz2, Lzma, Xz}, ","),
		Options:   nil,
		AllowVars: false,
		Writer:    utils.WriteStringArrayAnswer,
	},
	Username: utils.FreeStringQuestionInfo,
	Password: utils.FreeStringQuestionInfo,
	Proxy:    utils.FreeStringQuestionInfo,
	RemoteRepoChecksumPolicyType: {
		Options: []prompt.Suggest{
			{Text: GenerateIfAbsent},
			{Text: Fail},
			{Text: IgnoreAndGenerate},
			{Text: PassThru},
		},
		AllowVars: true,
		Writer:    utils.WriteStringAnswer,
	},
	HardFail:                          utils.BoolQuestionInfo,
	Offline:                           utils.BoolQuestionInfo,
	StoreArtifactsLocally:             utils.BoolQuestionInfo,
	SocketTimeoutMillis:               utils.IntQuestionInfo,
	LocalAddress:                      utils.FreeStringQuestionInfo,
	RetrievalCachePeriodSecs:          utils.IntQuestionInfo,
	FailedRetrievalCachePeriodSecs:    utils.IntQuestionInfo,
	MissedRetrievalCachePeriodSecs:    utils.IntQuestionInfo,
	UnusedArtifactsCleanupEnabled:     utils.BoolQuestionInfo,
	UnusedArtifactsCleanupPeriodHours: utils.IntQuestionInfo,
	AssumedOfflinePeriodSecs:          utils.IntQuestionInfo,
	FetchJarsEagerly:                  utils.BoolQuestionInfo,
	FetchSourcesEagerly:               utils.BoolQuestionInfo,
	RejectInvalidJars:                 utils.BoolQuestionInfo,
	ShareConfiguration:                utils.BoolQuestionInfo,
	SynchronizeProperties:             utils.BoolQuestionInfo,
	BlockMismatchingMimeTypes:         utils.BoolQuestionInfo,
	AllowAnyHostAuth:                  utils.BoolQuestionInfo,
	EnableCookieManagement:            utils.BoolQuestionInfo,
	BowerRegistryUrl:                  utils.FreeStringQuestionInfo,
	ComposerRegistryUrl:               utils.FreeStringQuestionInfo,
	PyPIRegistryUrl:                   utils.FreeStringQuestionInfo,
	//	VcsType : {[]string{Git}, writeStringAnswer},
	VcsGitProvider: {
		Options: []prompt.Suggest{
			{Text: Github},
			{Text: Bitbucket},
			{Text: Oldstash},
			{Text: Stash},
			{Text: Artifactory},
			{Text: Custom},
		},
		AllowVars: true,
		Writer:    utils.WriteStringAnswer,
	},
	VcsGitDownloadUrl:         utils.FreeStringQuestionInfo,
	BypassHeadRequests:        utils.BoolQuestionInfo,
	ClientTlsCertificate:      utils.FreeStringQuestionInfo,
	FeedContextPath:           utils.FreeStringQuestionInfo,
	DownloadContextPath:       utils.FreeStringQuestionInfo,
	V3FeedUrl:                 utils.FreeStringQuestionInfo,
	ListRemoteFolderItems:     utils.BoolQuestionInfo,
	EnableTokenAuthentication: utils.BoolQuestionInfo,
	PodsSpecsRepoUrl:          utils.FreeStringQuestionInfo,
	ContentSynchronisation: {
		Options:   utils.GetBoolSuggests(),
		AllowVars: true,
		Writer:    nil,
		Callback:  contentSynchronisationCallBack,
	},
	Repositories: utils.StringListQuestionInfo,
	ArtifactoryRequestsCanRetrieveRemoteArtifacts: utils.BoolQuestionInfo,
	KeyPair: utils.FreeStringQuestionInfo,
	PomRepositoryReferencesCleanupPolicy: {
		Options: []prompt.Suggest{
			{Text: DiscardActiveRefrence},
			{Text: DiscardAnyReference},
			{Text: Nothing},
		},
		AllowVars: true,
		Writer:    utils.WriteStringAnswer,
	},
	DefaultDeploymentRepo:          utils.FreeStringQuestionInfo,
	ForceMavenAuthentication:       utils.BoolQuestionInfo,
	ExternalDependenciesRemoteRepo: utils.FreeStringQuestionInfo,
}
