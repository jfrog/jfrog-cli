package repository

import (
	"encoding/json"
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/jfrog/jfrog-cli-go/artifactory/commands/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io/ioutil"
	"strconv"
	"strings"
)

type RepoTemplateCommand struct {
	path string
}

const (
	// Strings for prompt questions
	SelectConfigKeyMsg    = "Select the next configuration key, or type ':x' to exit"
	InsertValueMsg        = "Insert value for %s: "
	CommaSeparatedListMsg = " (as comma-separated list) "

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

	// Boolean answers
	True  = "true"
	False = "false"

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

var commonConfKeys = []prompt.Suggest{
	{Text: utils.WriteAndExist},
	{Text: Description},
	{Text: Notes},
	{Text: IncludePatterns},
	{Text: ExcludePatterns},
	{Text: RepoLayoutRef},
}

var localRemoteConfKeys = []prompt.Suggest{
	{Text: HandleReleases},
	{Text: HandleSnapshots},
	{Text: MaxUniqueSnapshots},
	{Text: SuppressPomConsistencyChecks},
	{Text: BlackedOut},
	{Text: DownloadRedirect},
	{Text: BlockPushingSchema1},
}

var localVirtualConfKeys = []prompt.Suggest{
	{Text: DebianTrivialLayout},
}

var remoteVirtualConfKeys = []prompt.Suggest{
	{Text: ExternalDependenciesEnabled},
	{Text: ExternalDependenciesPatterns},
}

var uniqueLocalConfKeys = []prompt.Suggest{
	{Text: ChecksumPolicyType},
	{Text: MaxUniqueTags},
	{Text: SnapshotVersionBehavior},
	{Text: XrayIndex},
	{Text: PropertySets},
	{Text: ArchiveBrowsingEnabled},
	{Text: CalculateYumMetadata},
	{Text: YumRootDepth},
	{Text: DockerApiVersion},
	{Text: EnableFileListsIndexing},
	{Text: OptionalIndexCompressionFormats},
}

func getLocalRepoConfKeys() []prompt.Suggest {
	localKeys := append(commonConfKeys, localRemoteConfKeys...)
	localKeys = append(localKeys, localVirtualConfKeys...)
	return append(localKeys, uniqueLocalConfKeys...)
}

var uniqueRemoteConfKeys = []prompt.Suggest{
	{Text: Username},
	{Text: Password},
	{Text: Proxy},
	{Text: RemoteRepoChecksumPolicyType},
	{Text: HardFail},
	{Text: Offline},
	{Text: StoreArtifactsLocally},
	{Text: SocketTimeoutMillis},
	{Text: LocalAddress},
	{Text: RetrievalCachePeriodSecs},
	{Text: FailedRetrievalCachePeriodSecs},
	{Text: MissedRetrievalCachePeriodSecs},
	{Text: UnusedArtifactsCleanupEnabled},
	{Text: UnusedArtifactsCleanupPeriodHours},
	{Text: AssumedOfflinePeriodSecs},
	{Text: FetchJarsEagerly},
	{Text: FetchSourcesEagerly},
	{Text: ShareConfiguration},
	{Text: SynchronizeProperties},
	{Text: BlockMismatchingMimeTypes},
	{Text: AllowAnyHostAuth},
	{Text: EnableCookieManagement},
	{Text: BowerRegistryUrl},
	{Text: ComposerRegistryUrl},
	{Text: PyPIRegistryUrl},
	{Text: VcsType},
	{Text: VcsGitProvider},
	{Text: VcsGitDownloadUrl},
	{Text: BypassHeadRequests},
	{Text: ClientTlsCertificate},
	{Text: FeedContextPath},
	{Text: DownloadContextPath},
	{Text: V3FeedUrl},
	{Text: ContentSynchronisation},
}

func getRemoteRepoConfKeys() []prompt.Suggest {
	remoteKeys := append(commonConfKeys, localRemoteConfKeys...)
	remoteKeys = append(remoteKeys, remoteVirtualConfKeys...)
	return append(remoteKeys, uniqueRemoteConfKeys...)
}

var uniqueVirtualConfKeys = []prompt.Suggest{
	{Text: Repositories},
	{Text: ArtifactoryRequestsCanRetrieveRemoteArtifacts},
	{Text: KeyPair},
	{Text: PomRepositoryReferencesCleanupPolicy},
	{Text: DefaultDeploymentRepo},
	{Text: ForceMavenAuthentication},
	{Text: ExternalDependenciesRemoteRepo},
}

func getVirtualRepoConfKeys() []prompt.Suggest {
	virtualKeys := append(commonConfKeys, localVirtualConfKeys...)
	virtualKeys = append(virtualKeys, remoteVirtualConfKeys...)
	return append(virtualKeys, uniqueVirtualConfKeys...)
}

func getAllPossibleOptionalRepoConfKeys() []prompt.Suggest {
	allKeys := append(commonConfKeys, prompt.Suggest{Text: PackageType})
	allKeys = append(allKeys, localRemoteConfKeys...)
	allKeys = append(allKeys, localVirtualConfKeys...)
	allKeys = append(allKeys, remoteVirtualConfKeys...)
	allKeys = append(allKeys, uniqueLocalConfKeys...)
	allKeys = append(allKeys, uniqueRemoteConfKeys...)
	allKeys = append(allKeys, prompt.Suggest{Text: Url})
	return append(allKeys, uniqueVirtualConfKeys...)
}

func NewRepoTemplateCommand() *RepoTemplateCommand {
	return &RepoTemplateCommand{}
}

func (rtc *RepoTemplateCommand) SetTemplatePath(path string) *RepoTemplateCommand {
	rtc.path = path
	return rtc
}

func (rtc *RepoTemplateCommand) Run() (err error) {
	repoTemplateQuestionnarie := &utils.InteractiveQuestionnaire{
		MandatoryQuestionsKeys: []string{TemplateType, Key},
		QuestionsMap:           questionMap,
	}
	err = repoTemplateQuestionnarie.Perform()
	if err != nil {
		return err
	}
	resBytes, err := json.Marshal(repoTemplateQuestionnarie.ConfigMap)
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

func writeStringAnswer(resultMap *map[string]interface{}, key, value string) error {
	(*resultMap)[key] = value
	return nil
}

func writeBoolAnswer(resultMap *map[string]interface{}, key, value string) error {
	if regexMatch := utils.VarPattern.FindStringSubmatch(value); regexMatch != nil {
		return writeStringAnswer(resultMap, key, value)
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return err
	}
	(*resultMap)[key] = boolValue
	return nil
}

func writeIntAnswer(resultMap *map[string]interface{}, key, value string) error {
	if regexMatch := utils.VarPattern.FindStringSubmatch(value); regexMatch != nil {
		return writeStringAnswer(resultMap, key, value)
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	(*resultMap)[key] = intValue
	return nil
}

func writeStringArrayAnswer(resultMap *map[string]interface{}, key, value string) error {
	if regexMatch := utils.VarPattern.FindStringSubmatch(value); regexMatch != nil {
		return writeStringAnswer(resultMap, key, value)
	}
	arrValue := strings.Split(value, ",")
	(*resultMap)[key] = arrValue
	return nil
}

var freeStringQuestionInfo = utils.QuestionInfo{
	Options:   nil,
	AllowVars: false,
	Writer:    writeStringAnswer,
}

func getBoolSuggests() []prompt.Suggest {
	return []prompt.Suggest{
		{Text: True},
		{Text: False},
	}
}

var boolQuestionInfo = utils.QuestionInfo{
	Options:   getBoolSuggests(),
	AllowVars: true,
	Writer:    writeBoolAnswer,
}

var intQuestionInfo = utils.QuestionInfo{
	Options:   nil,
	AllowVars: true,
	Writer:    writeIntAnswer,
}

var arrayStringQuestionInfo = utils.QuestionInfo{
	Msg:       "The value should be a comma separated list",
	Options:   nil,
	AllowVars: true,
	Writer:    writeStringArrayAnswer,
}

func templateTypeCallback(iq *utils.InteractiveQuestionnaire, templateType string) (string, error) {
	switch templateType {
	// For creation template rclass and packgeType are mandatory keys
	case Create:
		iq.MandatoryQuestionsKeys = append(iq.MandatoryQuestionsKeys, Rclass, PackageType)
	// For update template packageType is an optional common key to modify, and we have to offer all keys as optional to
	case Update:
		iq.OptionalKeysSuggests = getAllPossibleOptionalRepoConfKeys()
	}
	return "", nil
}

func rclassCallback(iq *utils.InteractiveQuestionnaire, rclass string) (string, error) {
	switch rclass {
	case Remote:
		iq.AskQuestion(iq.QuestionsMap[MandatoryUrl])
		iq.OptionalKeysSuggests = getRemoteRepoConfKeys()
	case Local:
		iq.OptionalKeysSuggests = getLocalRepoConfKeys()
	case Virtual:
		iq.OptionalKeysSuggests = getVirtualRepoConfKeys()

	}
	return "", nil
}

type contentSynchronisation struct {
	Enabled    bool
	Statistics struct{ Enabled bool }
	Properties struct{ Enabled bool }
	Source     struct{ OriginAbsenceDetection bool }
}

func contentSynchronisationCallBack(iq *utils.InteractiveQuestionnaire, enabled string) (value string, err error) {
	var cs contentSynchronisation
	cs.Enabled, err = strconv.ParseBool(enabled)
	if err != nil {
		return "", nil
	}
	enabled = utils.AskFromList("", "Insert the value for statistic.enable >", false, getBoolSuggests())
	cs.Statistics.Enabled, err = strconv.ParseBool(enabled)
	if err != nil {
		return "", nil
	}
	enabled = utils.AskFromList("", "Insert the value for properties.enable >", false, getBoolSuggests())
	cs.Properties.Enabled, err = strconv.ParseBool(enabled)
	if err != nil {
		return "", nil
	}
	enabled = utils.AskFromList("", "Insert the value for source.originAbsenceDetection >", false, getBoolSuggests())
	cs.Source.OriginAbsenceDetection, err = strconv.ParseBool(enabled)
	if err != nil {
		return "", nil
	}
	iq.ConfigMap[ContentSynchronisation] = cs
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
	TemplateType: {
		Options: []prompt.Suggest{
			{Text: Create, Description: "Template for creating a new repository"},
			{Text: Update, Description: "Template for updating an existing repository"},
		},
		Msg:          "Select the template type",
		PromptPrefix: ">",
		AllowVars:    false,
		Writer:       nil,
		MapKey:       "",
		Callback:     templateTypeCallback,
	},
	utils.OptionalKey: {
		Msg:          "Select the next property, or \":x\" to finish",
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
		Writer:       writeStringAnswer,
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
		Writer:       writeStringAnswer,
		MapKey:       Rclass,
		Callback:     rclassCallback,
	},
	PackageType: {
		Options: []prompt.Suggest{
			{Text: Generic},
			{Text: Maven},
			{Text: Gradle},
			{Text: Ivy},
			{Text: Sbt},
			{Text: Helm},
			{Text: Cocoapods},
			{Text: Opkg},
			{Text: Rpm},
			{Text: Nuget},
			{Text: Cran},
			{Text: Gems},
			{Text: Npm},
			{Text: Bower},
			{Text: Debian},
			{Text: Composer},
			{Text: Pypi},
			{Text: Docker},
			{Text: Vagrant},
			{Text: Gitlfs},
			{Text: Go},
			{Text: Yum},
			{Text: Conan},
			{Text: Chef},
			{Text: Puppet},
		},
		Msg:          "Select the repository's package type",
		PromptPrefix: ">",
		AllowVars:    false,
		Writer:       writeStringAnswer,
		MapKey:       PackageType,
		Callback:     nil, // TODO: implement pkgTypeCallback,
	},
	MandatoryUrl: {
		Msg:          "",
		PromptPrefix: "Insert the remote repository URL >",
		AllowVars:    true,
		Writer:       writeStringAnswer,
		MapKey:       Url,
		Callback:     nil,
	},
	Url:             freeStringQuestionInfo,
	Description:     freeStringQuestionInfo,
	Notes:           freeStringQuestionInfo,
	IncludePatterns: freeStringQuestionInfo,
	ExcludePatterns: freeStringQuestionInfo,
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
		Writer:    writeStringAnswer,
	},
	HandleReleases:               boolQuestionInfo,
	HandleSnapshots:              boolQuestionInfo,
	MaxUniqueSnapshots:           intQuestionInfo,
	SuppressPomConsistencyChecks: boolQuestionInfo,
	BlackedOut:                   boolQuestionInfo,
	DownloadRedirect:             boolQuestionInfo,
	BlockPushingSchema1:          boolQuestionInfo,
	DebianTrivialLayout:          boolQuestionInfo,
	ExternalDependenciesEnabled:  boolQuestionInfo,
	ExternalDependenciesPatterns: arrayStringQuestionInfo,
	ChecksumPolicyType: {
		Options: []prompt.Suggest{
			{Text: ClientChecksum},
			{Text: ServerGeneratedChecksums},
		},
		AllowVars: true,
		Writer:    writeStringAnswer,
	},
	MaxUniqueTags: intQuestionInfo,
	SnapshotVersionBehavior: {
		Options: []prompt.Suggest{
			{Text: Unique},
			{Text: NonUnique},
			{Text: Deployer},
		},
		AllowVars: true,
		Writer:    writeStringAnswer,
	},
	XrayIndex:              boolQuestionInfo,
	PropertySets:           arrayStringQuestionInfo,
	ArchiveBrowsingEnabled: boolQuestionInfo,
	CalculateYumMetadata:   boolQuestionInfo,
	YumRootDepth:           intQuestionInfo,
	DockerApiVersion: {
		Options: []prompt.Suggest{
			{Text: V1},
			{Text: V2},
		},
		AllowVars: true,
		Writer:    writeStringAnswer,
	},
	EnableFileListsIndexing: boolQuestionInfo,
	OptionalIndexCompressionFormats: {
		Msg:       "Enter a semicolon separated list of values from " + strings.Join([]string{Bz2, Lzma, Xz}, ","),
		Options:   nil,
		AllowVars: false,
		Writer:    writeStringArrayAnswer,
	},
	Username: freeStringQuestionInfo,
	Password: freeStringQuestionInfo,
	Proxy:    freeStringQuestionInfo,
	RemoteRepoChecksumPolicyType: {
		Options: []prompt.Suggest{
			{Text: GenerateIfAbsent},
			{Text: Fail},
			{Text: IgnoreAndGenerate},
			{Text: PassThru},
		},
		AllowVars: true,
		Writer:    writeStringAnswer,
	},
	HardFail:                          boolQuestionInfo,
	Offline:                           boolQuestionInfo,
	StoreArtifactsLocally:             boolQuestionInfo,
	SocketTimeoutMillis:               intQuestionInfo,
	LocalAddress:                      freeStringQuestionInfo,
	RetrievalCachePeriodSecs:          intQuestionInfo,
	FailedRetrievalCachePeriodSecs:    intQuestionInfo,
	MissedRetrievalCachePeriodSecs:    intQuestionInfo,
	UnusedArtifactsCleanupEnabled:     boolQuestionInfo,
	UnusedArtifactsCleanupPeriodHours: intQuestionInfo,
	AssumedOfflinePeriodSecs:          intQuestionInfo,
	FetchJarsEagerly:                  boolQuestionInfo,
	FetchSourcesEagerly:               boolQuestionInfo,
	ShareConfiguration:                boolQuestionInfo,
	SynchronizeProperties:             boolQuestionInfo,
	BlockMismatchingMimeTypes:         boolQuestionInfo,
	AllowAnyHostAuth:                  boolQuestionInfo,
	EnableCookieManagement:            boolQuestionInfo,
	BowerRegistryUrl:                  freeStringQuestionInfo,
	ComposerRegistryUrl:               freeStringQuestionInfo,
	PyPIRegistryUrl:                   freeStringQuestionInfo,
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
		Writer:    writeStringAnswer,
	},
	VcsGitDownloadUrl:    freeStringQuestionInfo,
	BypassHeadRequests:   boolQuestionInfo,
	ClientTlsCertificate: freeStringQuestionInfo,
	FeedContextPath:      freeStringQuestionInfo,
	DownloadContextPath:  freeStringQuestionInfo,
	V3FeedUrl:            freeStringQuestionInfo,
	ContentSynchronisation: {
		Options:   getBoolSuggests(),
		AllowVars: true,
		Writer:    nil,
		Callback:  contentSynchronisationCallBack,
	},
	Repositories: arrayStringQuestionInfo,
	ArtifactoryRequestsCanRetrieveRemoteArtifacts: boolQuestionInfo,
	KeyPair: freeStringQuestionInfo,
	PomRepositoryReferencesCleanupPolicy: {
		Options: []prompt.Suggest{
			{Text: DiscardActiveRefrence},
			{Text: DiscardAnyReference},
			{Text: Nothing},
		},
		AllowVars: true,
		Writer:    writeStringAnswer,
	},
	DefaultDeploymentRepo:          freeStringQuestionInfo,
	ForceMavenAuthentication:       boolQuestionInfo,
	ExternalDependenciesRemoteRepo: freeStringQuestionInfo,
}
