package cliutils

import (
	"fmt"
	"sort"
	"strconv"

	commonCliUtils "github.com/jfrog/jfrog-cli-core/v2/common/cliutils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"

	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

const (
	// CLI base commands keys
	Setup = "setup"
	Intro = "intro"

	// Artifactory's Commands Keys
	DeleteConfig           = "delete-config"
	Upload                 = "upload"
	Download               = "download"
	Move                   = "move"
	Copy                   = "copy"
	Delete                 = "delete"
	Properties             = "properties"
	Search                 = "search"
	BuildPublish           = "build-publish"
	BuildAppend            = "build-append"
	BuildScanLegacy        = "build-scan-legacy"
	BuildPromote           = "build-promote"
	BuildDiscard           = "build-discard"
	BuildAddDependencies   = "build-add-dependencies"
	BuildAddGit            = "build-add-git"
	BuildCollectEnv        = "build-collect-env"
	GitLfsClean            = "git-lfs-clean"
	Mvn                    = "mvn"
	MvnConfig              = "mvn-config"
	CocoapodsConfig        = "cocoapods-config"
	SwiftConfig            = "swift-config"
	Gradle                 = "gradle"
	GradleConfig           = "gradle-config"
	DockerPromote          = "docker-promote"
	Docker                 = "docker"
	DockerPush             = "docker-push"
	DockerPull             = "docker-pull"
	ContainerPull          = "container-pull"
	ContainerPush          = "container-push"
	BuildDockerCreate      = "build-docker-create"
	OcStartBuild           = "oc-start-build"
	NpmConfig              = "npm-config"
	Npm                    = "npm"
	NpmInstallCi           = "npm-install-ci"
	NpmPublish             = "npm-publish"
	PnpmConfig             = "pnpm-config"
	YarnConfig             = "yarn-config"
	Yarn                   = "yarn"
	NugetConfig            = "nuget-config"
	Nuget                  = "nuget"
	Dotnet                 = "dotnet"
	DotnetConfig           = "dotnet-config"
	Go                     = "go"
	GoConfig               = "go-config"
	GoPublish              = "go-publish"
	Pip                    = "pip"
	PipInstall             = "pip-install"
	PipConfig              = "pip-config"
	TerraformConfig        = "terraform-config"
	Terraform              = "terraform"
	Twine                  = "twine"
	Pipenv                 = "pipenv"
	PipenvConfig           = "pipenv-config"
	PipenvInstall          = "pipenv-install"
	PoetryConfig           = "poetry-config"
	Poetry                 = "poetry"
	Ping                   = "ping"
	RtCurl                 = "rt-curl"
	TemplateConsumer       = "template-consumer"
	RepoDelete             = "repo-delete"
	ReplicationDelete      = "replication-delete"
	PermissionTargetDelete = "permission-target-delete"
	// #nosec G101 -- False positive - no hardcoded credentials.
	ArtifactoryAccessTokenCreate = "artifactory-access-token-create"
	UserCreate                   = "user-create"
	UsersCreate                  = "users-create"
	UsersDelete                  = "users-delete"
	GroupCreate                  = "group-create"
	GroupAddUsers                = "group-add-users"
	GroupDelete                  = "group-delete"
	TransferConfig               = "transfer-config"
	TransferConfigMerge          = "transfer-config-merge"
	passphrase                   = "passphrase"

	// Distribution's Command Keys
	ReleaseBundleV1Create     = "release-bundle-v1-create"
	ReleaseBundleV1Update     = "release-bundle-v1-update"
	ReleaseBundleV1Sign       = "release-bundle-v1-sign"
	ReleaseBundleV1Distribute = "release-bundle-v1-distribute"
	ReleaseBundleV1Delete     = "release-bundle-v1-delete"

	// MC's Commands Keys
	McConfig       = "mc-config"
	LicenseAcquire = "license-acquire"
	LicenseDeploy  = "license-deploy"
	LicenseRelease = "license-release"
	JpdAdd         = "jpd-add"
	JpdDelete      = "jpd-delete"

	// Config commands keys
	AddConfig  = "config-add"
	EditConfig = "config-edit"

	// Project commands keys
	InitProject = "project-init"

	// TransferFiles commands keys
	TransferFiles = "transfer-files"

	// TransferInstall commands keys
	TransferInstall = "transfer-plugin-install"

	// Lifecycle commands keys
	ReleaseBundleCreate       = "release-bundle-create"
	ReleaseBundlePromote      = "release-bundle-promote"
	ReleaseBundleDistribute   = "release-bundle-distribute"
	ReleaseBundleDeleteLocal  = "release-bundle-delete-local"
	ReleaseBundleDeleteRemote = "release-bundle-delete-remote"
	ReleaseBundleExport       = "release-bundle-export"
	ReleaseBundleImport       = "release-bundle-import"

	// Access Token Create commands keys
	AccessTokenCreate = "access-token-create"

	// *** Artifactory Commands' flags ***
	// Base flags
	url         = "url"
	platformUrl = "platform-url"
	user        = "user"
	password    = "password"
	accessToken = "access-token"
	serverId    = "server-id"

	passwordStdin    = "password-stdin"
	accessTokenStdin = "access-token-stdin"

	// Ssh flags
	sshKeyPath    = "ssh-key-path"
	sshPassphrase = "ssh-passphrase"

	// Client certification flags
	ClientCertPath    = "client-cert-path"
	ClientCertKeyPath = "client-cert-key-path"
	InsecureTls       = "insecure-tls"

	// Sort & limit flags
	sortBy    = "sort-by"
	sortOrder = "sort-order"
	limit     = "limit"
	offset    = "offset"

	// Spec flags
	specFlag = "spec"
	specVars = "spec-vars"

	// Build info flags
	buildName   = "build-name"
	buildNumber = "build-number"
	module      = "module"

	// Generic commands flags
	exclusions              = "exclusions"
	recursive               = "recursive"
	flat                    = "flat"
	build                   = "build"
	excludeArtifacts        = "exclude-artifacts"
	includeDeps             = "include-deps"
	regexpFlag              = "regexp"
	retries                 = "retries"
	retryWaitTime           = "retry-wait-time"
	dryRun                  = "dry-run"
	explode                 = "explode"
	bypassArchiveInspection = "bypass-archive-inspection"
	includeDirs             = "include-dirs"
	props                   = "props"
	targetProps             = "target-props"
	excludeProps            = "exclude-props"
	failNoOp                = "fail-no-op"
	threads                 = "threads"
	syncDeletes             = "sync-deletes"
	quiet                   = "quiet"
	bundle                  = "bundle"
	publicGpgKey            = "gpg-key"
	archiveEntries          = "archive-entries"
	detailedSummary         = "detailed-summary"
	archive                 = "archive"
	syncDeletesQuiet        = syncDeletes + "-" + quiet
	antFlag                 = "ant"
	fromRt                  = "from-rt"
	transitive              = "transitive"
	Status                  = "status"
	MinSplit                = "min-split"
	SplitCount              = "split-count"
	ChunkSize               = "chunk-size"

	// Config flags
	interactive   = "interactive"
	EncPassword   = "enc-password"
	BasicAuthOnly = "basic-auth-only"
	Overwrite     = "overwrite"

	// Unique upload flags
	uploadPrefix      = "upload-"
	uploadExclusions  = uploadPrefix + exclusions
	uploadRecursive   = uploadPrefix + recursive
	uploadFlat        = uploadPrefix + flat
	uploadRegexp      = uploadPrefix + regexpFlag
	uploadExplode     = uploadPrefix + explode
	uploadTargetProps = uploadPrefix + targetProps
	uploadSyncDeletes = uploadPrefix + syncDeletes
	uploadArchive     = uploadPrefix + archive
	uploadMinSplit    = uploadPrefix + MinSplit
	uploadSplitCount  = uploadPrefix + SplitCount
	deb               = "deb"
	symlinks          = "symlinks"
	uploadAnt         = uploadPrefix + antFlag

	// Unique download flags
	downloadPrefix       = "download-"
	downloadRecursive    = downloadPrefix + recursive
	downloadFlat         = downloadPrefix + flat
	downloadExplode      = downloadPrefix + explode
	downloadProps        = downloadPrefix + props
	downloadExcludeProps = downloadPrefix + excludeProps
	downloadSyncDeletes  = downloadPrefix + syncDeletes
	downloadMinSplit     = downloadPrefix + MinSplit
	downloadSplitCount   = downloadPrefix + SplitCount
	validateSymlinks     = "validate-symlinks"
	skipChecksum         = "skip-checksum"

	// Unique move flags
	movePrefix       = "move-"
	moveRecursive    = movePrefix + recursive
	moveFlat         = movePrefix + flat
	moveProps        = movePrefix + props
	moveExcludeProps = movePrefix + excludeProps

	// Unique copy flags
	copyPrefix       = "copy-"
	copyRecursive    = copyPrefix + recursive
	copyFlat         = copyPrefix + flat
	copyProps        = copyPrefix + props
	copyExcludeProps = copyPrefix + excludeProps

	// Unique delete flags
	deletePrefix       = "delete-"
	deleteRecursive    = deletePrefix + recursive
	deleteProps        = deletePrefix + props
	deleteExcludeProps = deletePrefix + excludeProps
	deleteQuiet        = deletePrefix + quiet

	// Unique search flags
	searchInclude      = "include"
	searchPrefix       = "search-"
	searchRecursive    = searchPrefix + recursive
	searchProps        = searchPrefix + props
	searchExcludeProps = searchPrefix + excludeProps
	count              = "count"
	searchTransitive   = searchPrefix + transitive

	// Unique properties flags
	propertiesPrefix  = "props-"
	propsRecursive    = propertiesPrefix + recursive
	propsProps        = propertiesPrefix + props
	propsExcludeProps = propertiesPrefix + excludeProps

	// Unique go publish flags
	goPublishExclusions = GoPublish + exclusions

	// Unique build-publish flags
	buildPublishPrefix = "bp-"
	bpDryRun           = buildPublishPrefix + dryRun
	bpDetailedSummary  = buildPublishPrefix + detailedSummary
	envInclude         = "env-include"
	envExclude         = "env-exclude"
	buildUrl           = "build-url"
	Project            = "project"

	// Unique build-add-dependencies flags
	badPrefix    = "bad-"
	badDryRun    = badPrefix + dryRun
	badRecursive = badPrefix + recursive
	badRegexp    = badPrefix + regexpFlag
	badFromRt    = badPrefix + fromRt
	badModule    = badPrefix + module

	// Unique build-add-git flags
	configFlag = "config"

	// Unique build-scan flags
	fail   = "fail"
	rescan = "rescan"

	// Unique build-promote flags
	buildPromotePrefix  = "bpr-"
	bprDryRun           = buildPromotePrefix + dryRun
	bprProps            = buildPromotePrefix + props
	comment             = "comment"
	sourceRepo          = "source-repo"
	includeDependencies = "include-dependencies"
	copyFlag            = "copy"
	failFast            = "fail-fast"

	Async = "async"

	// Unique build-discard flags
	buildDiscardPrefix = "bdi-"
	bdiAsync           = buildDiscardPrefix + Async
	maxDays            = "max-days"
	maxBuilds          = "max-builds"
	excludeBuilds      = "exclude-builds"
	deleteArtifacts    = "delete-artifacts"

	repo = "repo"

	// Unique git-lfs-clean flags
	glcPrefix = "glc-"
	glcDryRun = glcPrefix + dryRun
	glcQuiet  = glcPrefix + quiet
	glcRepo   = glcPrefix + repo
	refs      = "refs"

	// Build tool config flags
	global          = "global"
	serverIdResolve = "server-id-resolve"
	serverIdDeploy  = "server-id-deploy"
	repoResolve     = "repo-resolve"
	repoDeploy      = "repo-deploy"

	// Unique maven-config flags
	repoResolveReleases  = "repo-resolve-releases"
	repoResolveSnapshots = "repo-resolve-snapshots"
	repoDeployReleases   = "repo-deploy-releases"
	repoDeploySnapshots  = "repo-deploy-snapshots"
	includePatterns      = "include-patterns"
	excludePatterns      = "exclude-patterns"

	// Unique gradle-config flags
	usesPlugin          = "uses-plugin"
	UseWrapper          = "use-wrapper"
	deployMavenDesc     = "deploy-maven-desc"
	deployIvyDesc       = "deploy-ivy-desc"
	ivyDescPattern      = "ivy-desc-pattern"
	ivyArtifactsPattern = "ivy-artifacts-pattern"

	// Build tool flags
	deploymentThreads = "deployment-threads"
	skipLogin         = "skip-login"

	// Unique docker promote flags
	dockerPromotePrefix = "docker-promote-"
	targetDockerImage   = "target-docker-image"
	sourceTag           = "source-tag"
	targetTag           = "target-tag"
	dockerPromoteCopy   = dockerPromotePrefix + Copy

	// Unique build docker create
	imageFile = "image-file"

	// Unique oc start-build flags
	ocStartBuildPrefix = "oc-start-build-"
	ocStartBuildRepo   = ocStartBuildPrefix + repo

	// Unique npm flags
	npmPrefix          = "npm-"
	npmDetailedSummary = npmPrefix + detailedSummary

	// Unique nuget/dotnet config flags
	nugetV2 = "nuget-v2"

	// Unique go flags
	noFallback = "no-fallback"

	// Unique Terraform flags
	namespace = "namespace"
	provider  = "provider"
	tag       = "tag"

	// Template user flags
	vars = "vars"

	// User Management flags
	csv            = "csv"
	usersCreateCsv = "users-create-csv"
	usersDeleteCsv = "users-delete-csv"
	UsersGroups    = "users-groups"
	Replace        = "replace"
	Admin          = "admin"

	// Mutual *-access-token-create flags
	Groups      = "groups"
	GrantAdmin  = "grant-admin"
	Expiry      = "expiry"
	Refreshable = "refreshable"
	Audience    = "audience"

	// Unique artifactory-access-token-create flags
	artifactoryAccessTokenCreatePrefix = "rt-atc-"
	rtAtcGroups                        = artifactoryAccessTokenCreatePrefix + Groups
	rtAtcGrantAdmin                    = artifactoryAccessTokenCreatePrefix + GrantAdmin
	rtAtcExpiry                        = artifactoryAccessTokenCreatePrefix + Expiry
	rtAtcRefreshable                   = artifactoryAccessTokenCreatePrefix + Refreshable
	rtAtcAudience                      = artifactoryAccessTokenCreatePrefix + Audience

	// Unique access-token-create flags
	accessTokenCreatePrefix = "atc-"
	atcProject              = accessTokenCreatePrefix + Project
	Scope                   = "scope"
	atcScope                = accessTokenCreatePrefix + Scope
	Description             = "description"
	atcDescription          = accessTokenCreatePrefix + Description
	Reference               = "reference"
	atcReference            = accessTokenCreatePrefix + Reference
	atcGroups               = accessTokenCreatePrefix + Groups
	atcGrantAdmin           = accessTokenCreatePrefix + GrantAdmin
	atcExpiry               = accessTokenCreatePrefix + Expiry
	atcRefreshable          = accessTokenCreatePrefix + Refreshable
	atcAudience             = accessTokenCreatePrefix + Audience

	// Unique Xray Flags for upload/publish commands
	xrayScan = "scan"

	// Unique config transfer flags
	Force            = "force"
	Verbose          = "verbose"
	SourceWorkingDir = "source-working-dir"
	TargetWorkingDir = "target-working-dir"

	// *** Distribution Commands' flags ***
	// Base flags
	distUrl = "dist-url"

	// Unique release-bundle-* v1 flags
	releaseBundleV1Prefix = "rbv1-"
	rbDryRun              = releaseBundleV1Prefix + dryRun
	rbRepo                = releaseBundleV1Prefix + repo
	rbPassphrase          = releaseBundleV1Prefix + passphrase
	distTarget            = releaseBundleV1Prefix + target
	rbDetailedSummary     = releaseBundleV1Prefix + detailedSummary
	sign                  = "sign"
	desc                  = "desc"
	releaseNotesPath      = "release-notes-path"
	releaseNotesSyntax    = "release-notes-syntax"
	deleteFromDist        = "delete-from-dist"

	// Common release-bundle-* v1&v2 flags
	DistRules      = "dist-rules"
	site           = "site"
	city           = "city"
	countryCodes   = "country-codes"
	sync           = "sync"
	maxWaitMinutes = "max-wait-minutes"
	CreateRepo     = "create-repo"

	// *** Xray Commands' flags ***
	// Base flags
	xrUrl = "xr-url"

	// Unique offline-update flags
	licenseId = "license-id"
	from      = "from"
	to        = "to"
	Version   = "version"
	target    = "target"
	Stream    = "stream"
	Periodic  = "periodic"

	// Unique scan flags
	scanPrefix          = "scan-"
	scanRecursive       = scanPrefix + recursive
	scanRegexp          = scanPrefix + regexpFlag
	scanAnt             = scanPrefix + antFlag
	xrOutput            = "format"
	BypassArchiveLimits = "bypass-archive-limits"

	// Audit commands
	auditPrefix                  = "audit-"
	useWrapperAudit              = auditPrefix + UseWrapper
	ExcludeTestDeps              = "exclude-test-deps"
	DepType                      = "dep-type"
	ThirdPartyContextualAnalysis = "third-party-contextual-analysis"
	RequirementsFile             = "requirements-file"
	watches                      = "watches"
	workingDirs                  = "working-dirs"
	ExclusionsAudit              = auditPrefix + exclusions
	repoPath                     = "repo-path"
	licenses                     = "licenses"
	vuln                         = "vuln"
	ExtendedTable                = "extended-table"
	MinSeverity                  = "min-severity"
	FixableOnly                  = "fixable-only"
	// *** Mission Control Commands' flags ***
	missionControlPrefix = "mc-"
	curationThreads      = "curation-threads"
	curationOutput       = "curation-format"

	// Authentication flags
	mcUrl         = missionControlPrefix + url
	mcAccessToken = missionControlPrefix + accessToken

	// Unique config flags
	mcInteractive = missionControlPrefix + interactive

	// Unique license-deploy flags
	licenseCount = "license-count"

	// *** Config Commands' flags ***
	configPrefix      = "config-"
	configPlatformUrl = configPrefix + url
	configRtUrl       = "artifactory-url"
	configDistUrl     = "distribution-url"
	configXrUrl       = "xray-url"
	configMcUrl       = "mission-control-url"
	configPlUrl       = "pipelines-url"
	configAccessToken = configPrefix + accessToken
	configUser        = configPrefix + user
	configPassword    = configPrefix + password
	configInsecureTls = configPrefix + InsecureTls

	// *** Project Commands' flags ***
	projectPath = "path"

	// *** Completion Commands' flags ***
	Completion = "completion"
	Install    = "install"

	// *** TransferFiles Commands' flags ***
	transferFilesPrefix = "transfer-files-"
	Filestore           = "filestore"
	IgnoreState         = "ignore-state"
	ProxyKey            = "proxy-key"
	transferFilesStatus = transferFilesPrefix + "status"
	Stop                = "stop"
	PreChecks           = "prechecks"

	// Transfer flags
	IncludeRepos    = "include-repos"
	ExcludeRepos    = "exclude-repos"
	IncludeProjects = "include-projects"
	ExcludeProjects = "exclude-projects"

	// *** JFrog Pipelines Commands' flags ***
	// Base flags
	branch       = "branch"
	Trigger      = "trigger"
	pipelineName = "pipeline-name"
	name         = "name"
	Validate     = "validate"
	Resources    = "resources"
	monitor      = "monitor"
	repository   = "repository"
	singleBranch = "single-branch"
	Sync         = "sync"
	SyncStatus   = "sync-status"

	// *** TransferInstall Commands' flags ***
	installPluginPrefix  = "install-"
	installPluginVersion = installPluginPrefix + Version
	InstallPluginSrcDir  = "dir"
	InstallPluginHomeDir = "home-dir"

	// Unique lifecycle flags
	lifecyclePrefix      = "lc-"
	lcSync               = lifecyclePrefix + Sync
	lcProject            = lifecyclePrefix + Project
	Builds               = "builds"
	lcBuilds             = lifecyclePrefix + Builds
	ReleaseBundles       = "release-bundles"
	lcReleaseBundles     = lifecyclePrefix + ReleaseBundles
	SigningKey           = "signing-key"
	lcSigningKey         = lifecyclePrefix + SigningKey
	PathMappingPattern   = "mapping-pattern"
	lcPathMappingPattern = lifecyclePrefix + PathMappingPattern
	PathMappingTarget    = "mapping-target"
	lcPathMappingTarget  = lifecyclePrefix + PathMappingTarget
	lcDryRun             = lifecyclePrefix + dryRun
	lcIncludeRepos       = lifecyclePrefix + IncludeRepos
	lcExcludeRepos       = lifecyclePrefix + ExcludeRepos
)

var flagsMap = map[string]cli.Flag{
	// Common commands flags
	platformUrl: cli.StringFlag{
		Name:  url,
		Usage: "[Optional] JFrog platform URL. (example: https://acme.jfrog.io)` `",
	},
	user: cli.StringFlag{
		Name:  user,
		Usage: "[Optional] JFrog username.` `",
	},
	password: cli.StringFlag{
		Name:  password,
		Usage: "[Optional] JFrog password.` `",
	},
	accessToken: cli.StringFlag{
		Name:  accessToken,
		Usage: "[Optional] JFrog access token.` `",
	},
	serverId: cli.StringFlag{
		Name:  serverId,
		Usage: "[Optional] Server ID configured using the 'jf config' command.` `",
	},
	passwordStdin: cli.BoolFlag{
		Name:  passwordStdin,
		Usage: "[Default: false] Set to true if you'd like to provide the password via stdin.` `",
	},
	accessTokenStdin: cli.BoolFlag{
		Name:  accessTokenStdin,
		Usage: "[Default: false] Set to true if you'd like to provide the access token via stdin.` `",
	},
	// Artifactory's commands flags
	url: cli.StringFlag{
		Name:  url,
		Usage: "[Optional] JFrog Artifactory URL. (example: https://acme.jfrog.io/artifactory)` `",
	},
	sshKeyPath: cli.StringFlag{
		Name:  sshKeyPath,
		Usage: "[Optional] SSH key file path.` `",
	},
	sshPassphrase: cli.StringFlag{
		Name:  sshPassphrase,
		Usage: "[Optional] SSH key passphrase.` `",
	},
	ClientCertPath: cli.StringFlag{
		Name:  ClientCertPath,
		Usage: "[Optional] Client certificate file in PEM format.` `",
	},
	ClientCertKeyPath: cli.StringFlag{
		Name:  ClientCertKeyPath,
		Usage: "[Optional] Private key file for the client certificate in PEM format.` `",
	},
	sortBy: cli.StringFlag{
		Name:  sortBy,
		Usage: fmt.Sprintf("[Optional] List of semicolon-separated(;) fields to sort by. The fields must be part of the 'items' AQL domain. For more information, see %sjfrog-artifactory-documentation/artifactory-query-language` `", coreutils.JFrogHelpUrl),
	},
	sortOrder: cli.StringFlag{
		Name:  sortOrder,
		Usage: "[Default: asc] The order by which fields in the 'sort-by' option should be sorted. Accepts 'asc' or 'desc'.` `",
	},
	limit: cli.StringFlag{
		Name:  limit,
		Usage: "[Optional] The maximum number of items to fetch. Usually used with the 'sort-by' option.` `",
	},
	offset: cli.StringFlag{
		Name:  offset,
		Usage: "[Optional] The offset from which to fetch items (i.e. how many items should be skipped). Usually used with the 'sort-by' option.` `",
	},
	specFlag: cli.StringFlag{
		Name:  specFlag,
		Usage: "[Optional] Path to a File Spec.` `",
	},
	specVars: cli.StringFlag{
		Name:  specVars,
		Usage: "[Optional] List of semicolon-separated(;) variables in the form of \"key1=value1;key2=value2;...\" (wrapped by quotes) to be replaced in the File Spec. In the File Spec, the variables should be used as follows: ${key1}.` `",
	},
	buildName: cli.StringFlag{
		Name:  buildName,
		Usage: "[Optional] Providing this option will collect and record build info for this build name. Build number option is mandatory when this option is provided.` `",
	},
	buildNumber: cli.StringFlag{
		Name:  buildNumber,
		Usage: "[Optional] Providing this option will collect and record build info for this build number. Build name option is mandatory when this option is provided.` `",
	},
	module: cli.StringFlag{
		Name:  module,
		Usage: "[Optional] Optional module name for the build-info. Build name and number options are mandatory when this option is provided.` `",
	},
	exclusions: cli.StringFlag{
		Name:  exclusions,
		Usage: "[Optional] List of semicolon-separated(;) exclusions. Exclusions can include the * and the ? wildcards.` `",
	},
	uploadExclusions: cli.StringFlag{
		Name:  exclusions,
		Usage: "[Optional] List of semicolon-separated(;) exclude patterns. Exclude patterns may contain the * and the ? wildcards or a regex pattern, according to the value of the 'regexp' option.` `",
	},
	build: cli.StringFlag{
		Name:  build,
		Usage: "[Optional] If specified, only artifacts of the specified build are matched. The property format is build-name/build-number. If you do not specify the build number, the artifacts are filtered by the latest build number. If the build is assigned to a specific project please provide the project key using the --project flag.` `",
	},
	excludeArtifacts: cli.StringFlag{
		Name:  excludeArtifacts,
		Usage: "[Default: false] If specified, build artifacts are not matched. Used together with the --build flag.` `",
	},
	includeDeps: cli.StringFlag{
		Name:  includeDeps,
		Usage: "[Default: false] If specified, also dependencies of the specified build are matched. Used together with the --build flag.` `",
	},
	includeDirs: cli.BoolFlag{
		Name:  includeDirs,
		Usage: "[Default: false] Set to true if you'd like to also apply the source path pattern for directories and not just for files.` `",
	},
	failNoOp: cli.BoolFlag{
		Name:  failNoOp,
		Usage: "[Default: false] Set to true if you'd like the command to return exit code 2 in case of no files are affected.` `",
	},
	threads: cli.StringFlag{
		Name:  threads,
		Value: "",
		Usage: "[Default: " + strconv.Itoa(commonCliUtils.Threads) + "] Number of working threads.` `",
	},
	retries: cli.StringFlag{
		Name:  retries,
		Usage: "[Default: " + strconv.Itoa(Retries) + "] Number of HTTP retries.` `",
	},
	retryWaitTime: cli.StringFlag{
		Name:  retryWaitTime,
		Usage: "[Default: 0] Number of seconds or milliseconds to wait between retries. The numeric value should either end with s for seconds or ms for milliseconds (for example: 10s or 100ms).` `",
	},
	InsecureTls: cli.BoolFlag{
		Name:  InsecureTls,
		Usage: "[Default: false] Set to true to skip TLS certificates verification.` `",
	},
	bundle: cli.StringFlag{
		Name:  bundle,
		Usage: "[Optional] If specified, only artifacts of the specified bundle are matched. The value format is bundle-name/bundle-version.` `",
	},
	publicGpgKey: cli.StringFlag{
		Name:  publicGpgKey,
		Usage: "[Optional] Path to the public GPG key file located on the file system, used to validate downloaded release bundles.` `",
	},
	archiveEntries: cli.StringFlag{
		Name:  archiveEntries,
		Usage: "[Optional] This option is no longer supported since version 7.90.5 of Artifactory. If specified, only archive artifacts containing entries matching this pattern are matched. You can use wildcards to specify multiple artifacts.` `",
	},
	detailedSummary: cli.BoolFlag{
		Name:  detailedSummary,
		Usage: "[Default: false] Set to true to include a list of the affected files in the command summary.` `",
	},
	interactive: cli.BoolTFlag{
		Name:  interactive,
		Usage: "[Default: true, unless $CI is true] Set to false if you do not want the config command to be interactive. If true, the --url option becomes optional.` `",
	},
	EncPassword: cli.BoolTFlag{
		Name:  EncPassword,
		Usage: "[Default: true] If set to false then the configured password will not be encrypted using Artifactory's encryption API.` `",
	},
	Overwrite: cli.BoolFlag{
		Name:  Overwrite,
		Usage: "[Default: false] Overwrites the instance configuration if an instance with the same ID already exists.` `",
	},
	BasicAuthOnly: cli.BoolFlag{
		Name: BasicAuthOnly,
		Usage: "[Default: false] Set to true to disable replacing username and password/API key with an automatically created access token that's refreshed hourly. " +
			"Username and password/API key will still be used with commands which use external tools or the JFrog Distribution service. " +
			"Can only be passed along with username and password/API key options.` `",
	},
	deb: cli.StringFlag{
		Name:  deb,
		Usage: "[Optional] Used for Debian packages in the form of distribution/component/architecture. If the value for distribution, component or architecture includes a slash, the slash should be escaped with a back-slash.` `",
	},
	uploadRecursive: cli.BoolTFlag{
		Name:  recursive,
		Usage: "[Default: true] Set to false if you do not wish to collect artifacts in sub-folders to be uploaded to Artifactory.` `",
	},
	uploadFlat: cli.BoolFlag{
		Name:  flat,
		Usage: "[Default: false] If set to false, files are uploaded according to their file system hierarchy.` `",
	},
	uploadRegexp: cli.BoolFlag{
		Name:  regexpFlag,
		Usage: "[Default: false] Set to true to use a regular expression instead of wildcards expression to collect files to upload.` `",
	},
	uploadAnt: cli.BoolFlag{
		Name:  antFlag,
		Usage: "[Default: false] Set to true to use an ant pattern instead of wildcards expression to collect files to upload.` `",
	},
	dryRun: cli.BoolFlag{
		Name:  dryRun,
		Usage: "[Default: false] Set to true to disable communication with Artifactory.` `",
	},
	uploadExplode: cli.BoolFlag{
		Name:  explode,
		Usage: "[Default: false] Set to true to extract an archive after it is deployed to Artifactory.` `",
	},
	symlinks: cli.BoolFlag{
		Name:  symlinks,
		Usage: "[Default: false] Set to true to preserve symbolic links structure in Artifactory.` `",
	},
	uploadTargetProps: cli.StringFlag{
		Name:  targetProps,
		Usage: "[Optional] List of semicolon-separated(;) properties in the form of \"key1=value1;key2=value2;...\". Those properties will be attached to the uploaded artifacts.` `",
	},
	uploadSyncDeletes: cli.StringFlag{
		Name:  syncDeletes,
		Usage: "[Optional] Specific path in Artifactory, under which to sync artifacts after the upload. After the upload, this path will include only the artifacts uploaded during this upload operation. The other files under this path will be deleted.` `",
	},
	uploadArchive: cli.StringFlag{
		Name:  archive,
		Usage: "[Optional] Set to \"zip\" to pack and deploy the files to Artifactory inside a ZIP archive. Currently, the only packaging format supported is zip.` `",
	},
	uploadMinSplit: cli.StringFlag{
		Name:  MinSplit,
		Usage: "[Default: " + strconv.Itoa(UploadMinSplitMb) + "] The minimum file size in MiB required to attempt a multi-part upload. This option, as well as the functionality of multi-part upload, requires Artifactory with S3 or GCP storage.` `",
	},
	uploadSplitCount: cli.StringFlag{
		Name:  SplitCount,
		Usage: "[Default: " + strconv.Itoa(UploadSplitCount) + "] The maximum number of parts that can be concurrently uploaded per file during a multi-part upload. Set to 0 to disable multi-part upload. This option, as well as the functionality of multi-part upload, requires Artifactory with S3 or GCP storage.` `",
	},
	ChunkSize: cli.StringFlag{
		Name:  ChunkSize,
		Usage: "[Default: " + strconv.Itoa(UploadChunkSizeMb) + "] The upload chunk size in MiB that can be concurrently uploaded during a multi-part upload. This option, as well as the functionality of multi-part upload, requires Artifactory with S3 or GCP storage.` `",
	},
	syncDeletesQuiet: cli.BoolFlag{
		Name:  quiet,
		Usage: "[Default: $CI] Set to true to skip the sync-deletes confirmation message.` `",
	},
	downloadRecursive: cli.BoolTFlag{
		Name:  recursive,
		Usage: "[Default: true] Set to false if you do not wish to include the download of artifacts inside sub-folders in Artifactory.` `",
	},
	downloadFlat: cli.BoolFlag{
		Name:  flat,
		Usage: "[Default: false] Set to true if you do not wish to have the Artifactory repository path structure created locally for your downloaded files.` `",
	},
	downloadMinSplit: cli.StringFlag{
		Name:  MinSplit,
		Value: "",
		Usage: "[Default: " + strconv.Itoa(DownloadMinSplitKb) + "] Minimum file size in KB to split into ranges when downloading. Set to -1 for no splits.` `",
	},
	skipChecksum: cli.BoolFlag{
		Name:  skipChecksum,
		Usage: "[Default: false] Set to true to skip checksum verification when downloading.` `",
	},
	downloadSplitCount: cli.StringFlag{
		Name:  SplitCount,
		Value: "",
		Usage: "[Default: " + strconv.Itoa(DownloadSplitCount) + "] Number of parts to split a file when downloading. Set to 0 for no splits.` `",
	},
	downloadExplode: cli.BoolFlag{
		Name:  explode,
		Usage: "[Default: false] Set to true to extract an archive after it is downloaded from Artifactory.` `",
	},
	bypassArchiveInspection: cli.BoolFlag{
		Name:  bypassArchiveInspection,
		Usage: "[Default: false] Set to true to bypass the archive security inspection before it is unarchived. Used with the 'explode' option.` `",
	},
	validateSymlinks: cli.BoolFlag{
		Name:  validateSymlinks,
		Usage: "[Default: false] Set to true to perform a checksum validation when downloading symbolic links.` `",
	},
	downloadProps: cli.StringFlag{
		Name:  props,
		Usage: "[Optional] List of semicolon-separated(;) properties in the form of \"key1=value1;key2=value2;...\". Only artifacts with these properties will be downloaded.` `",
	},
	downloadExcludeProps: cli.StringFlag{
		Name:  excludeProps,
		Usage: "[Optional] List of semicolon-separated(;) properties in the form of \"key1=value1;key2=value2;...\". Only artifacts without the specified properties will be downloaded.` `",
	},
	downloadSyncDeletes: cli.StringFlag{
		Name:  syncDeletes,
		Usage: "[Optional] Specific path in the local file system, under which to sync dependencies after the download. After the download, this path will include only the dependencies downloaded during this download operation. The other files under this path will be deleted.` `",
	},
	moveRecursive: cli.BoolTFlag{
		Name:  recursive,
		Usage: "[Default: true] Set to false if you do not wish to move artifacts inside sub-folders in Artifactory.` `",
	},
	moveFlat: cli.BoolFlag{
		Name:  flat,
		Usage: "[Default: false] If set to false, files are moved according to their file system hierarchy.` `",
	},
	moveProps: cli.StringFlag{
		Name:  props,
		Usage: "[Optional] List of semicolon-separated(;) properties in the form of \"key1=value1;key2=value2;...\". Only artifacts with these properties will be moved.` `",
	},
	moveExcludeProps: cli.StringFlag{
		Name:  excludeProps,
		Usage: "[Optional] List of semicolon-separated(;) properties in the form of \"key1=value1;key2=value2;...\". Only artifacts without the specified properties will be moved.` `",
	},
	copyRecursive: cli.BoolTFlag{
		Name:  recursive,
		Usage: "[Default: true] Set to false if you do not wish to copy artifacts inside sub-folders in Artifactory.` `",
	},
	copyFlat: cli.BoolFlag{
		Name:  flat,
		Usage: "[Default: false] If set to false, files are copied according to their file system hierarchy.` `",
	},
	copyProps: cli.StringFlag{
		Name:  props,
		Usage: "[Optional] List of semicolon-separated(;) properties in the form of \"key1=value1;key2=value2;...\". Only artifacts with these properties will be copied.` `",
	},
	copyExcludeProps: cli.StringFlag{
		Name:  excludeProps,
		Usage: "[Optional] List of semicolon-separated(;) properties in the form of \"key1=value1;key2=value2;...\". Only artifacts without the specified properties will be copied.` `",
	},
	deleteRecursive: cli.BoolTFlag{
		Name:  recursive,
		Usage: "[Default: true] Set to false if you do not wish to delete artifacts inside sub-folders in Artifactory.` `",
	},
	deleteProps: cli.StringFlag{
		Name:  props,
		Usage: "[Optional] List of semicolon-separated(;) properties in the form of \"key1=value1;key2=value2;...\". Only artifacts with these properties will be deleted.` `",
	},
	deleteExcludeProps: cli.StringFlag{
		Name:  excludeProps,
		Usage: "[Optional] List of semicolon-separated(;) properties in the form of \"key1=value1;key2=value2;...\". Only artifacts without the specified properties will be deleted.` `",
	},
	deleteQuiet: cli.BoolFlag{
		Name:  quiet,
		Usage: "[Default: $CI] Set to true to skip the delete confirmation message.` `",
	},
	searchRecursive: cli.BoolTFlag{
		Name:  recursive,
		Usage: "[Default: true] Set to false if you do not wish to search artifacts inside sub-folders in Artifactory.` `",
	},
	count: cli.BoolFlag{
		Name:  count,
		Usage: "[Optional] Set to true to display only the total of files or folders found.` `",
	},
	searchProps: cli.StringFlag{
		Name:  props,
		Usage: "[Optional] List of semicolon-separated(;) properties in the form of \"key1=value1;key2=value2;...\". Only artifacts with these properties will be returned.` `",
	},
	searchExcludeProps: cli.StringFlag{
		Name:  excludeProps,
		Usage: "[Optional] List of semicolon-separated(;) properties in the form of \"key1=value1;key2=value2;...\". Only artifacts without the specified properties will be returned` `",
	},
	searchTransitive: cli.BoolFlag{
		Name:  transitive,
		Usage: "[Default: false] Set to true to look for artifacts also in remote repositories. The search will run on the first five remote repositories within the virtual repository. Available on Artifactory version 7.17.0 or higher.` `",
	},
	searchInclude: cli.StringFlag{
		Name:  searchInclude,
		Usage: fmt.Sprintf("[Optional] List of semicolon-separated(;) fields in the form of \"value1;value2;...\". Only the path and the fields that are specified will be returned. The fields must be part of the 'items' AQL domain. For the full supported items list, check %sjfrog-artifactory-documentation/artifactory-query-language` `", coreutils.JFrogHelpUrl),
	},
	propsRecursive: cli.BoolTFlag{
		Name:  recursive,
		Usage: "[Default: true] When false, artifacts inside sub-folders in Artifactory will not be affected.` `",
	},
	propsProps: cli.StringFlag{
		Name:  props,
		Usage: "[Optional] List of semicolon-separated(;) properties in the form of \"key1=value1;key2=value2;...\". Only artifacts with these properties are affected.` `",
	},
	propsExcludeProps: cli.StringFlag{
		Name:  excludeProps,
		Usage: "[Optional] List of semicolon-separated(;) properties in the form of \"key1=value1;key2=value2;...\". Only artifacts without the specified properties are affected` `",
	},
	buildUrl: cli.StringFlag{
		Name:  buildUrl,
		Usage: "[Optional] Can be used for setting the CI server build URL in the build-info.` `",
	},
	Project: cli.StringFlag{
		Name:  Project,
		Usage: "[Optional] JFrog Artifactory project key.` `",
	},
	bpDryRun: cli.BoolFlag{
		Name:  dryRun,
		Usage: "[Default: false] Set to true to get a preview of the recorded build info, without publishing it to Artifactory.` `",
	},
	bpDetailedSummary: cli.BoolFlag{
		Name:  detailedSummary,
		Usage: "[Default: false] Set to true to get a command summary with details about the build info artifact.` `",
	},
	envInclude: cli.StringFlag{
		Name:  envInclude,
		Usage: "[Default: *] List of patterns in the form of \"value1;value2;...\" Only environment variables match those patterns will be included.` `",
	},
	envExclude: cli.StringFlag{
		Name:  envExclude,
		Usage: "[Default: *password*;*psw*;*secret*;*key*;*token*;*auth*] List of case insensitive patterns in the form of \"value1;value2;...\". Environment variables match those patterns will be excluded.` `",
	},
	badRecursive: cli.BoolTFlag{
		Name:  recursive,
		Usage: "[Default: true] Set to false if you do not wish to collect artifacts in sub-folders to be added to the build info.` `",
	},
	badModule: cli.StringFlag{
		Name:  module,
		Usage: "[Optional] Optional module name in the build-info for adding the dependency.` `",
	},
	badRegexp: cli.BoolFlag{
		Name:  regexpFlag,
		Usage: "[Default: false] Set to true to use a regular expression instead of wildcards expression to collect files to be added to the build info.` `",
	},
	badDryRun: cli.BoolFlag{
		Name:  dryRun,
		Usage: "[Default: false] Set to true to only get a summary of the dependencies that will be added to the build info.` `",
	},
	badFromRt: cli.BoolFlag{
		Name:  fromRt,
		Usage: "[Default: false] Set true to search the files in Artifactory, rather than on the local file system. The --regexp option is not supported when --from-rt is set to true.` `",
	},
	configFlag: cli.StringFlag{
		Name:  configFlag,
		Usage: "[Optional] Path to a configuration file.` `",
	},
	fail: cli.BoolTFlag{
		Name:  fail,
		Usage: "[Default: true] Set to false if you do not wish the command to return exit code 3, even if the 'Fail Build' rule is matched by Xray.` `",
	},
	Status: cli.StringFlag{
		Name:  Status,
		Usage: "[Optional] Build promotion status.` `",
	},
	comment: cli.StringFlag{
		Name:  comment,
		Usage: "[Optional] Build promotion comment.` `",
	},
	sourceRepo: cli.StringFlag{
		Name:  sourceRepo,
		Usage: "[Optional] Build promotion source repository.` `",
	},
	includeDependencies: cli.BoolFlag{
		Name:  includeDependencies,
		Usage: "[Default: false] If true, the build dependencies are also promoted.` `",
	},
	copyFlag: cli.BoolFlag{
		Name:  copyFlag,
		Usage: "[Default: false] If true, the build artifacts and dependencies are copied to the target repository, otherwise they are moved.` `",
	},
	failFast: cli.BoolTFlag{
		Name:  failFast,
		Usage: "[Default: true] If true, fail and abort the operation upon receiving an error.` `",
	},
	bprDryRun: cli.BoolFlag{
		Name:  dryRun,
		Usage: "[Default: false] If true, promotion is only simulated. The build is not promoted.` `",
	},
	bprProps: cli.StringFlag{
		Name:  props,
		Usage: "[Optional] List of semicolon-separated(;) properties in the form of \"key1=value1;key2=value2;...\" to be attached to the build artifacts.` `",
	},
	targetDockerImage: cli.StringFlag{
		Name:  "target-docker-image",
		Usage: "[Optional] Docker target image name.` `",
	},
	sourceTag: cli.StringFlag{
		Name:  "source-tag",
		Usage: "[Optional] The tag name to promote.` `",
	},
	targetTag: cli.StringFlag{
		Name:  "target-tag",
		Usage: "[Optional] The target tag to assign the image after promotion.` `",
	},
	dockerPromoteCopy: cli.BoolFlag{
		Name:  "copy",
		Usage: "[Default: false] If set true, the Docker image is copied to the target repository, otherwise it is moved.` `",
	},
	maxDays: cli.StringFlag{
		Name:  maxDays,
		Usage: "[Optional] The maximum number of days to keep builds in Artifactory.` `",
	},
	maxBuilds: cli.StringFlag{
		Name:  maxBuilds,
		Usage: "[Optional] The maximum number of builds to store in Artifactory.` `",
	},
	excludeBuilds: cli.StringFlag{
		Name:  excludeBuilds,
		Usage: "[Optional] List of comma-separated(,) build numbers in the form of \"value1,value2,...\", that should not be removed from Artifactory.` `",
	},
	deleteArtifacts: cli.BoolFlag{
		Name:  deleteArtifacts,
		Usage: "[Default: false] If set to true, automatically removes build artifacts stored in Artifactory.` `",
	},
	bdiAsync: cli.BoolFlag{
		Name:  Async,
		Usage: "[Default: false] If set to true, build discard will run asynchronously and will not wait for response.` `",
	},
	refs: cli.StringFlag{
		Name:  refs,
		Usage: "[Default: refs/remotes/*] List of comma-separated(,) Git references in the form of \"ref1,ref2,...\" which should be preserved.` `",
	},
	glcRepo: cli.StringFlag{
		Name:  repo,
		Usage: "[Optional] Local Git LFS repository which should be cleaned. If omitted, this is detected from the Git repository.` `",
	},
	glcDryRun: cli.BoolFlag{
		Name:  dryRun,
		Usage: "[Default: false] If true, cleanup is only simulated. No files are actually deleted.` `",
	},
	glcQuiet: cli.BoolFlag{
		Name:  quiet,
		Usage: "[Default: $CI] Set to true to skip the delete confirmation message.` `",
	},
	global: cli.BoolFlag{
		Name:  global,
		Usage: "[Default: false] Set to true if you'd like the configuration to be global (for all projects). Specific projects can override the global configuration.` `",
	},
	serverIdResolve: cli.StringFlag{
		Name:  serverIdResolve,
		Usage: "[Optional] Artifactory server ID for resolution. The server should be configured using the 'jfrog c add' command.` `",
	},
	serverIdDeploy: cli.StringFlag{
		Name:  serverIdDeploy,
		Usage: "[Optional] Artifactory server ID for deployment. The server should be configured using the 'jfrog c add' command.` `",
	},
	repoResolveReleases: cli.StringFlag{
		Name:  repoResolveReleases,
		Usage: "[Optional] Resolution repository for release dependencies.` `",
	},
	repoResolveSnapshots: cli.StringFlag{
		Name:  repoResolveSnapshots,
		Usage: "[Optional] Resolution repository for snapshot dependencies.` `",
	},
	repoDeployReleases: cli.StringFlag{
		Name:  repoDeployReleases,
		Usage: "[Optional] Deployment repository for release artifacts.` `",
	},
	repoDeploySnapshots: cli.StringFlag{
		Name:  repoDeploySnapshots,
		Usage: "[Optional] Deployment repository for snapshot artifacts.` `",
	},
	includePatterns: cli.StringFlag{
		Name:  includePatterns,
		Usage: "[Optional] Filter deployed artifacts by setting a wildcard pattern that specifies which artifacts to include. You may provide multiple patterns separated by ', '.` `",
	},
	excludePatterns: cli.StringFlag{
		Name:  excludePatterns,
		Usage: "[Optional] Filter deployed artifacts by setting a wildcard pattern that specifies which artifacts to exclude. You may provide multiple patterns separated by ', '.` `",
	},
	repoResolve: cli.StringFlag{
		Name:  repoResolve,
		Usage: "[Optional] Repository for dependencies resolution.` `",
	},
	repoDeploy: cli.StringFlag{
		Name:  repoDeploy,
		Usage: "[Optional] Repository for artifacts deployment.` `",
	},
	usesPlugin: cli.BoolFlag{
		Name:  usesPlugin,
		Usage: "[Default: false] Set to true if the Gradle Artifactory Plugin is already applied in the build script.` `",
	},
	deployMavenDesc: cli.BoolTFlag{
		Name:  deployMavenDesc,
		Usage: "[Default: true] Set to false if you do not wish to deploy Maven descriptors.` `",
	},
	deployIvyDesc: cli.BoolTFlag{
		Name:  deployIvyDesc,
		Usage: "[Default: true] Set to false if you do not wish to deploy Ivy descriptors.` `",
	},
	ivyDescPattern: cli.StringFlag{
		Name:  ivyDescPattern,
		Usage: "[Default: '[organization]/[module]/ivy-[revision].xml' Set the deployed Ivy descriptor pattern.` `",
	},
	ivyArtifactsPattern: cli.StringFlag{
		Name:  ivyArtifactsPattern,
		Usage: "[Default: '[organization]/[module]/[revision]/[artifact]-[revision](-[classifier]).[ext]' Set the deployed Ivy artifacts pattern.` `",
	},
	deploymentThreads: cli.StringFlag{
		Name:  threads,
		Value: "",
		Usage: "[Default: " + strconv.Itoa(commonCliUtils.Threads) + "] Number of threads for uploading build artifacts.` `",
	},
	skipLogin: cli.BoolFlag{
		Name:  skipLogin,
		Usage: "[Default: false] Set to true if you'd like the command to skip performing docker login.` `",
	},
	npmDetailedSummary: cli.BoolFlag{
		Name:  detailedSummary,
		Usage: "[Default: false] Set to true to include a list of the affected files in the command summary.` `",
	},
	nugetV2: cli.BoolFlag{
		Name:  nugetV2,
		Usage: "[Default: false] Set to true if you'd like to use the NuGet V2 protocol when restoring packages from Artifactory.` `",
	},
	noFallback: cli.BoolTFlag{
		Name:  noFallback,
		Usage: "[Default: false] Set to true to avoid downloading packages from the VCS, if they are missing in Artifactory.` `",
	},
	namespace: cli.StringFlag{
		Name:  namespace,
		Usage: "[Mandatory] Terraform namespace.` `",
	},
	provider: cli.StringFlag{
		Name:  provider,
		Usage: "[Mandatory] Terraform provider.` `",
	},
	tag: cli.StringFlag{
		Name:  tag,
		Usage: "[Mandatory] Terraform package tag.` `",
	},
	vars: cli.StringFlag{
		Name:  vars,
		Usage: "[Optional] List of semicolon-separated(;) variables in the form of \"key1=value1;key2=value2;...\" (wrapped by quotes) to be replaced in the template. In the template, the variables should be used as follows: ${key1}.` `",
	},
	rtAtcGroups: cli.StringFlag{
		Name: Groups,
		Usage: "[Default: *] A list of comma-separated(,) groups for the access token to be associated with. " +
			"Specify * to indicate that this is a 'user-scoped token', i.e., the token provides the same access privileges that the current subject has, and is therefore evaluated dynamically. " +
			"A non-admin user can only provide a scope that is a subset of the groups to which he belongs` `",
	},
	rtAtcGrantAdmin: cli.BoolFlag{
		Name:  GrantAdmin,
		Usage: "[Default: false] Set to true to provide admin privileges to the access token. This is only available for administrators.` `",
	},
	rtAtcExpiry: cli.StringFlag{
		Name:  Expiry,
		Usage: "[Default: " + strconv.Itoa(ArtifactoryTokenExpiry) + "] The time in seconds for which the token will be valid. To specify a token that never expires, set to zero. Non-admin may only set a value that is equal to or less than the default 3600.` `",
	},
	rtAtcRefreshable: cli.BoolFlag{
		Name:  Refreshable,
		Usage: "[Default: false] Set to true if you'd like the token to be refreshable. A refresh token will also be returned in order to be used to generate a new token once it expires.` `",
	},
	rtAtcAudience: cli.StringFlag{
		Name:  Audience,
		Usage: "[Optional] A space-separated list of the other Artifactory instances or services that should accept this token identified by their Artifactory Service IDs, as obtained by the 'jfrog rt curl api/system/service_id' command.` `",
	},
	usersCreateCsv: cli.StringFlag{
		Name:  csv,
		Usage: "[Mandatory] Path to a CSV file with the users' details. The first row of the file is reserved for the cells' headers. It must include \"username\",\"password\",\"email\"` `",
	},
	usersDeleteCsv: cli.StringFlag{
		Name:  csv,
		Usage: "[Optional] Path to a CSV file with the users' details. The first row of the file is reserved for the cells' headers. It must include \"username\"` `",
	},
	UsersGroups: cli.StringFlag{
		Name:  UsersGroups,
		Usage: "[Optional] A list of comma-separated(,) groups for the new users to be associated with.` `",
	},
	Replace: cli.BoolFlag{
		Name:  Replace,
		Usage: "[Default: false] Set to true if you'd like existing users or groups to be replaced.` `",
	},
	Admin: cli.BoolFlag{
		Name:  Admin,
		Usage: "[Default: false] Set to true if you'd like to create an admin user.` `",
	},
	ocStartBuildRepo: cli.StringFlag{
		Name:  repo,
		Usage: "[Mandatory] The name of the repository to which the image was pushed.` `",
	},
	Force: cli.BoolFlag{
		Name:  Force,
		Usage: "[Default: false] Set to true to allow config transfer to a non-empty Artifactory server.` `",
	},
	Verbose: cli.BoolFlag{
		Name:  Verbose,
		Usage: "[Default: false] Set to true to increase verbosity during the export configuration from the source Artifactory phase.` `",
	},
	SourceWorkingDir: cli.StringFlag{
		Name:  SourceWorkingDir,
		Usage: "[Default: $JFROG_CLI_TEMP_DIR] Local working directory on the source Artifactory server.` `",
	},
	TargetWorkingDir: cli.StringFlag{
		Name:  TargetWorkingDir,
		Usage: "[Default: '/storage'] Local working directory on the target Artifactory server.` `",
	},

	// Distribution's commands Flags
	distUrl: cli.StringFlag{
		Name:  url,
		Usage: "[Optional] JFrog Distribution URL. (example: https://acme.jfrog.io/distribution)` `",
	},
	rbDryRun: cli.BoolFlag{
		Name:  dryRun,
		Usage: "[Default: false] Set to true to disable communication with JFrog Distribution.` `",
	},
	rbDetailedSummary: cli.BoolFlag{
		Name:  detailedSummary,
		Usage: "[Default: false] Set to true to get a command summary with details about the release bundle artifact.` `",
	},
	sign: cli.BoolFlag{
		Name:  sign,
		Usage: "[Default: false] If set to true, automatically signs the release bundle version.` `",
	},
	desc: cli.StringFlag{
		Name:  desc,
		Usage: "[Optional] Description of the release bundle.` `",
	},
	releaseNotesPath: cli.StringFlag{
		Name:  releaseNotesPath,
		Usage: "[Optional] Path to a file describes the release notes for the release bundle version.` `",
	},
	releaseNotesSyntax: cli.StringFlag{
		Name:  "release-notes-syntax",
		Usage: "[Default: plain_text] The syntax for the release notes. Can be one of 'markdown', 'asciidoc', or 'plain_text` `",
	},
	rbPassphrase: cli.StringFlag{
		Name:  passphrase,
		Usage: "[Optional] The passphrase for the signing key.` `",
	},
	distTarget: cli.StringFlag{
		Name: target,
		Usage: "[Optional] The target path for distributed artifacts on the edge node. If not specified, the artifacts will have the same path and name on the edge node, as on the source Artifactory server. " +
			"For flexibility in specifying the distribution path, you can include placeholders in the form of {1}, {2} which are replaced by corresponding tokens in the pattern path that are enclosed in parenthesis.` `",
	},
	rbRepo: cli.StringFlag{
		Name:  repo,
		Usage: "[Optional] A repository name at source Artifactory to store release bundle artifacts in. If not provided, Artifactory will use the default one.` `",
	},
	DistRules: cli.StringFlag{
		Name:  DistRules,
		Usage: "[Optional] Path to distribution rules.` `",
	},
	site: cli.StringFlag{
		Name:  site,
		Usage: "[Default: '*'] Wildcard filter for site name.` `",
	},
	city: cli.StringFlag{
		Name:  city,
		Usage: "[Default: '*'] Wildcard filter for site city name.` `",
	},
	countryCodes: cli.StringFlag{
		Name:  countryCodes,
		Usage: "[Default: '*'] List of semicolon-separated(;) wildcard filters for site country codes.` `",
	},
	sync: cli.BoolFlag{
		Name:  sync,
		Usage: "[Default: false] Set to true to enable sync distribution (the command execution will end when the distribution process ends).` `",
	},
	maxWaitMinutes: cli.StringFlag{
		Name:  maxWaitMinutes,
		Usage: "[Default: 60] Max minutes to wait for sync distribution.` `",
	},
	deleteFromDist: cli.BoolFlag{
		Name:  deleteFromDist,
		Usage: "[Default: false] Set to true to delete release bundle version in JFrog Distribution itself after deletion is complete in the specified Edge node/s.` `",
	},
	targetProps: cli.StringFlag{
		Name:  targetProps,
		Usage: "[Optional] List of semicolon-separated(;) properties, in the form of \"key1=value1;key2=value2;...\" to be added to the artifacts after distribution of the release bundle.` `",
	},

	// Xray's commands Flags
	xrUrl: cli.StringFlag{
		Name:  url,
		Usage: "[Optional] JFrog Xray URL. (example: https://acme.jfrog.io/xray)` `",
	},
	xrayScan: cli.StringFlag{
		Name:  xrayScan,
		Usage: "[Default: false] Set if you'd like all files to be scanned by Xray on the local file system prior to the upload, and skip the upload if any of the files are found vulnerable.` `",
	},
	licenseId: cli.StringFlag{
		Name:  licenseId,
		Usage: "[Mandatory] Xray license ID.` `",
	},
	from: cli.StringFlag{
		Name:  from,
		Usage: "[Optional] From update date in YYYY-MM-DD format.` `",
	},
	to: cli.StringFlag{
		Name:  to,
		Usage: "[Optional] To update date in YYYY-MM-DD format.` `",
	},
	Version: cli.StringFlag{
		Name:  Version,
		Usage: "[Optional] Xray API version.` `",
	},
	target: cli.StringFlag{
		Name:  target,
		Usage: "[Default: ./] Path for downloaded update files.` `",
	},
	Periodic: cli.BoolFlag{
		Name:  Periodic,
		Usage: fmt.Sprintf("[Default: false] Set to true to get the Xray DBSync V3 Periodic Package (Use with %s flag).` `", Stream),
	},
	useWrapperAudit: cli.BoolTFlag{
		Name:  UseWrapper,
		Usage: "[Default: true] Set to false if you wish to not use the gradle or maven wrapper.` `",
	},
	ExcludeTestDeps: cli.BoolFlag{
		Name:  ExcludeTestDeps,
		Usage: "[Default: false] [Gradle] Set to true if you'd like to exclude Gradle test dependencies from Xray scanning.` `",
	},
	DepType: cli.StringFlag{
		Name:  DepType,
		Usage: "[Default: all] [npm] Defines npm dependencies type. Possible values are: all, devOnly and prodOnly` `",
	},
	RequirementsFile: cli.StringFlag{
		Name:  RequirementsFile,
		Usage: "[Optional] [Pip] Defines pip requirements file name. For example: 'requirements.txt'.` `",
	},
	FixableOnly: cli.BoolFlag{
		Name:  FixableOnly,
		Usage: "[Optional] Set to true if you wish to display issues that have a fixed version only.` `",
	},
	MinSeverity: cli.StringFlag{
		Name:  MinSeverity,
		Usage: "[Optional] Set the minimum severity of issues to display. The following values are accepted: Low, Medium, High or Critical.` `",
	},
	watches: cli.StringFlag{
		Name:  watches,
		Usage: "[Optional] A comma-separated(,) list of Xray watches, to determine Xray's violations creation.` `",
	},
	workingDirs: cli.StringFlag{
		Name:  workingDirs,
		Usage: "[Optional] A comma-separated(,) list of relative working directories, to determine audit targets locations.` `",
	},
	ExclusionsAudit: cli.StringFlag{
		Name:  exclusions,
		Usage: "[Default: *node_modules*;*target*;*venv*;*test*] List of semicolon-separated(;) exclusions, utilized to skip sub-projects from undergoing an audit. These exclusions may incorporate the * and ? wildcards.` `",
	},
	ExtendedTable: cli.BoolFlag{
		Name:  ExtendedTable,
		Usage: "[Default: false] Set to true if you'd like the table to include extended fields such as 'CVSS' & 'Xray Issue Id'. Ignored if provided 'format' is not 'table'.` `",
	},
	UseWrapper: cli.BoolFlag{
		Name:  UseWrapper,
		Usage: "[Default: false] Set to true if you wish to use the wrapper.` `",
	},
	licenses: cli.BoolFlag{
		Name:  licenses,
		Usage: "[Default: false] Set to true if you'd like to receive licenses from Xray scanning.` `",
	},
	vuln: cli.BoolFlag{
		Name:  vuln,
		Usage: "[Default: false] Set to true if you'd like to receive an additional view of all vulnerabilities, regardless of the policy configured in Xray. Ignored if provided 'format' is 'sarif'.` `",
	},
	repoPath: cli.StringFlag{
		Name:  repoPath,
		Usage: "[Optional] Target repo path, to enable Xray to determine watches accordingly.` `",
	},
	scanRecursive: cli.BoolTFlag{
		Name:  recursive,
		Usage: "[Default: true] Set to false if you do not wish to collect artifacts in sub-folders to be scanned by Xray.` `",
	},
	scanRegexp: cli.BoolFlag{
		Name:  regexpFlag,
		Usage: "[Default: false] Set to true to use a regular expression instead of wildcards expression to collect files to scan.` `",
	},
	scanAnt: cli.BoolFlag{
		Name:  antFlag,
		Usage: "[Default: false] Set to true to use an ant pattern instead of wildcards expression to collect files to scan.` `",
	},
	xrOutput: cli.StringFlag{
		Name:  xrOutput,
		Usage: "[Default: table] Defines the output format of the command. Acceptable values are: table, json, simple-json and sarif. Note: the json format doesn't include information about scans that are included as part of the Advanced Security package.` `",
	},
	BypassArchiveLimits: cli.BoolFlag{
		Name:  BypassArchiveLimits,
		Usage: "[Default: false] Set to true to bypass the indexer-app archive limits.` `",
	},
	Mvn: cli.BoolFlag{
		Name:  Mvn,
		Usage: "[Default: false] Set to true to request audit for a Maven project.` `",
	},
	Gradle: cli.BoolFlag{
		Name:  Gradle,
		Usage: "[Default: false] Set to true to request audit for a Gradle project.` `",
	},
	Npm: cli.BoolFlag{
		Name:  Npm,
		Usage: "[Default: false] Set to true to request audit for an npm project.` `",
	},
	Yarn: cli.BoolFlag{
		Name:  Yarn,
		Usage: "[Default: false] Set to true to request audit for a Yarn project.` `",
	},
	Nuget: cli.BoolFlag{
		Name:  Nuget,
		Usage: "[Default: false] Set to true to request audit for a .NET project.` `",
	},
	Pip: cli.BoolFlag{
		Name:  Pip,
		Usage: "[Default: false] Set to true to request audit for a Pip project.` `",
	},
	Pipenv: cli.BoolFlag{
		Name:  Pipenv,
		Usage: "[Default: false] Set to true to request audit for a Pipenv project.` `",
	},
	Poetry: cli.BoolFlag{
		Name:  Poetry,
		Usage: "[Default: false] Set to true to request audit for a Poetry project.` `",
	},
	Go: cli.BoolFlag{
		Name:  Go,
		Usage: "[Default: false] Set to true to request audit for a Go project.` `",
	},
	goPublishExclusions: cli.StringFlag{
		Name:  exclusions,
		Usage: "[Optional] List of semicolon-separated(;) exclusions. Exclusions can include the * and the ? wildcards.` `",
	},
	rescan: cli.BoolFlag{
		Name:  rescan,
		Usage: "[Default: false] Set to true when scanning an already successfully scanned build, for example after adding an ignore rule.` `",
	},
	curationThreads: cli.StringFlag{
		Name:  threads,
		Value: "",
		Usage: "[Default: 10] Number of working threads.` `",
	},
	curationOutput: cli.StringFlag{
		Name:  xrOutput,
		Usage: "[Default: table] Defines the output format of the command. Acceptable values are: table, json.` `",
	},

	// Mission Control's commands Flags
	mcUrl: cli.StringFlag{
		Name:  url,
		Usage: "[Optional] JFrog Mission Control URL. (example: https://acme.jfrog.io/mc)` `",
	},
	mcAccessToken: cli.StringFlag{
		Name:  accessToken,
		Usage: "[Optional] Mission Control Admin token.` `",
	},
	mcInteractive: cli.BoolTFlag{
		Name:  interactive,
		Usage: "[Default: true] Set to false if you do not want the config command to be interactive. If true, the other command options become optional.",
	},
	licenseCount: cli.StringFlag{
		Name:  licenseCount,
		Value: "",
		Usage: "[Default: " + strconv.Itoa(DefaultLicenseCount) + "] The number of licenses to deploy. The minimum value is 1.` `",
	},
	imageFile: cli.StringFlag{
		Name:  imageFile,
		Usage: "[Mandatory] Path to a file which includes one line in the following format: <IMAGE-TAG>@sha256:<MANIFEST-SHA256>.` `",
	},

	// Config commands Flags
	configPlatformUrl: cli.StringFlag{
		Name:  url,
		Usage: "[Optional] JFrog platform URL. (example: https://acme.jfrog.io)` `",
	},
	configRtUrl: cli.StringFlag{
		Name:  configRtUrl,
		Usage: "[Optional] JFrog Artifactory URL. (example: https://acme.jfrog.io/artifactory)` `",
	},
	configDistUrl: cli.StringFlag{
		Name:  configDistUrl,
		Usage: "[Optional] JFrog Distribution URL. (example: https://acme.jfrog.io/distribution)` `",
	},
	configXrUrl: cli.StringFlag{
		Name:  configXrUrl,
		Usage: "[Optional] JFrog Xray URL. (example: https://acme.jfrog.io/xray)` `",
	},
	configMcUrl: cli.StringFlag{
		Name:  configMcUrl,
		Usage: "[Optional] JFrog Mission Control URL. (example: https://acme.jfrog.io/mc)` `",
	},
	configPlUrl: cli.StringFlag{
		Name:  configPlUrl,
		Usage: "[Optional] JFrog Pipelines URL. (example: https://acme.jfrog.io/pipelines)` `",
	},
	configUser: cli.StringFlag{
		Name:  user,
		Usage: "[Optional] JFrog Platform username.` `",
	},
	configPassword: cli.StringFlag{
		Name:  password,
		Usage: "[Optional] JFrog Platform password or API key.` `",
	},
	configAccessToken: cli.StringFlag{
		Name:  accessToken,
		Usage: "[Optional] JFrog Platform access token.` `",
	},
	configInsecureTls: cli.StringFlag{
		Name:  InsecureTls,
		Usage: "[Default: false] Set to true to skip TLS certificates verification, while encrypting the Artifactory password during the config process.` `",
	},
	projectPath: cli.StringFlag{
		Name:  projectPath,
		Usage: "[Default: ./] Full path to the code project.` `",
	},
	Install: cli.BoolFlag{
		Name:  Install,
		Usage: "[Default: false] Set to true to install the completion script instead of printing it to the standard output.` `",
	},
	CreateRepo: cli.BoolFlag{
		Name:  CreateRepo,
		Usage: "[Default: false] Set to true to create the repository on the edge if it does not exist.` `",
	},
	Filestore: cli.BoolFlag{
		Name:  Filestore,
		Usage: "[Default: false] Set to true to make the transfer mechanism check for the existence of artifacts on the target filestore. Used when the files are already expected to be located on the filestore.` `",
	},
	IncludeRepos: cli.StringFlag{
		Name:  IncludeRepos,
		Usage: "[Optional] List of semicolon-separated(;) repositories to include in the transfer. You can use wildcards to specify patterns for the repositories' names.` `",
	},
	ExcludeRepos: cli.StringFlag{
		Name:  ExcludeRepos,
		Usage: "[Optional] List of semicolon-separated(;) repositories to exclude from the transfer. You can use wildcards to specify patterns for the repositories' names.` `",
	},
	IncludeProjects: cli.StringFlag{
		Name:  IncludeProjects,
		Usage: "[Optional] List of semicolon-separated(;) JFrog Project keys to include in the transfer. You can use wildcards to specify patterns for the JFrog Project keys.` `",
	},
	ExcludeProjects: cli.StringFlag{
		Name:  ExcludeProjects,
		Usage: "[Optional] List of semicolon-separated(;) JFrog Projects to exclude from the transfer. You can use wildcards to specify patterns for the project keys.` `",
	},
	IgnoreState: cli.BoolFlag{
		Name:  IgnoreState,
		Usage: "[Default: false] Set to true to ignore the saved state from previous transfer-files operations.` `",
	},
	ProxyKey: cli.StringFlag{
		Name:  ProxyKey,
		Usage: "[Optional] The key of an HTTP proxy configuration in Artifactory. This proxy will be used for the transfer traffic between the source and target instances. To configure this proxy, go to \"Proxies | Configuration | Proxy Configuration\" in the JFrog Administration UI.` `",
	},
	transferFilesStatus: cli.BoolFlag{
		Name:  Status,
		Usage: "[Default: false] Set to true to show the status of the transfer-files command currently in progress.` `",
	},
	branch: cli.StringFlag{
		Name:  branch,
		Usage: "[Mandatory] Branch name to filter.` `",
	},
	pipelineName: cli.StringFlag{
		Name:  pipelineName,
		Usage: "[Optional] Pipeline name to filter.` `",
	},
	monitor: cli.BoolFlag{
		Name:  monitor,
		Usage: "[Default: false] Monitor pipeline status.` `",
	},
	repository: cli.StringFlag{
		Name:  repository,
		Usage: "[Mandatory] Repository name to filter resource.` `",
	},
	singleBranch: cli.BoolFlag{
		Name:  singleBranch,
		Usage: "[Default: false] Single branch to filter multi branches and single branch pipelines sources.` `",
	},
	Stop: cli.BoolFlag{
		Name:  Stop,
		Usage: "[Default: false] Set to true to stop the transfer-files command currently in progress. Useful when running the transfer-files command in the background.` `",
	},
	installPluginVersion: cli.StringFlag{
		Name:  Version,
		Usage: "[Default: latest] The plugin version to download and install.` `",
	},
	InstallPluginSrcDir: cli.StringFlag{
		Name:  InstallPluginSrcDir,
		Usage: "[Optional] The local directory that contains the plugin files to install.` `",
	},
	InstallPluginHomeDir: cli.StringFlag{
		Name:  InstallPluginHomeDir,
		Usage: "[Default: /opt/jfrog] The local JFrog home directory to install the plugin in.` `",
	},
	PreChecks: cli.BoolFlag{
		Name:  PreChecks,
		Usage: "[Default: false] Set to true to run pre-transfer checks.` `",
	},
	lcSync: cli.BoolTFlag{
		Name:  Sync,
		Usage: "[Default: true] Set to false to run asynchronously.` `",
	},
	lcProject: cli.StringFlag{
		Name:  Project,
		Usage: "[Optional] Project key associated with the Release Bundle version.` `",
	},
	lcBuilds: cli.StringFlag{
		Name:   Builds,
		Usage:  "[Optional] Path to a JSON file containing information of the source builds from which to create a release bundle.` `",
		Hidden: true,
	},
	lcReleaseBundles: cli.StringFlag{
		Name:   ReleaseBundles,
		Usage:  "[Optional] Path to a JSON file containing information of the source release bundles from which to create a release bundle.` `",
		Hidden: true,
	},
	lcSigningKey: cli.StringFlag{
		Name:  SigningKey,
		Usage: "[Mandatory] The GPG/RSA key-pair name given in Artifactory.` `",
	},
	lcPathMappingPattern: cli.StringFlag{
		Name:  PathMappingPattern,
		Usage: "[Optional] Specify along with '" + PathMappingTarget + "' to distribute artifacts to a different path on the edge node. You can use wildcards to specify multiple artifacts.` `",
	},
	lcPathMappingTarget: cli.StringFlag{
		Name: PathMappingTarget,
		Usage: "[Optional] The target path for distributed artifacts on the edge node. If not specified, the artifacts will have the same path and name on the edge node, as on the source Artifactory server. " +
			"For flexibility in specifying the distribution path, you can include placeholders in the form of {1}, {2} which are replaced by corresponding tokens in the pattern path that are enclosed in parenthesis.` `",
	},
	lcDryRun: cli.BoolFlag{
		Name:  dryRun,
		Usage: "[Default: false] Set to true to only simulate the distribution of the release bundle.` `",
	},
	ThirdPartyContextualAnalysis: cli.BoolFlag{
		Name:   ThirdPartyContextualAnalysis,
		Usage:  "Default: false] [npm] when set, the Contextual Analysis scan also uses the code of the project dependencies to determine the applicability of the vulnerability.",
		Hidden: true,
	},
	lcIncludeRepos: cli.StringFlag{
		Name: IncludeRepos,
		Usage: "[Optional] List of semicolon-separated(;) repositories to include in the promotion. If this property is left undefined, all repositories (except those specifically excluded) are included in the promotion. " +
			"If one or more repositories are specifically included, all other repositories are excluded.` `",
	},
	lcExcludeRepos: cli.StringFlag{
		Name:  ExcludeRepos,
		Usage: "[Optional] List of semicolon-separated(;) repositories to exclude from the promotion.` `",
	},
	atcProject: cli.StringFlag{
		Name:  Project,
		Usage: "[Optional] The project for which this token is created. Enter the project name on which you want to apply this token.` `",
	},
	atcGrantAdmin: cli.BoolFlag{
		Name:  GrantAdmin,
		Usage: "[Default: false] Set to true to provide admin privileges to the access token. This is only available for administrators.` `",
	},
	atcGroups: cli.StringFlag{
		Name: Groups,
		Usage: "[Optional] A list of comma-separated(,) groups for the access token to be associated with. " +
			"This is only available for administrators.` `",
	},
	atcScope: cli.StringFlag{
		Name:  Scope,
		Usage: "[Optional] The scope of access that the token provides. This is only available for administrators.` `",
	},
	atcExpiry: cli.StringFlag{
		Name: Expiry,
		Usage: "[Optional] The amount of time, in seconds, it would take for the token to expire. Must be non-negative." +
			"If not provided, the platform default will be used. To specify a token that never expires, set to zero. " +
			"Non-admin may only set a value that is equal or lower than the platform default that was set by an administrator (1 year by default).` `",
	},
	atcRefreshable: cli.BoolFlag{
		Name:  Refreshable,
		Usage: "[Default: false] Set to true if you'd like the token to be refreshable. A refresh token will also be returned in order to be used to generate a new token once it expires.` `",
	},
	atcDescription: cli.StringFlag{
		Name:  Description,
		Usage: "[Optional] Free text token description. Useful for filtering and managing tokens. Limited to 1024 characters.` `",
	},
	atcAudience: cli.StringFlag{
		Name:  Audience,
		Usage: "[Optional] A space-separated list of the other instances or services that should accept this token identified by their Service-IDs.` `",
	},
	atcReference: cli.BoolFlag{
		Name:  Reference,
		Usage: "[Default: false] Generate a Reference Token (alias to Access Token) in addition to the full token (available from Artifactory 7.38.10)` `",
	},
}

var commandFlags = map[string][]string{
	AddConfig: {
		interactive, EncPassword, configPlatformUrl, configRtUrl, configDistUrl, configXrUrl, configMcUrl, configPlUrl, configUser, configPassword, configAccessToken, sshKeyPath, sshPassphrase, ClientCertPath,
		ClientCertKeyPath, BasicAuthOnly, configInsecureTls, Overwrite, passwordStdin, accessTokenStdin,
	},
	EditConfig: {
		interactive, EncPassword, configPlatformUrl, configRtUrl, configDistUrl, configXrUrl, configMcUrl, configPlUrl, configUser, configPassword, configAccessToken, sshKeyPath, sshPassphrase, ClientCertPath,
		ClientCertKeyPath, BasicAuthOnly, configInsecureTls, passwordStdin, accessTokenStdin,
	},
	DeleteConfig: {
		deleteQuiet,
	},
	Upload: {
		url, user, password, accessToken, sshPassphrase, sshKeyPath, serverId, ClientCertPath, uploadTargetProps,
		ClientCertKeyPath, specFlag, specVars, buildName, buildNumber, module, uploadExclusions, deb,
		uploadRecursive, uploadFlat, uploadRegexp, retries, retryWaitTime, dryRun, uploadExplode, symlinks, includeDirs,
		failNoOp, threads, uploadSyncDeletes, syncDeletesQuiet, InsecureTls, detailedSummary, Project,
		uploadAnt, uploadArchive, uploadMinSplit, uploadSplitCount, ChunkSize,
	},
	Download: {
		url, user, password, accessToken, sshPassphrase, sshKeyPath, serverId, ClientCertPath,
		ClientCertKeyPath, specFlag, specVars, buildName, buildNumber, module, exclusions, sortBy,
		sortOrder, limit, offset, downloadRecursive, downloadFlat, build, includeDeps, excludeArtifacts, downloadMinSplit, downloadSplitCount,
		retries, retryWaitTime, dryRun, downloadExplode, bypassArchiveInspection, validateSymlinks, bundle, publicGpgKey, includeDirs,
		downloadProps, downloadExcludeProps, failNoOp, threads, archiveEntries, downloadSyncDeletes, syncDeletesQuiet, InsecureTls, detailedSummary, Project,
		skipChecksum,
	},
	Move: {
		url, user, password, accessToken, sshPassphrase, sshKeyPath, serverId, ClientCertPath,
		ClientCertKeyPath, specFlag, specVars, exclusions, sortBy, sortOrder, limit, offset, moveRecursive,
		moveFlat, dryRun, build, includeDeps, excludeArtifacts, moveProps, moveExcludeProps, failNoOp, threads, archiveEntries,
		InsecureTls, retries, retryWaitTime, Project,
	},
	Copy: {
		url, user, password, accessToken, sshPassphrase, sshKeyPath, serverId, ClientCertPath,
		ClientCertKeyPath, specFlag, specVars, exclusions, sortBy, sortOrder, limit, offset, copyRecursive,
		copyFlat, dryRun, build, includeDeps, excludeArtifacts, bundle, copyProps, copyExcludeProps, failNoOp, threads,
		archiveEntries, InsecureTls, retries, retryWaitTime, Project,
	},
	Delete: {
		url, user, password, accessToken, sshPassphrase, sshKeyPath, serverId, ClientCertPath,
		ClientCertKeyPath, specFlag, specVars, exclusions, sortBy, sortOrder, limit, offset,
		deleteRecursive, dryRun, build, includeDeps, excludeArtifacts, deleteQuiet, deleteProps, deleteExcludeProps, failNoOp, threads, archiveEntries,
		InsecureTls, retries, retryWaitTime, Project,
	},
	Search: {
		url, user, password, accessToken, sshPassphrase, sshKeyPath, serverId, ClientCertPath,
		ClientCertKeyPath, specFlag, specVars, exclusions, sortBy, sortOrder, limit, offset,
		searchRecursive, build, includeDeps, excludeArtifacts, count, bundle, includeDirs, searchProps, searchExcludeProps, failNoOp, archiveEntries,
		InsecureTls, searchTransitive, retries, retryWaitTime, Project, searchInclude,
	},
	Properties: {
		url, user, password, accessToken, sshPassphrase, sshKeyPath, serverId, ClientCertPath,
		ClientCertKeyPath, specFlag, specVars, exclusions, sortBy, sortOrder, limit, offset,
		propsRecursive, build, includeDeps, excludeArtifacts, bundle, includeDirs, failNoOp, threads, archiveEntries, propsProps, propsExcludeProps,
		InsecureTls, retries, retryWaitTime, Project,
	},
	BuildPublish: {
		url, user, password, accessToken, sshPassphrase, sshKeyPath, serverId, buildUrl, bpDryRun,
		envInclude, envExclude, InsecureTls, Project, bpDetailedSummary,
	},
	BuildAppend: {
		url, user, password, accessToken, sshPassphrase, sshKeyPath, serverId, buildUrl, bpDryRun,
		envInclude, envExclude, InsecureTls, Project,
	},
	BuildAddDependencies: {
		specFlag, specVars, uploadExclusions, badRecursive, badRegexp, badDryRun, Project, badFromRt, serverId, badModule,
	},
	BuildAddGit: {
		configFlag, serverId, Project,
	},
	BuildCollectEnv: {
		Project,
	},
	BuildDockerCreate: {
		buildName, buildNumber, module, url, user, password, accessToken, sshPassphrase, sshKeyPath,
		serverId, imageFile, Project,
	},
	OcStartBuild: {
		buildName, buildNumber, module, Project, serverId, ocStartBuildRepo,
	},
	BuildScanLegacy: {
		url, user, password, accessToken, sshPassphrase, sshKeyPath, serverId, fail, InsecureTls,
		Project,
	},
	BuildPromote: {
		url, user, password, accessToken, sshPassphrase, sshKeyPath, serverId, Status, comment,
		sourceRepo, includeDependencies, copyFlag, failFast, bprDryRun, bprProps, InsecureTls, Project,
	},
	BuildDiscard: {
		url, user, password, accessToken, sshPassphrase, sshKeyPath, serverId, maxDays, maxBuilds,
		excludeBuilds, deleteArtifacts, bdiAsync, InsecureTls, Project,
	},
	GitLfsClean: {
		url, user, password, accessToken, sshPassphrase, sshKeyPath, serverId, refs, glcRepo, glcDryRun,
		glcQuiet, InsecureTls, retries, retryWaitTime,
	},
	CocoapodsConfig: {
		global, serverIdResolve, repoResolve,
	},
	SwiftConfig: {
		global, serverIdResolve, repoResolve,
	},
	MvnConfig: {
		global, serverIdResolve, serverIdDeploy, repoResolveReleases, repoResolveSnapshots, repoDeployReleases, repoDeploySnapshots, includePatterns, excludePatterns, UseWrapper,
	},
	GradleConfig: {
		global, serverIdResolve, serverIdDeploy, repoResolve, repoDeploy, usesPlugin, UseWrapper, deployMavenDesc,
		deployIvyDesc, ivyDescPattern, ivyArtifactsPattern,
	},
	Mvn: {
		buildName, buildNumber, deploymentThreads, InsecureTls, Project, detailedSummary, xrayScan, xrOutput,
	},
	Gradle: {
		buildName, buildNumber, deploymentThreads, Project, detailedSummary, xrayScan, xrOutput,
	},
	Docker: {
		buildName, buildNumber, module, Project,
		serverId, skipLogin, threads, detailedSummary, watches, repoPath, licenses, xrOutput, fail, ExtendedTable, BypassArchiveLimits, MinSeverity, FixableOnly, vuln,
	},
	DockerPush: {
		buildName, buildNumber, module, Project,
		serverId, skipLogin, threads, detailedSummary,
	},
	DockerPull: {
		buildName, buildNumber, module, Project,
		serverId, skipLogin,
	},
	DockerPromote: {
		targetDockerImage, sourceTag, targetTag, dockerPromoteCopy, url, user, password, accessToken, sshPassphrase, sshKeyPath,
		serverId,
	},
	ContainerPush: {
		buildName, buildNumber, module, url, user, password, accessToken, sshPassphrase, sshKeyPath,
		serverId, skipLogin, threads, Project, detailedSummary,
	},
	ContainerPull: {
		buildName, buildNumber, module, url, user, password, accessToken, sshPassphrase, sshKeyPath,
		serverId, skipLogin, Project,
	},
	NpmConfig: {
		global, serverIdResolve, serverIdDeploy, repoResolve, repoDeploy,
	},
	NpmInstallCi: {
		buildName, buildNumber, module, Project,
	},
	NpmPublish: {
		buildName, buildNumber, module, Project, npmDetailedSummary, xrayScan, xrOutput,
	},
	PnpmConfig: {
		global, serverIdResolve, repoResolve,
	},
	YarnConfig: {
		global, serverIdResolve, repoResolve,
	},
	Yarn: {
		buildName, buildNumber, module, Project,
	},
	NugetConfig: {
		global, serverIdResolve, repoResolve, nugetV2,
	},
	Nuget: {
		buildName, buildNumber, module, Project,
	},
	DotnetConfig: {
		global, serverIdResolve, repoResolve, nugetV2,
	},
	Dotnet: {
		buildName, buildNumber, module, Project,
	},
	GoConfig: {
		global, serverIdResolve, serverIdDeploy, repoResolve, repoDeploy,
	},
	GoPublish: {
		url, user, password, accessToken, buildName, buildNumber, module, Project, detailedSummary, goPublishExclusions,
	},
	Go: {
		buildName, buildNumber, module, Project, noFallback,
	},
	TerraformConfig: {
		global, serverIdDeploy, repoDeploy,
	},
	Terraform: {
		namespace, provider, tag, exclusions,
		buildName, buildNumber, module, Project,
	},
	Twine: {
		buildName, buildNumber, module, Project,
	},
	TransferConfig: {
		Force, Verbose, IncludeRepos, ExcludeRepos, SourceWorkingDir, TargetWorkingDir, PreChecks,
	},
	TransferConfigMerge: {
		IncludeRepos, ExcludeRepos, IncludeProjects, ExcludeProjects,
	},
	Ping: {
		url, user, password, accessToken, sshPassphrase, sshKeyPath, serverId, ClientCertPath,
		ClientCertKeyPath, InsecureTls,
	},
	RtCurl: {
		serverId,
	},
	PipConfig: {
		global, serverIdResolve, serverIdDeploy, repoResolve, repoDeploy,
	},
	PipInstall: {
		buildName, buildNumber, module, Project,
	},
	PipenvConfig: {
		global, serverIdResolve, serverIdDeploy, repoResolve, repoDeploy,
	},
	PipenvInstall: {
		buildName, buildNumber, module, Project,
	},
	PoetryConfig: {
		global, serverIdResolve, repoResolve,
	},
	Poetry: {
		buildName, buildNumber, module, Project,
	},
	ReleaseBundleV1Create: {
		distUrl, user, password, accessToken, serverId, specFlag, specVars, targetProps,
		rbDryRun, sign, desc, exclusions, releaseNotesPath, releaseNotesSyntax, rbPassphrase, rbRepo, InsecureTls, distTarget, rbDetailedSummary,
	},
	ReleaseBundleV1Update: {
		distUrl, user, password, accessToken, serverId, specFlag, specVars, targetProps,
		rbDryRun, sign, desc, exclusions, releaseNotesPath, releaseNotesSyntax, rbPassphrase, rbRepo, InsecureTls, distTarget, rbDetailedSummary,
	},
	ReleaseBundleV1Sign: {
		distUrl, user, password, accessToken, serverId, rbPassphrase, rbRepo,
		InsecureTls, rbDetailedSummary,
	},
	ReleaseBundleV1Distribute: {
		distUrl, user, password, accessToken, serverId, rbDryRun, DistRules,
		site, city, countryCodes, sync, maxWaitMinutes, InsecureTls, CreateRepo,
	},
	ReleaseBundleV1Delete: {
		distUrl, user, password, accessToken, serverId, rbDryRun, DistRules,
		site, city, countryCodes, sync, maxWaitMinutes, InsecureTls, deleteFromDist, deleteQuiet,
	},
	TemplateConsumer: {
		url, user, password, accessToken, sshPassphrase, sshKeyPath, serverId, ClientCertPath,
		ClientCertKeyPath, vars,
	},
	RepoDelete: {
		url, user, password, accessToken, sshPassphrase, sshKeyPath, serverId, ClientCertPath,
		ClientCertKeyPath, deleteQuiet,
	},
	ReplicationDelete: {
		url, user, password, accessToken, sshPassphrase, sshKeyPath, serverId, ClientCertPath,
		ClientCertKeyPath, deleteQuiet,
	},
	PermissionTargetDelete: {
		url, user, password, accessToken, sshPassphrase, sshKeyPath, serverId, ClientCertPath,
		ClientCertKeyPath, deleteQuiet,
	},
	ArtifactoryAccessTokenCreate: {
		url, user, password, accessToken, sshPassphrase, sshKeyPath, serverId, ClientCertPath,
		ClientCertKeyPath, rtAtcGroups, rtAtcGrantAdmin, rtAtcExpiry, rtAtcRefreshable, rtAtcAudience,
	},
	AccessTokenCreate: {
		platformUrl, user, password, accessToken, sshPassphrase, sshKeyPath, serverId, ClientCertPath, ClientCertKeyPath,
		atcProject, atcGrantAdmin, atcGroups, atcScope, atcExpiry,
		atcRefreshable, atcDescription, atcAudience, atcReference,
	},
	UserCreate: {
		url, user, password, accessToken, sshPassphrase, sshKeyPath, serverId,
		UsersGroups, Replace, Admin,
	},
	UsersCreate: {
		url, user, password, accessToken, sshPassphrase, sshKeyPath, serverId,
		usersCreateCsv, UsersGroups, Replace,
	},
	UsersDelete: {
		url, user, password, accessToken, sshPassphrase, sshKeyPath, serverId,
		usersDeleteCsv, deleteQuiet,
	},
	GroupCreate: {
		url, user, password, accessToken, sshPassphrase, sshKeyPath, serverId,
		Replace,
	},
	GroupAddUsers: {
		url, user, password, accessToken, sshPassphrase, sshKeyPath, serverId,
	},
	GroupDelete: {
		url, user, password, accessToken, sshPassphrase, sshKeyPath, serverId, deleteQuiet,
	},
	TransferFiles: {
		Filestore, IncludeRepos, ExcludeRepos, IgnoreState, ProxyKey, transferFilesStatus, Stop, PreChecks,
	},
	TransferInstall: {
		installPluginVersion, InstallPluginSrcDir, InstallPluginHomeDir,
	},
	ReleaseBundleCreate: {
		platformUrl, user, password, accessToken, serverId, lcSigningKey, lcSync, lcProject, lcBuilds, lcReleaseBundles,
		specFlag, specVars,
	},
	ReleaseBundlePromote: {
		platformUrl, user, password, accessToken, serverId, lcSigningKey, lcSync, lcProject, lcIncludeRepos, lcExcludeRepos,
	},
	ReleaseBundleDistribute: {
		platformUrl, user, password, accessToken, serverId, lcProject, DistRules, site, city, countryCodes,
		lcDryRun, CreateRepo, lcPathMappingPattern, lcPathMappingTarget, lcSync, maxWaitMinutes,
	},
	ReleaseBundleDeleteLocal: {
		platformUrl, user, password, accessToken, serverId, deleteQuiet, lcSync, lcProject,
	},
	ReleaseBundleDeleteRemote: {
		platformUrl, user, password, accessToken, serverId, deleteQuiet, lcDryRun, DistRules, site, city, countryCodes,
		lcSync, maxWaitMinutes, lcProject,
	},
	ReleaseBundleExport: {
		platformUrl, user, password, accessToken, serverId, lcPathMappingTarget, lcPathMappingPattern, Project,
		downloadMinSplit, downloadSplitCount,
	},
	ReleaseBundleImport: {
		user, password, accessToken, serverId, platformUrl,
	},
	// Mission Control's commands
	McConfig: {
		mcUrl, mcAccessToken, mcInteractive,
	},
	LicenseAcquire: {
		mcUrl, mcAccessToken,
	},
	LicenseDeploy: {
		mcUrl, mcAccessToken, licenseCount,
	},
	LicenseRelease: {
		mcUrl, mcAccessToken,
	},
	JpdAdd: {
		mcUrl, mcAccessToken,
	},
	JpdDelete: {
		mcUrl, mcAccessToken,
	},
	// Project commands
	InitProject: {
		projectPath, serverId,
	},
	// Completion commands
	Completion: {
		Install,
	},
	// CLI base commands
	Intro: {},
	// Pipelines commands
	Status: {
		branch, serverId, pipelineName, monitor, singleBranch,
	},
	Trigger: {
		serverId, singleBranch,
	},
	Validate: {
		Resources, serverId,
	},
	Version: {
		serverId,
	},
	Sync: {
		serverId,
	},
	SyncStatus: {
		branch, repository, serverId,
	},
}

func GetCommandFlags(cmd string) []cli.Flag {
	flagList, ok := commandFlags[cmd]
	if !ok {
		log.Error("The command \"", cmd, "\" is not found in commands flags map.")
		return nil
	}
	return buildAndSortFlags(flagList)
}

func buildAndSortFlags(keys []string) (flags []cli.Flag) {
	for _, flag := range keys {
		flags = append(flags, flagsMap[flag])
	}
	sort.Slice(flags, func(i, j int) bool { return flags[i].GetName() < flags[j].GetName() })
	return
}
