package cliutils

import (
	"sort"
	"strconv"

	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

const (
	// Artifactory's Commands Keys
	RtConfig                = "rt-config"
	DeleteConfig            = "delete-config"
	Upload                  = "upload"
	Download                = "download"
	Move                    = "move"
	Copy                    = "copy"
	Delete                  = "delete"
	Properties              = "properties"
	Search                  = "search"
	BuildPublish            = "build-publish"
	BuildAppend             = "build-append"
	BuildScan               = "build-scan"
	BuildPromote            = "build-promote"
	BuildDistribute         = "build-distribute"
	BuildDiscard            = "build-discard"
	BuildAddDependencies    = "build-add-dependencies"
	BuildAddGit             = "build-add-git"
	BuildCollectEnv         = "build-collect-env"
	GitLfsClean             = "git-lfs-clean"
	Mvn                     = "mvn"
	MvnConfig               = "mvn-config"
	Gradle                  = "gradle"
	GradleConfig            = "gradle-config"
	DockerPromote           = "docker-promote"
	ContainerPull           = "container-pull"
	ContainerPush           = "container-push"
	BuildDockerCreate       = "build-docker-create"
	NpmConfig               = "npm-config"
	Npm                     = "npm"
	NpmPublish              = "npmPublish"
	NugetConfig             = "nuget-config"
	Nuget                   = "nuget"
	Dotnet                  = "dotnet"
	DotnetConfig            = "dotnet-config"
	Go                      = "go"
	GoConfig                = "go-config"
	GoPublish               = "go-publish"
	GoRecursivePublish      = "go-recursive-publish"
	PipInstall              = "pip-install"
	PipConfig               = "pip-config"
	Ping                    = "ping"
	RtCurl                  = "rt-curl"
	ReleaseBundleCreate     = "release-bundle-create"
	ReleaseBundleUpdate     = "release-bundle-update"
	ReleaseBundleSign       = "release-bundle-sign"
	ReleaseBundleDistribute = "release-bundle-distribute"
	ReleaseBundleDelete     = "release-bundle-delete"
	TemplateConsumer        = "template-consumer"
	RepoDelete              = "repo-delete"
	ReplicationDelete       = "replication-delete"
	PermissionTargetDelete  = "permission-target-delete"
	AccessTokenCreate       = "access-token-create"
	UsersCreate             = "users-create"
	UsersDelete             = "users-delete"
	GroupCreate             = "group-create"
	GroupAddUsers           = "group-add-users"
	GroupDelete             = "group-delete"

	// MC's Commands Keys
	McConfig       = "mc-config"
	LicenseAcquire = "license-acquire"
	LicenseDeploy  = "license-deploy"
	LicenseRelease = "license-release"
	JpdAdd         = "jpd-add"
	JpdDelete      = "jpd-delete"
	// XRay's Commands Keys
	XrCurl        = "xr-curl"
	OfflineUpdate = "offline-update"

	// Config commands keys
	Config = "config"

	// *** Artifactory Commands' flags ***
	// Base flags
	url         = "url"
	distUrl     = "dist-url"
	user        = "user"
	password    = "password"
	apikey      = "apikey"
	accessToken = "access-token"
	serverId    = "server-id"

	// Deprecated flags
	deprecatedPrefix      = "deprecated-"
	deprecatedUrl         = deprecatedPrefix + url
	deprecatedUser        = deprecatedPrefix + user
	deprecatedPassword    = deprecatedPrefix + password
	deprecatedApikey      = deprecatedPrefix + apikey
	deprecatedAccessToken = deprecatedPrefix + accessToken
	deprecatedserverId    = deprecatedPrefix + serverId

	// Ssh flags
	sshKeyPath    = "ssh-key-path"
	sshPassPhrase = "ssh-passphrase"

	// Client certification flags
	clientCertPath    = "client-cert-path"
	clientCertKeyPath = "client-cert-key-path"
	insecureTls       = "insecure-tls"

	// Sort & limit flags
	sortBy    = "sort-by"
	sortOrder = "sort-order"
	limit     = "limit"
	offset    = "offset"

	// Spec flags
	spec     = "spec"
	specVars = "spec-vars"

	// Build info flags
	buildName   = "build-name"
	buildNumber = "build-number"
	module      = "module"

	// Generic commands flags
	excludePatterns  = "exclude-patterns"
	exclusions       = "exclusions"
	recursive        = "recursive"
	flat             = "flat"
	build            = "build"
	excludeArtifacts = "exclude-artifacts"
	includeDeps      = "include-deps"
	regexpFlag       = "regexp"
	retries          = "retries"
	dryRun           = "dry-run"
	explode          = "explode"
	includeDirs      = "include-dirs"
	props            = "props"
	targetProps      = "target-props"
	excludeProps     = "exclude-props"
	failNoOp         = "fail-no-op"
	threads          = "threads"
	syncDeletes      = "sync-deletes"
	quiet            = "quiet"
	bundle           = "bundle"
	archiveEntries   = "archive-entries"
	detailedSummary  = "detailed-summary"
	archive          = "archive"
	syncDeletesQuiet = syncDeletes + "-" + quiet
	antFlag          = "ant"
	fromRt           = "from-rt"
	transitive       = "transitive"

	// Config flags
	interactive   = "interactive"
	encPassword   = "enc-password"
	basicAuthOnly = "basic-auth-only"

	// Unique upload flags
	uploadPrefix          = "upload-"
	uploadExcludePatterns = uploadPrefix + excludePatterns
	uploadExclusions      = uploadPrefix + exclusions
	uploadRecursive       = uploadPrefix + recursive
	uploadFlat            = uploadPrefix + flat
	uploadRegexp          = uploadPrefix + regexpFlag
	uploadRetries         = uploadPrefix + retries
	uploadExplode         = uploadPrefix + explode
	uploadProps           = uploadPrefix + props
	uploadTargetProps     = uploadPrefix + targetProps
	uploadSyncDeletes     = uploadPrefix + syncDeletes
	uploadArchive         = uploadPrefix + archive
	deb                   = "deb"
	symlinks              = "symlinks"
	uploadAnt             = uploadPrefix + antFlag

	// Unique download flags
	downloadPrefix       = "download-"
	downloadRecursive    = downloadPrefix + recursive
	downloadFlat         = downloadPrefix + flat
	downloadRetries      = downloadPrefix + retries
	downloadExplode      = downloadPrefix + explode
	downloadProps        = downloadPrefix + props
	downloadExcludeProps = downloadPrefix + excludeProps
	downloadSyncDeletes  = downloadPrefix + syncDeletes
	minSplit             = "min-split"
	splitCount           = "split-count"
	validateSymlinks     = "validate-symlinks"

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

	// Unique build-publish flags
	buildPublishPrefix = "bp-"
	bpDryRun           = buildPublishPrefix + dryRun
	envInclude         = "env-include"
	envExclude         = "env-exclude"
	buildUrl           = "build-url"
	project            = "project"

	// Unique build-add-dependencies flags
	badPrefix    = "bad-"
	badDryRun    = badPrefix + dryRun
	badRecursive = badPrefix + recursive
	badRegexp    = badPrefix + regexpFlag
	badFromRt    = badPrefix + fromRt

	// Unique build-add-git flags
	configFlag = "config"

	// Unique build-scan flags
	fail = "fail"

	// Unique build-promote flags
	buildPromotePrefix  = "bpr-"
	bprDryRun           = buildPromotePrefix + dryRun
	bprProps            = buildPromotePrefix + props
	status              = "status"
	comment             = "comment"
	sourceRepo          = "source-repo"
	includeDependencies = "include-dependencies"
	copyFlag            = "copy"

	async = "async"

	// Unique build-distribute flags
	buildDistributePrefix = "bd-"
	bdDryRun              = buildDistributePrefix + dryRun
	bdAsync               = buildDistributePrefix + async
	sourceRepos           = "source-repos"
	passphrase            = "passphrase"
	publish               = "publish"
	override              = "override"

	// Unique build-discard flags
	buildDiscardPrefix = "bdi-"
	bdiAsync           = buildDiscardPrefix + async
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

	// Unique gradle-config flags
	usesPlugin          = "uses-plugin"
	useWrapper          = "use-wrapper"
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

	// Unique npm flags
	npmPrefix  = "npm-"
	npmThreads = npmPrefix + threads
	npmArgs    = "npm-args"

	// Unique nuget flags
	NugetArgs    = "nuget-args"
	SolutionRoot = "solution-root"
	// This flag is different than the nugetV2 since it is hidden and used in the 'nuget' cmd, and not the 'nugetc' cmd.
	LegacyNugetV2 = "nuget-v2-protocol"

	// Unique nuget/dotnet config flags
	nugetV2 = "nuget-v2"

	// Unique go flags
	deps        = "deps"
	noRegistry  = "no-registry"
	publishDeps = "publish-deps"
	// Deprecated.
	self = "self"

	// Unique release-bundle flags
	releaseBundlePrefix = "rb-"
	rbDryRun            = releaseBundlePrefix + dryRun
	rbRepo              = releaseBundlePrefix + repo
	rbPassphrase        = releaseBundlePrefix + passphrase
	distTarget          = releaseBundlePrefix + target
	sign                = "sign"
	desc                = "desc"
	releaseNotesPath    = "release-notes-path"
	releaseNotesSyntax  = "release-notes-syntax"
	distRules           = "dist-rules"
	site                = "site"
	city                = "city"
	countryCodes        = "country-codes"
	sync                = "sync"
	maxWaitMinutes      = "max-wait-minutes"
	deleteFromDist      = "delete-from-dist"

	// Template user flags
	vars = "vars"

	// User Management flags
	csv            = "csv"
	usersCreateCsv = "users-create-csv"
	usersDeleteCsv = "users-delete-csv"
	usersGroups    = "users-groups"
	replace        = "replace"

	// Unique access-token-create flags
	groups      = "groups"
	grantAdmin  = "grant-admin"
	expiry      = "expiry"
	refreshable = "refreshable"
	audience    = "audience"

	// *** Xray Commands' flags ***
	// Unique offline-update flags
	licenseId = "license-id"
	from      = "from"
	to        = "to"
	version   = "version"
	target    = "target"

	// *** Mission Control Commands' flags ***
	missionControlPrefix = "mc-"

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
	configXrUrl       = "xray-url"
	configMcUrl       = "mission-control-url"
	configPlUrl       = "pipelines-url"
	configAccessToken = configPrefix + accessToken
	configUser        = configPrefix + user
	configPassword    = configPrefix + password
	configApiKey      = configPrefix + apikey
	configInsecureTls = configPrefix + insecureTls
)

var flagsMap = map[string]cli.Flag{
	// Artifactory's commands Flags
	url: cli.StringFlag{
		Name:  url,
		Usage: "[Optional] Artifactory URL.` `",
	},
	deprecatedUrl: cli.StringFlag{
		Name:   url,
		Usage:  "[Deprecated] [Optional] Artifactory URL.` `",
		Hidden: true,
	},
	distUrl: cli.StringFlag{
		Name:  distUrl,
		Usage: "[Optional] Distribution URL.` `",
	},
	user: cli.StringFlag{
		Name:  user,
		Usage: "[Optional] Artifactory username.` `",
	},
	deprecatedUser: cli.StringFlag{
		Name:   user,
		Usage:  "[Deprecated] [Optional] Artifactory username.` `",
		Hidden: true,
	},
	password: cli.StringFlag{
		Name:  password,
		Usage: "[Optional] Artifactory password.` `",
	},
	deprecatedPassword: cli.StringFlag{
		Name:   password,
		Usage:  "[Deprecated] [Optional] Artifactory password.` `",
		Hidden: true,
	},
	apikey: cli.StringFlag{
		Name:  apikey,
		Usage: "[Optional] Artifactory API key.` `",
	},
	deprecatedApikey: cli.StringFlag{
		Name:   apikey,
		Usage:  "[Deprecated] [Optional] Artifactory API key.` `",
		Hidden: true,
	},
	accessToken: cli.StringFlag{
		Name:  accessToken,
		Usage: "[Optional] Artifactory access token.` `",
	},
	deprecatedAccessToken: cli.StringFlag{
		Name:   accessToken,
		Usage:  "[Deprecated] [Optional] Artifactory access token.` `",
		Hidden: true,
	},
	serverId: cli.StringFlag{
		Name:  serverId,
		Usage: "[Optional] Server ID configured using the config command.` `",
	},
	deprecatedserverId: cli.StringFlag{
		Name:   serverId,
		Usage:  "[Deprecated] [Optional] Artifactory server ID configured using the config command.` `",
		Hidden: true,
	},
	sshKeyPath: cli.StringFlag{
		Name:  sshKeyPath,
		Usage: "[Optional] SSH key file path.` `",
	},
	sshPassPhrase: cli.StringFlag{
		Name:  sshPassPhrase,
		Usage: "[Optional] SSH key passphrase.` `",
	},
	clientCertPath: cli.StringFlag{
		Name:  clientCertPath,
		Usage: "[Optional] Client certificate file in PEM format.` `",
	},
	clientCertKeyPath: cli.StringFlag{
		Name:  clientCertKeyPath,
		Usage: "[Optional] Private key file for the client certificate in PEM format.` `",
	},
	sortBy: cli.StringFlag{
		Name:  sortBy,
		Usage: "[Optional] A list of semicolon-separated fields to sort by. The fields must be part of the 'items' AQL domain. For more information, see https://www.jfrog.com/confluence/display/RTF/Artifactory+Query+Language#ArtifactoryQueryLanguage-EntitiesandFields` `",
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
	spec: cli.StringFlag{
		Name:  spec,
		Usage: "[Optional] Path to a File Spec.` `",
	},
	specVars: cli.StringFlag{
		Name:  specVars,
		Usage: "[Optional] List of variables in the form of \"key1=value1;key2=value2;...\" to be replaced in the File Spec. In the File Spec, the variables should be used as follows: ${key1}.` `",
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
	excludePatterns: cli.StringFlag{
		Name:   excludePatterns,
		Usage:  "[Optional] Semicolon-separated list of exclude patterns. Exclude patterns may contain the * and the ? wildcards. Unlike the Source path, it must not include the repository name at the beginning of the path.` `",
		Hidden: true,
	},
	exclusions: cli.StringFlag{
		Name:  exclusions,
		Usage: "[Optional] Semicolon-separated list of exclusions. Exclusions can include the * and the ? wildcards.` `",
	},
	uploadExcludePatterns: cli.StringFlag{
		Name:   excludePatterns,
		Usage:  "[Optional] Semicolon-separated list of exclude patterns. Exclude patterns may contain the * and the ? wildcards or a regex pattern, according to the value of the 'regexp' option.` `",
		Hidden: true,
	},
	uploadExclusions: cli.StringFlag{
		Name:  exclusions,
		Usage: "[Optional] Semicolon-separated list of exclude patterns. Exclude patterns may contain the * and the ? wildcards or a regex pattern, according to the value of the 'regexp' option.` `",
	},
	build: cli.StringFlag{
		Name:  build,
		Usage: "[Optional] If specified, only artifacts of the specified build are matched. The property format is build-name/build-number. If you do not specify the build number, the artifacts are filtered by the latest build number.` `",
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
		Usage: "[Default: " + strconv.Itoa(Threads) + "] Number of working threads.` `",
	},
	insecureTls: cli.BoolFlag{
		Name:  insecureTls,
		Usage: "[Default: false] Set to true to skip TLS certificates verification.` `",
	},
	bundle: cli.StringFlag{
		Name:  bundle,
		Usage: "[Optional] If specified, only artifacts of the specified bundle are matched. The value format is bundle-name/bundle-version.` `",
	},
	archiveEntries: cli.StringFlag{
		Name:  archiveEntries,
		Usage: "[Optional] If specified, only archive artifacts containing entries matching this pattern are matched. You can use wildcards to specify multiple artifacts.` `",
	},
	detailedSummary: cli.BoolFlag{
		Name:  detailedSummary,
		Usage: "[Default: false] Set to true to include a list of the affected files in the command summary.` `",
	},
	interactive: cli.BoolTFlag{
		Name:  interactive,
		Usage: "[Default: true, unless $CI is true] Set to false if you do not want the config command to be interactive. If true, the --url option becomes optional.` `",
	},
	encPassword: cli.BoolTFlag{
		Name:  encPassword,
		Usage: "[Default: true] If set to false then the configured password will not be encrypted using Artifactory's encryption API.` `",
	},
	basicAuthOnly: cli.BoolFlag{
		Name: basicAuthOnly,
		Usage: "[Default: false] Set to true to disable replacing username and password/API key with automatically created access token that's refreshed hourly. " +
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
	uploadFlat: cli.BoolTFlag{
		Name:  flat,
		Usage: "[Default: true] If set to false, files are uploaded according to their file system hierarchy.` `",
	},
	uploadRegexp: cli.BoolFlag{
		Name:  regexpFlag,
		Usage: "[Default: false] Set to true to use a regular expression instead of wildcards expression to collect files to upload.` `",
	},
	uploadAnt: cli.BoolFlag{
		Name:  antFlag,
		Usage: "[Default: false] Set to true to use an ant pattern instead of wildcards expression to collect files to upload.` `",
	},
	uploadRetries: cli.StringFlag{
		Name:  retries,
		Usage: "[Default: " + strconv.Itoa(Retries) + "] Number of upload retries.` `",
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
	uploadProps: cli.StringFlag{
		Name:   props,
		Usage:  "[Deprecated] [Optional] List of properties in the form of \"key1=value1;key2=value2,...\". Those properties will be attached to the uploaded artifacts.` `",
		Hidden: true,
	},
	uploadTargetProps: cli.StringFlag{
		Name:  props,
		Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\". Those properties will be attached to the uploaded artifacts.` `",
	},
	uploadSyncDeletes: cli.StringFlag{
		Name:  syncDeletes,
		Usage: "[Optional] Specific path in Artifactory, under which to sync artifacts after the upload. After the upload, this path will include only the artifacts uploaded during this upload operation. The other files under this path will be deleted.` `",
	},
	uploadArchive: cli.StringFlag{
		Name:  archive,
		Usage: "[Optional] Set to \"zip\" to deploy the files to Artifactory in a ZIP archive.` `",
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
	minSplit: cli.StringFlag{
		Name:  minSplit,
		Value: "",
		Usage: "[Default: " + strconv.Itoa(DownloadMinSplitKb) + "] Minimum file size in KB to split into ranges when downloading. Set to -1 for no splits.` `",
	},
	splitCount: cli.StringFlag{
		Name:  splitCount,
		Value: "",
		Usage: "[Default: " + strconv.Itoa(DownloadSplitCount) + "] Number of parts to split a file when downloading. Set to 0 for no splits.` `",
	},
	downloadRetries: cli.StringFlag{
		Name:  retries,
		Usage: "[Default: " + strconv.Itoa(Retries) + "] Number of download retries.` `",
	},
	downloadExplode: cli.BoolFlag{
		Name:  explode,
		Usage: "[Default: false] Set to true to extract an archive after it is downloaded from Artifactory.` `",
	},
	validateSymlinks: cli.BoolFlag{
		Name:  validateSymlinks,
		Usage: "[Default: false] Set to true to perform a checksum validation when downloading symbolic links.` `",
	},
	downloadProps: cli.StringFlag{
		Name:  props,
		Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\". Only artifacts with these properties will be downloaded.` `",
	},
	downloadExcludeProps: cli.StringFlag{
		Name:  excludeProps,
		Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\". Only artifacts without the specified properties will be downloaded.` `",
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
		Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\". Only artifacts with these properties will be moved.` `",
	},
	moveExcludeProps: cli.StringFlag{
		Name:  excludeProps,
		Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\". Only artifacts without the specified properties will be moved.` `",
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
		Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\". Only artifacts with these properties will be copied.` `",
	},
	copyExcludeProps: cli.StringFlag{
		Name:  excludeProps,
		Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\". Only artifacts without the specified properties will be copied.` `",
	},
	deleteRecursive: cli.BoolTFlag{
		Name:  recursive,
		Usage: "[Default: true] Set to false if you do not wish to delete artifacts inside sub-folders in Artifactory.` `",
	},
	deleteProps: cli.StringFlag{
		Name:  props,
		Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\". Only artifacts with these properties will be deleted.` `",
	},
	deleteExcludeProps: cli.StringFlag{
		Name:  excludeProps,
		Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\". Only artifacts without the specified properties will be deleted.` `",
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
		Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\". Only artifacts with these properties will be returned.` `",
	},
	searchExcludeProps: cli.StringFlag{
		Name:  excludeProps,
		Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\". Only artifacts without the specified properties will be returned` `",
	},
	searchTransitive: cli.BoolFlag{
		Name:  transitive,
		Usage: "[Default: false] Set to true to look for artifacts also in remote repositories. Available on Artifactory version 7.17.0 or higher.` `",
	},
	propsRecursive: cli.BoolTFlag{
		Name:  recursive,
		Usage: "[Default: true] When false, artifacts inside sub-folders in Artifactory will not be affected.` `",
	},
	propsProps: cli.StringFlag{
		Name:  props,
		Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\". Only artifacts with these properties are affected.` `",
	},
	propsExcludeProps: cli.StringFlag{
		Name:  excludeProps,
		Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\". Only artifacts without the specified properties are affected` `",
	},
	buildUrl: cli.StringFlag{
		Name:  buildUrl,
		Usage: "[Optional] Can be used for setting the CI server build URL in the build-info.` `",
	},
	project: cli.StringFlag{
		Name:  project,
		Usage: "[Optional] Artifactory project key.` `",
	},
	bpDryRun: cli.BoolFlag{
		Name:  dryRun,
		Usage: "[Default: false] Set to true to get a preview of the recorded build info, without publishing it to Artifactory.` `",
	},
	envInclude: cli.StringFlag{
		Name:  envInclude,
		Usage: "[Default: *] List of patterns in the form of \"value1;value2;...\" Only environment variables match those patterns will be included.` `",
	},
	envExclude: cli.StringFlag{
		Name:  envExclude,
		Usage: "[Default: *password*;*psw*;*secret*;*key*;*token*] List of case insensitive patterns in the form of \"value1;value2;...\". Environment variables match those patterns will be excluded.` `",
	},
	badRecursive: cli.BoolTFlag{
		Name:  recursive,
		Usage: "[Default: true] Set to false if you do not wish to collect artifacts in sub-folders to be added to the build info.` `",
	},
	badRegexp: cli.BoolFlag{
		Name:  regexpFlag,
		Usage: "[Default: false] Set to true to use a regular expression instead of wildcards expression to collect files to be added to the build info.` `",
	},
	badDryRun: cli.BoolFlag{
		Name:  dryRun,
		Usage: "[Default: false] Set to true to only get a summery of the dependencies that will be added to the build info.` `",
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
	status: cli.StringFlag{
		Name:  status,
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
		Usage: "[Default: false] If set to true, the build dependencies are also promoted.` `",
	},
	copyFlag: cli.BoolFlag{
		Name:  copyFlag,
		Usage: "[Default: false] If set true, the build artifacts and dependencies are copied to the target repository, otherwise they are moved.` `",
	},
	bprDryRun: cli.BoolFlag{
		Name:  dryRun,
		Usage: "[Default: false] If true, promotion is only simulated. The build is not promoted.` `",
	},
	bprProps: cli.StringFlag{
		Name:  props,
		Usage: "[Optional] List of properties in the form of \"key1=value1;key2=value2,...\". A list of properties to attach to the build artifacts.` `",
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
	sourceRepos: cli.StringFlag{
		Name:  sourceRepos,
		Usage: "[Optional] List of local repositories in the form of \"repo1,repo2,...\" from which build artifacts should be deployed.` `",
	},
	passphrase: cli.StringFlag{
		Name:  passphrase,
		Usage: "[Optional] If specified, Artifactory will GPG sign the build deployed to Bintray and apply the specified passphrase.` `",
	},
	publish: cli.BoolTFlag{
		Name:  publish,
		Usage: "[Default: true] If true, builds are published when deployed to Bintray.` `",
	},
	override: cli.BoolFlag{
		Name:  override,
		Usage: "[Default: false] If true, Artifactory overwrites builds already existing in the target path in Bintray.` `",
	},
	bdAsync: cli.BoolFlag{
		Name:  async,
		Usage: "[Default: false] If true, the build will be distributed asynchronously.` `",
	},
	bdDryRun: cli.BoolFlag{
		Name:  dryRun,
		Usage: "[Default: false] If true, distribution is only simulated. No files are actually moved.` `",
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
		Usage: "[Optional] List of build numbers in the form of \"value1,value2,...\", that should not be removed from Artifactory.` `",
	},
	deleteArtifacts: cli.BoolFlag{
		Name:  deleteArtifacts,
		Usage: "[Default: false] If set to true, automatically removes build artifacts stored in Artifactory.` `",
	},
	bdiAsync: cli.BoolFlag{
		Name:  async,
		Usage: "[Default: false] If set to true, build discard will run asynchronously and will not wait for response.` `",
	},
	refs: cli.StringFlag{
		Name:  refs,
		Usage: "[Default: refs/remotes/*] List of Git references in the form of \"ref1,ref2,...\" which should be preserved.` `",
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
		Usage: "[Optional] Artifactory server ID for resolution. The server should configured using the 'jfrog rt c' command.` `",
	},
	serverIdDeploy: cli.StringFlag{
		Name:  serverIdDeploy,
		Usage: "[Optional] Artifactory server ID for deployment. The server should configured using the 'jfrog rt c' command.` `",
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
	useWrapper: cli.BoolFlag{
		Name:  useWrapper,
		Usage: "[Default: false] Set to true if you'd like to use the Gradle wrapper.` `",
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
		Usage: "[Default: " + strconv.Itoa(Threads) + "] Number of threads for uploading build artifacts.` `",
	},
	skipLogin: cli.BoolFlag{
		Name:  skipLogin,
		Usage: "[Default: false] Set to true if you'd like the command to skip performing docker login.` `",
	},
	npmArgs: cli.StringFlag{
		Name:   npmArgs,
		Usage:  "[Deprecated] [Optional] A list of npm arguments and options in the form of \"--arg1=value1 --arg2=value2\"` `",
		Hidden: true,
	},
	npmThreads: cli.StringFlag{
		Name:  threads,
		Value: "",
		Usage: "[Default: 3] Number of working threads for build-info collection.` `",
	},
	NugetArgs: cli.StringFlag{
		Name:   NugetArgs,
		Usage:  "[Deprecated] [Optional] A list of NuGet arguments and options in the form of \"arg1 arg2 arg3\"` `",
		Hidden: true,
	},
	SolutionRoot: cli.StringFlag{
		Name:   SolutionRoot,
		Usage:  "[Deprecated] [Default: .] Path to the root directory of the solution. If the directory includes more than one sln files, then the first argument passed in the --nuget-args option should be the name (not the path) of the sln file.` `",
		Hidden: true,
	},
	LegacyNugetV2: cli.BoolFlag{
		Name:   LegacyNugetV2,
		Usage:  "[Deprecated] [Default: false] Set to true if you'd like to use the NuGet V2 protocol when restoring packages from Artifactory.` `",
		Hidden: true,
	},
	nugetV2: cli.BoolFlag{
		Name:  nugetV2,
		Usage: "[Default: false] Set to true if you'd like to use the NuGet V2 protocol when restoring packages from Artifactory.` `",
	},
	deps: cli.StringFlag{
		Name:  deps,
		Value: "",
		Usage: "[Optional] List of project dependencies in the form of \"dep1-name:version,dep2-name:version...\" to be published to Artifactory. Use \"ALL\" to publish all dependencies.` `",
	},
	self: cli.BoolTFlag{
		Name:   self,
		Usage:  "[Deprecated] [Default: true] Set false to skip publishing the project package zip file to Artifactory..` `",
		Hidden: true,
	},
	noRegistry: cli.BoolFlag{
		Name:   noRegistry,
		Usage:  "[Deprecated] [Default: false] Set to true if you don't want to use Artifactory as your proxy` `",
		Hidden: true,
	},
	publishDeps: cli.BoolFlag{
		Name:   publishDeps,
		Usage:  "[Deprecated] [Default: false] Set to true if you wish to publish missing dependencies to Artifactory` `",
		Hidden: true,
	},
	rbDryRun: cli.BoolFlag{
		Name:  dryRun,
		Usage: "[Default: false] Set to true to disable communication with JFrog Distribution.` `",
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
		Usage: "[Optional] The passphrase for the signing key. ` `",
	},
	distTarget: cli.StringFlag{
		Name: target,
		Usage: "[Optional] The target path for distributed artifacts on the edge node. If not specified, the artifacts will have the same path and name on the edge node, as on the source Artifactory server. " +
			"For flexibility in specifying the distribution path, you can include placeholders in the form of {1}, {2} which are replaced by corresponding tokens in the pattern path that are enclosed in parenthesis. ` `",
	},
	rbRepo: cli.StringFlag{
		Name:  repo,
		Usage: "[Optional] A repository name at source Artifactory to store release bundle artifacts in. If not provided, Artifactory will use the default one.` `",
	},
	distRules: cli.StringFlag{
		Name:  distRules,
		Usage: "Path to distribution rules.` `",
	},
	site: cli.StringFlag{
		Name:  site,
		Usage: "[Default: '*'] Wildcard filter for site name. ` `",
	},
	city: cli.StringFlag{
		Name:  city,
		Usage: "[Default: '*'] Wildcard filter for site city name. ` `",
	},
	countryCodes: cli.StringFlag{
		Name:  countryCodes,
		Usage: "[Default: '*'] Semicolon-separated list of wildcard filters for site country codes. ` `",
	},
	sync: cli.BoolFlag{
		Name:  sync,
		Usage: "[Default: false] Set to true to enable sync distribution (the command execution will end when the distribution process ends).` `",
	},
	maxWaitMinutes: cli.StringFlag{
		Name:  maxWaitMinutes,
		Usage: "[Default: 60] Max minutes to wait for sync distribution. ` `",
	},
	deleteFromDist: cli.BoolFlag{
		Name:  deleteFromDist,
		Usage: "[Default: false] Set to true to delete release bundle version in JFrog Distribution itself after deletion is complete in the specified Edge node/s.` `",
	},
	targetProps: cli.StringFlag{
		Name:  targetProps,
		Usage: "[Optional] The list of properties, in the form of key1=value1;key2=value2,..., to be added to the artifacts after distribution of the release bundle.` `",
	},
	vars: cli.StringFlag{
		Name:  vars,
		Usage: "[Optional] List of variables in the form of \"key1=value1;key2=value2;...\" to be replaced in the template. In the template, the variables should be used as follows: ${key1}.` `",
	},
	groups: cli.StringFlag{
		Name: groups,
		Usage: "[Default: *] A list of comma-separated groups for the access token to be associated with. " +
			"Specify * to indicate that this is a 'user-scoped token', i.e., the token provides the same access privileges that the current subject has, and is therefore evaluated dynamically. " +
			"A non-admin user can only provide a scope that is a subset of the groups to which he belongs` `",
	},
	grantAdmin: cli.BoolFlag{
		Name:  grantAdmin,
		Usage: "[Default: false] Set to true to provides admin privileges to the access token. This is only available for administrators.` `",
	},
	expiry: cli.StringFlag{
		Name:  expiry,
		Usage: "[Default: " + strconv.Itoa(TokenExpiry) + "] The time in seconds for which the token will be valid. To specify a token that never expires, set to zero. Non-admin can only set a value that is equal to or less than the default 3600.` `",
	},
	refreshable: cli.BoolFlag{
		Name:  refreshable,
		Usage: "[Default: false] Set to true if you'd like the the token to be refreshable. A refresh token will also be returned in order to be used to generate a new token once it expires.` `",
	},
	audience: cli.StringFlag{
		Name:  audience,
		Usage: "[Optional] A space-separate list of the other Artifactory instances or services that should accept this token identified by their Artifactory Service IDs, as obtained by the 'jfrog rt curl api/system/service_id' command.` `",
	},
	usersCreateCsv: cli.StringFlag{
		Name:  csv,
		Usage: "[Mandatory] Path to a csv file with the users' details. The first row of the file is reserved for the cells' headers. It must include \"username\",\"password\",\"email\"` `",
	},
	usersDeleteCsv: cli.StringFlag{
		Name:  csv,
		Usage: "[Optional] Path to a csv file with the users' details. The first row of the file is reserved for the cells' headers. It must include \"username\"` `",
	},
	usersGroups: cli.StringFlag{
		Name:  usersGroups,
		Usage: "[Optional] A list of comma-separated groups for the new users to be associated with.` `",
	},
	replace: cli.BoolFlag{
		Name:  replace,
		Usage: "[Optional] Set to true if you'd like existing users or groups to be replaced.` `",
	},
	// Xray's commands Flags
	licenseId: cli.StringFlag{
		Name:  licenseId,
		Usage: "[Mandatory] Xray license ID` `",
	},
	from: cli.StringFlag{
		Name:  from,
		Usage: "[Optional] From update date in YYYY-MM-DD format.` `",
	},
	to: cli.StringFlag{
		Name:  to,
		Usage: "[Optional] To update date in YYYY-MM-DD format.` `",
	},
	version: cli.StringFlag{
		Name:  version,
		Usage: "[Optional] Xray API version.` `",
	},
	target: cli.StringFlag{
		Name:  target,
		Usage: "[Default: ./] Path for downloaded update files.` `",
	},
	// Mission Control's commands Flags
	mcUrl: cli.StringFlag{
		Name:  url,
		Usage: "[Optional] Mission Control URL.` `",
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
		Usage: "[Default: " + strconv.Itoa(DefaultLicenseCount) + "] The number of licenses to deploy. Minimum value is 1.` `",
	},
	imageFile: cli.StringFlag{
		Name:  imageFile,
		Usage: "[Mandatory] Path to a file which includes one line in the following format: <IMAGE-TAG>@sha256:<MANIFEST-SHA256>.` `",
	},
	// Config commands Flags
	configPlatformUrl: cli.StringFlag{
		Name:  url,
		Usage: "[Optional] JFrog platform URL.` `",
	},
	configRtUrl: cli.StringFlag{
		Name:  configRtUrl,
		Usage: "[Optional] Artifactory URL.` `",
	},
	configXrUrl: cli.StringFlag{
		Name:  configXrUrl,
		Usage: "[Optional] Xray URL.` `",
	},
	configMcUrl: cli.StringFlag{
		Name:  configMcUrl,
		Usage: "[Optional] Mission Control URL.` `",
	},
	configPlUrl: cli.StringFlag{
		Name:  configPlUrl,
		Usage: "[Optional] Pipelines URL.` `",
	},
	configUser: cli.StringFlag{
		Name:  user,
		Usage: "[Optional] JFrog Platform username. ` `",
	},
	configPassword: cli.StringFlag{
		Name:  password,
		Usage: "[Optional] JFrog Platform password. ` `",
	},
	configApiKey: cli.StringFlag{
		Name:  apikey,
		Usage: "[Optional] JFrog Platform API key. ` `",
	},
	configAccessToken: cli.StringFlag{
		Name:  accessToken,
		Usage: "[Optional] JFrog Platform access token. ` `",
	},
	configInsecureTls: cli.StringFlag{
		Name:  insecureTls,
		Usage: "[Default: false] Set to true to skip TLS certificates verification, while encrypting the Artifactory password during the config process.` `",
	},
}

var commandFlags = map[string][]string{
	Config: {
		interactive, encPassword, configPlatformUrl, configRtUrl, distUrl, configXrUrl, configMcUrl, configPlUrl, configUser, configPassword, configApiKey, configAccessToken, sshKeyPath, clientCertPath,
		clientCertKeyPath, basicAuthOnly, configInsecureTls,
	},
	// Deprecated
	RtConfig: {
		interactive, encPassword, url, distUrl, user, password, apikey, accessToken, sshKeyPath, clientCertPath,
		clientCertKeyPath, basicAuthOnly, configInsecureTls,
	},
	DeleteConfig: {
		deleteQuiet,
	},
	Upload: {
		url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath, serverId, clientCertPath, targetProps,
		clientCertKeyPath, spec, specVars, buildName, buildNumber, module, uploadExcludePatterns, uploadExclusions, deb,
		uploadRecursive, uploadFlat, uploadRegexp, uploadRetries, dryRun, uploadExplode, symlinks, includeDirs,
		uploadProps, failNoOp, threads, uploadSyncDeletes, syncDeletesQuiet, insecureTls, detailedSummary, project,
		uploadAnt, uploadArchive,
	},
	Download: {
		url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath, serverId, clientCertPath,
		clientCertKeyPath, spec, specVars, buildName, buildNumber, module, excludePatterns, exclusions, sortBy,
		sortOrder, limit, offset, downloadRecursive, downloadFlat, build, includeDeps, excludeArtifacts, minSplit, splitCount,
		downloadRetries, dryRun, downloadExplode, validateSymlinks, bundle, includeDirs, downloadProps, downloadExcludeProps,
		failNoOp, threads, archiveEntries, downloadSyncDeletes, syncDeletesQuiet, insecureTls, detailedSummary, project,
	},
	Move: {
		url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath, serverId, clientCertPath,
		clientCertKeyPath, spec, specVars, excludePatterns, exclusions, sortBy, sortOrder, limit, offset, moveRecursive,
		moveFlat, dryRun, build, includeDeps, excludeArtifacts, moveProps, moveExcludeProps, failNoOp, threads, archiveEntries, insecureTls,
	},
	Copy: {
		url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath, serverId, clientCertPath,
		clientCertKeyPath, spec, specVars, excludePatterns, exclusions, sortBy, sortOrder, limit, offset, copyRecursive,
		copyFlat, dryRun, build, includeDeps, excludeArtifacts, bundle, copyProps, copyExcludeProps, failNoOp, threads, archiveEntries, insecureTls,
	},
	Delete: {
		url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath, serverId, clientCertPath,
		clientCertKeyPath, spec, specVars, excludePatterns, exclusions, sortBy, sortOrder, limit, offset,
		deleteRecursive, dryRun, build, includeDeps, excludeArtifacts, deleteQuiet, deleteProps, deleteExcludeProps, failNoOp, threads, archiveEntries,
		insecureTls,
	},
	Search: {
		url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath, serverId, clientCertPath,
		clientCertKeyPath, spec, specVars, excludePatterns, exclusions, sortBy, sortOrder, limit, offset,
		searchRecursive, build, includeDeps, excludeArtifacts, count, bundle, includeDirs, searchProps, searchExcludeProps, failNoOp, archiveEntries,
		insecureTls, searchTransitive,
	},
	Properties: {
		url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath, serverId, clientCertPath,
		clientCertKeyPath, spec, specVars, excludePatterns, exclusions, sortBy, sortOrder, limit, offset,
		propsRecursive, build, includeDeps, excludeArtifacts, bundle, includeDirs, failNoOp, threads, archiveEntries, propsProps, propsExcludeProps,
		insecureTls,
	},
	BuildPublish: {
		url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath, serverId, buildUrl, bpDryRun,
		envInclude, envExclude, insecureTls, project,
	},
	BuildAppend: {
		url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath, serverId, buildUrl, bpDryRun,
		envInclude, envExclude, insecureTls, project,
	},
	BuildAddDependencies: {
		spec, specVars, uploadExcludePatterns, uploadExclusions, badRecursive, badRegexp, badDryRun, project, badFromRt, serverId,
	},
	BuildAddGit: {
		configFlag, serverId, project,
	},
	BuildCollectEnv: {
		project,
	},
	BuildDockerCreate: {
		buildName, buildNumber, module, url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath,
		serverId, imageFile, project,
	},
	BuildScan: {
		url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath, serverId, fail, insecureTls,
		project,
	},
	BuildPromote: {
		url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath, serverId, status, comment,
		sourceRepo, includeDependencies, copyFlag, bprDryRun, bprProps, insecureTls, project,
	},
	BuildDistribute: {
		url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath, serverId, sourceRepos, passphrase,
		publish, override, bdAsync, bdDryRun, insecureTls,
	},
	BuildDiscard: {
		url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath, serverId, maxDays, maxBuilds,
		excludeBuilds, deleteArtifacts, bdiAsync, insecureTls, project,
	},
	GitLfsClean: {
		url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath, serverId, refs, glcRepo, glcDryRun,
		glcQuiet, insecureTls,
	},
	MvnConfig: {
		global, serverIdResolve, serverIdDeploy, repoResolveReleases, repoResolveSnapshots, repoDeployReleases, repoDeploySnapshots,
	},
	GradleConfig: {
		global, serverIdResolve, serverIdDeploy, repoResolve, repoDeploy, usesPlugin, useWrapper, deployMavenDesc,
		deployIvyDesc, ivyDescPattern, ivyArtifactsPattern,
	},
	Mvn: {
		buildName, buildNumber, deploymentThreads, insecureTls, project,
	},
	Gradle: {
		buildName, buildNumber, deploymentThreads, project,
	},
	DockerPromote: {
		targetDockerImage, sourceTag, targetTag, dockerPromoteCopy, url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath,
		serverId,
	},
	ContainerPush: {
		buildName, buildNumber, module, url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath,
		serverId, skipLogin, threads, project,
	},
	ContainerPull: {
		buildName, buildNumber, module, url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath,
		serverId, skipLogin, project,
	},
	NpmConfig: {
		global, serverIdResolve, serverIdDeploy, repoResolve, repoDeploy,
	},
	Npm: {
		npmArgs, deprecatedUrl, deprecatedUser, deprecatedPassword, deprecatedApikey, deprecatedAccessToken, buildName,
		buildNumber, module, npmThreads, project,
	},
	NpmPublish: {
		npmArgs, deprecatedUrl, deprecatedUser, deprecatedPassword, deprecatedApikey, deprecatedAccessToken, buildName,
		buildNumber, module, project,
	},
	NugetConfig: {
		global, serverIdResolve, repoResolve, nugetV2,
	},
	Nuget: {
		NugetArgs, SolutionRoot, LegacyNugetV2, deprecatedUrl, deprecatedUser, deprecatedPassword, deprecatedApikey,
		deprecatedAccessToken, buildName, buildNumber, module, project,
	},
	DotnetConfig: {
		global, serverIdResolve, repoResolve, nugetV2,
	},
	Dotnet: {
		buildName, buildNumber, module, project,
	},
	GoConfig: {
		global, serverIdResolve, serverIdDeploy, repoResolve, repoDeploy,
	},
	GoPublish: {
		deps, self, url, user, password, apikey, accessToken, deprecatedserverId, buildName, buildNumber, module, project,
	},
	Go: {
		noRegistry, publishDeps, deprecatedUrl, deprecatedUser, deprecatedPassword, deprecatedApikey,
		deprecatedAccessToken, buildName, buildNumber, module, project,
	},
	GoRecursivePublish: {
		url, user, password, apikey, accessToken, serverId,
	},
	Ping: {
		url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath, serverId, clientCertPath,
		clientCertKeyPath, insecureTls,
	},
	RtCurl: {
		serverId,
	},
	PipConfig: {
		global, serverIdResolve, repoResolve,
	},
	PipInstall: {
		buildName, buildNumber, module, project,
	},
	ReleaseBundleCreate: {
		url, distUrl, user, password, apikey, accessToken, sshKeyPath, sshPassPhrase, serverId, spec, specVars, targetProps,
		rbDryRun, sign, desc, exclusions, releaseNotesPath, releaseNotesSyntax, rbPassphrase, rbRepo, insecureTls, distTarget,
	},
	ReleaseBundleUpdate: {
		url, distUrl, user, password, apikey, accessToken, sshKeyPath, sshPassPhrase, serverId, spec, specVars, targetProps,
		rbDryRun, sign, desc, exclusions, releaseNotesPath, releaseNotesSyntax, rbPassphrase, rbRepo, insecureTls, distTarget,
	},
	ReleaseBundleSign: {
		url, distUrl, user, password, apikey, accessToken, sshKeyPath, sshPassPhrase, serverId, rbPassphrase, rbRepo,
		insecureTls,
	},
	ReleaseBundleDistribute: {
		url, distUrl, user, password, apikey, accessToken, sshKeyPath, sshPassPhrase, serverId, rbDryRun, distRules,
		site, city, countryCodes, sync, maxWaitMinutes, insecureTls,
	},
	ReleaseBundleDelete: {
		url, distUrl, user, password, apikey, accessToken, sshKeyPath, sshPassPhrase, serverId, rbDryRun, distRules,
		site, city, countryCodes, sync, maxWaitMinutes, insecureTls, deleteFromDist, deleteQuiet,
	},
	TemplateConsumer: {
		url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath, serverId, clientCertPath,
		clientCertKeyPath, vars,
	},
	RepoDelete: {
		url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath, serverId, clientCertPath,
		clientCertKeyPath, deleteQuiet,
	},
	ReplicationDelete: {
		url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath, serverId, clientCertPath,
		clientCertKeyPath, deleteQuiet,
	},
	PermissionTargetDelete: {
		url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath, serverId, clientCertPath,
		clientCertKeyPath, deleteQuiet,
	},
	AccessTokenCreate: {
		url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath, serverId, clientCertPath,
		clientCertKeyPath, groups, grantAdmin, expiry, refreshable, audience,
	},
	UsersCreate: {
		url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath, serverId,
		usersCreateCsv, usersGroups, replace,
	},
	UsersDelete: {
		url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath, serverId,
		usersDeleteCsv, deleteQuiet,
	},
	GroupCreate: {
		url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath, serverId,
		replace,
	},
	GroupAddUsers: {
		url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath, serverId,
	},
	GroupDelete: {
		url, user, password, apikey, accessToken, sshPassPhrase, sshKeyPath, serverId, deleteQuiet,
	},
	// Xray's commands
	OfflineUpdate: {
		licenseId, from, to, version, target,
	},
	XrCurl: {
		serverId,
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
}

func GetCommandFlags(cmd string) []cli.Flag {
	flagList, ok := commandFlags[cmd]
	if !ok {
		log.Error("The command \"", cmd, "\" does not found in commands flags map.")
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

// This function is used for mvn and gradle command validation
func GetBasicBuildToolsFlags() (flags []cli.Flag) {
	basicBuildToolsFlags := []string{url, distUrl, user, password, apikey, accessToken, serverId}
	return buildAndSortFlags(basicBuildToolsFlags)
}

var deprecatedFlags = []string{deprecatedUrl, deprecatedUser, deprecatedPassword, deprecatedApikey, deprecatedAccessToken}

// This function is used for legacy (deprecated) nuget command validation
func GetLegacyNugetFlags() (flags []cli.Flag) {
	legacyNugetFlags := []string{NugetArgs, SolutionRoot, LegacyNugetV2}
	legacyNugetFlags = append(legacyNugetFlags, deprecatedFlags...)
	return buildAndSortFlags(legacyNugetFlags)
}

// This function is used for legacy (deprecated) npm command validation
func GetLegacyNpmFlags() (flags []cli.Flag) {
	legacyNpmFlags := append(deprecatedFlags, npmArgs)
	return buildAndSortFlags(legacyNpmFlags)
}

// This function is used for legacy (deprecated) go command validation
func GetLegacyGoFlags() (flags []cli.Flag) {
	legacyGoFlags := []string{noRegistry, publishDeps}
	legacyGoFlags = append(legacyGoFlags, deprecatedFlags...)
	return buildAndSortFlags(legacyGoFlags)
}
