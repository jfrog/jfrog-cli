package tests

import (
	"path/filepath"

	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	servicesUtils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	clientUtils "github.com/jfrog/jfrog-client-go/utils"
)

const (
	AccessEndpoint                                        = "access/"
	ArchiveEntriesDownload                                = "archive_entries_download_spec.json"
	ArchiveEntriesUpload                                  = "archive_entries_upload_spec.json"
	ArtifactoryEndpoint                                   = "artifactory/"
	LifecycleEndpoint                                     = "lifecycle/"
	BuildAddDepsDoubleRemoteSpec                          = "build_add_deps_double_remote_spec.json"
	BuildAddDepsDoubleSpec                                = "build_add_deps_double_spec.json"
	BuildAddDepsRemoteSpec                                = "build_add_deps_simple_remote_spec.json"
	BuildAddDepsSpec                                      = "build_add_deps_simple_spec.json"
	BuildDownloadSpec                                     = "build_download_spec.json"
	BuildDownloadSpecDepsOnly                             = "build_download_spec_deps_only.json"
	BuildDownloadSpecExcludeArtifacts                     = "build_download_spec_exclude_artifacts.json"
	BuildDownloadSpecIncludeDeps                          = "build_download_spec_include_deps.json"
	BuildDownloadSpecNoBuildNumber                        = "build_download_spec_no_build_number.json"
	BuildDownloadSpecNoBuildNumberWithSort                = "build_download_spec_no_build_number_with_sort.json"
	BuildDownloadSpecNoPattern                            = "build_download_spec_no_pattern.json"
	BundleDownloadGpgSpec                                 = "bundle_download_gpg_spec.json"
	BundleDownloadSpec                                    = "bundle_download_spec.json"
	BundleDownloadSpecNoPattern                           = "bundle_download_spec_no_pattern.json"
	CopyByBuildPatternAllSpec                             = "move_copy_delete_by_build_pattern_all_spec.json"
	CopyByBuildSpec                                       = "move_copy_delete_by_build_spec.json"
	CopyByBundleAssertSpec                                = "copy_by_bundle_assert_spec.json"
	CopyByBundleSpec                                      = "copy_by_bundle_spec.json"
	CopyItemsSpec                                         = "copy_items_spec.json"
	CopyMoveSimpleSpec                                    = "copy_move_simple.json"
	CpMvDlByBuildAssertSpec                               = "copy_by_build_assert_spec.json"
	DebianTestRepositoryConfig                            = "debian_test_repository_config.json"
	DebianUploadSpec                                      = "upload_debian_spec.json"
	DelSpecExclusions                                     = "delete_spec_exclusions.json"
	DeleteSimpleSpec                                      = "delete_simple_spec.json"
	DeleteSpec                                            = "delete_spec.json"
	DeleteSpecWildcardInRepo                              = "delete_spec_wildcard.json"
	DevRepoRepositoryConfig                               = "dev_repo_repository_config.json"
	DistributionCreateByAql                               = "dist_create_by_aql.json"
	DistributionCreateWithMapping                         = "dist_create_with_mapping.json"
	DistributionCreateWithPatternAndTarget                = "dist_create_with_pattern_and_target.json"
	DistributionMappingDownload                           = "dist_mapping_download_spec.json"
	DistributionRepoConfig1                               = "dist_repository_config1.json"
	DistributionRepoConfig2                               = "dist_repository_config2.json"
	DistributionRules                                     = "distribution_rules.json"
	DistributionSetDeletePropsSpec                        = "dist_set_delete_props_spec.json"
	DistributionUploadSpecA                               = "dist_upload_spec_a.json"
	DistributionUploadSpecB                               = "dist_upload_spec_b.json"
	DockerLocalPromoteRepositoryConfig                    = "docker_local_promote_repository_config.json"
	DockerLocalRepositoryConfig                           = "docker_local_repository_config.json"
	DockerRemoteRepositoryConfig                          = "docker_remote_repository_config.json"
	DockerVirtualRepositoryConfig                         = "docker_virtual_repository_config.json"
	DownloadAllRepo1TestResources                         = "download_all_repo1_test_resources.json"
	DownloadAndExplodeArchives                            = "download_and_explode_archives.json"
	DownloadEmptyDirs                                     = "download_empty_dir_spec.json"
	DownloadSpecExclusions                                = "download_spec_exclusions.json"
	DownloadWildcardRepo                                  = "download_wildcard_repo.json"
	DownloadWithoutExplodeArchives                        = "download_without_explode_archives.json"
	GitLfsAssertSpec                                      = "git_lfs_assert_spec.json"
	GitLfsTestRepositoryConfig                            = "git_lfs_test_repository_config.json"
	GoLocalRepositoryConfig                               = "go_local_repository_config.json"
	GoPublishRepoExcludes                                 = "go_publish_repo_excludes.json"
	GoRemoteRepositoryConfig                              = "go_remote_repository_config.json"
	GoVirtualRepositoryConfig                             = "go_virtual_repository_config.json"
	GradleConfig                                          = "gradle.yaml"
	GradleRemoteRepositoryConfig                          = "gradle_remote_repository_config.json"
	GradleRepositoryConfig                                = "gradle_repository_config.json"
	GradleServerIDUsesPluginConfig                        = "gradle_server_id_uses_plugin.yaml"
	HttpsProxyEnvVar                                      = "PROXY_HTTPS_PORT"
	KanikoConfig                                          = "kaniko_config.json"
	LifecycleAql                                          = "lifecycle-aql.json"
	LifecycleArtifacts                                    = "lifecycle-artifacts.json"
	LifecycleBuilds12                                     = "lifecycle-builds-1-2.json"
	LifecycleBuilds3                                      = "lifecycle-builds-3.json"
	LifecycleReleaseBundles                               = "lifecycle-release-bundles.json"
	MavenConfig                                           = "maven.yaml"
	MavenIncludeExcludePatternsConfig                     = "maven_include_exclude_patterns.yaml"
	MavenRemoteRepositoryConfig                           = "maven_remote_repository_config.json"
	MavenRepositoryConfig1                                = "maven_repository_config1.json"
	MavenRepositoryConfig2                                = "maven_repository_config2.json"
	MavenWithoutDeployerConfig                            = "maven_without_deployer.yaml"
	MoveCopySpecExclusions                                = "move_copy_spec_exclusions.json"
	NpmLocalRepositoryConfig                              = "npm_local_repository_config.json"
	NpmRemoteRepositoryConfig                             = "npm_remote_repository_config.json"
	NugetRemoteRepositoryConfig                           = "nuget_remote_repository_config.json"
	Out                                                   = "out"
	PipenvRemoteRepositoryConfig                          = "pipenv_remote_repository_config.json"
	PipenvVirtualRepositoryConfig                         = "pipenv_virtual_repository_config.json"
	ProdRepo1RepositoryConfig                             = "prod_repo1_repository_config.json"
	ProdRepo2RepositoryConfig                             = "prod_repo2_repository_config.json"
	PypiLocalRepositoryConfig                             = "pypi_local_repository_config.json"
	PypiRemoteRepositoryConfig                            = "pypi_remote_repository_config.json"
	PypiVirtualRepositoryConfig                           = "pypi_virtual_repository_config.json"
	ReplicationTempCreate                                 = "replication_push_create.json"
	Repo1RepositoryConfig                                 = "repo1_repository_config.json"
	Repo2RepositoryConfig                                 = "repo2_repository_config.json"
	RepoDetailsUrl                                        = "api/repositories/"
	SearchAllDevRepo                                      = "search_all_dev_repo.json"
	SearchAllGradle                                       = "search_all_gradle.json"
	SearchAllMaven                                        = "search_all_maven.json"
	SearchAllNpm                                          = "search_all_npm.json"
	SearchAllProdRepo1                                    = "search_all_prod_repo1.json"
	SearchAllProdRepo2                                    = "search_all_prod_repo2.json"
	SearchAllRepo1                                        = "search_all_repo1.json"
	SearchDistRepoByInSuffix                              = "search_dist_repo_by_in_suffix.json"
	SearchPromotedDocker                                  = "search_promoted_docker.json"
	SearchRepo1ByInSuffix                                 = "search_repo1_by_in_suffix.json"
	SearchRepo1IncludeDirs                                = "search_repo1_include_dirs.json"
	SearchRepo1NonExistFile                               = "search_repo1_ant_test_file.json"
	SearchRepo1NonExistFileAntExclusions                  = "search_repo1_ant_and_exclusions_test_file.json"
	SearchRepo1TestResources                              = "search_repo1_test_resources.json"
	SearchRepo2                                           = "search_repo2.json"
	SearchSimplePlaceholders                              = "search_simple_placeholders.json"
	SearchTargetInRepo2                                   = "search_target_in_repo2.json"
	SearchTxt                                             = "search_txt.json"
	ServerId                                              = "testServerId"
	SetDeletePropsSpec                                    = "set_delete_props_spec.json"
	SplitUploadSpecA                                      = "upload_split_spec_a.json"
	SplitUploadSpecB                                      = "upload_split_spec_b.json"
	Temp                                                  = "tmp"
	TerraformLocalRepositoryConfig                        = "terraform_local_repository_config.json"
	UploadAntPattern                                      = "upload_ant_pattern.json"
	UploadAntPatternExclusions                            = "upload_ant_pattern_exclusions.json"
	UploadAsArchive                                       = "upload_as_archive.json"
	UploadAsArchiveEmptyDirs                              = "upload_archive_empty_dir_spec.json"
	UploadAsArchiveToDir                                  = "upload_as_archive_to_dir.json"
	UploadDevSpec                                         = "upload_dev_spec.json"
	UploadDevSpecA                                        = "upload_dev_spec_a.json"
	UploadDevSpecB                                        = "upload_dev_spec_b.json"
	UploadEmptyDirs                                       = "upload_empty_dir_spec.json"
	UploadExcludeIncludeDir                               = "upload_exclude_include_dir.json"
	UploadExcludeIncludeDirAntPattern                     = "upload_exclude_include_dir_ant_pattern.json"
	UploadExcludeIncludeDirAntPattern2                    = "upload_exclude_include_dir_ant_pattern2.json"
	UploadExcludeIncludeDirs                              = "upload_exclude_include_dirs.json"
	UploadExcludeIncludeDirsFlat                          = "upload_exclude_include_dirs_flat.json"
	UploadExcludeRepo                                     = "upload_exclude_repo.json"
	UploadExcludeRepoContent                              = "upload_exclude_repo_content.json"
	UploadExcludeRepoContentPart2                         = "upload_exclude_repo_content_part_2.json"
	UploadFileWithParenthesesAndPlaceholdersAndRegexpSpec = "upload_file_with_parentheses_and_placeholders_and_regexp.json"
	UploadFileWithParenthesesAndPlaceholdersSpec          = "upload_file_with_parentheses_and_placeholders.json"
	UploadFileWithParenthesesAndRegexpSpec                = "upload_file_with_parentheses_and_regexp.json"
	UploadFileWithParenthesesSpec                         = "upload_file_with_parentheses.json"
	UploadFlatNonRecursive                                = "upload_flat_non_recursive.json"
	UploadFlatRecursive                                   = "upload_flat_recursive.json"
	UploadMultipleFileSpecs                               = "upload_multiple_file_specs.json"
	UploadPrefixFiles                                     = "upload_prefix_files.json"
	UploadSimplePlaceholders                              = "upload_simple_placeholders.json"
	UploadSpecExclude                                     = "upload_spec_exclude.json"
	UploadSpecExcludeRegex                                = "upload_spec_exclude_regex.json"
	UploadTempWildcard                                    = "upload_temp_wildcard.json"
	UploadWithPropsSpec                                   = "upload_with_props_spec.json"
	UploadWithPropsSpecDeleteExcludeProps                 = "upload_with_props_spec_delete_exclude_props.json"
	UploadWorkingDirectoryAsArchive                       = "upload_archive_wd.json"
	VirtualRepositoryConfig                               = "specs_virtual_repository_config.json"
	WinBuildAddDepsSpec                                   = "win_simple_build_add_deps_spec.json"
	WinSimpleDownloadSpec                                 = "win_simple_download_spec.json"
	WinSimpleUploadSpec                                   = "win_simple_upload_spec.json"
	XrayEndpoint                                          = "xray/"
	YarnRemoteRepositoryConfig                            = "yarn_remote_repository_config.json"
	ReleaseLifecycleImportDependencySpec                  = "import_bundle_repo_dependency.json"
)

var (
	// Repositories
	DistRepo1                      = "cli-dist1"
	DistRepo2                      = "cli-dist2"
	GoRepo                         = "cli-go"
	GoRemoteRepo                   = "cli-go-remote"
	GoVirtualRepo                  = "cli-go-virtual"
	TerraformRepo                  = "cli-terraform"
	GradleRepo                     = "cli-gradle"
	MvnRemoteRepo                  = "cli-mvn-remote"
	GradleRemoteRepo               = "cli-gradle-remote"
	MvnRepo1                       = "cli-mvn1"
	MvnRepo2                       = "cli-mvn2"
	NpmRepo                        = "cli-npm"
	NpmRemoteRepo                  = "cli-npm-remote"
	NugetRemoteRepo                = "cli-nuget-remote"
	YarnRemoteRepo                 = "cli-yarn-remote"
	PypiLocalRepo                  = "cli-pypi-local"
	PypiRemoteRepo                 = "cli-pypi-remote"
	PypiVirtualRepo                = "cli-pypi-virtual"
	PipenvRemoteRepo               = "cli-pipenv-pypi-remote"
	PipenvVirtualRepo              = "cli-pipenv-pypi-virtual"
	DockerLocalRepo                = "cli-docker-local"
	DockerLocalPromoteRepo         = "cli-docker-local-promote"
	DockerRemoteRepo               = "cli-docker-remote"
	DockerVirtualRepo              = "cli-docker-virtual"
	RtDebianRepo                   = "cli-debian"
	ReleaseLifecycleDependencyRepo = "npm-prod"
	RtLfsRepo                      = "cli-lfs"
	RtRepo1                        = "cli-rt1"
	RtRepo2                        = "cli-rt2"
	RtVirtualRepo                  = "cli-rt-virtual"
	// Repositories that are assigned to an environment.
	RtDevRepo   = "cli-rt-dev"
	RtProdRepo1 = "cli-rt-prod1"
	RtProdRepo2 = "cli-rt-prod2"
	// These are not actual repositories. These patterns are meant to be used in both Repo1 and Repo2.
	RtRepo1And2            = "cli-rt*"
	RtRepo1And2Placeholder = "cli-rt(*)"

	BundleName                  = "cli-dist-bundle"
	DockerBuildName             = "cli-docker-build"
	DockerImageName             = "cli-docker-image"
	DotnetBuildName             = "cli-dotnet-build"
	GoBuildName                 = "cli-go-build"
	GradleBuildName             = "cli-gradle-build"
	MvnBuildName                = "cli-maven-build"
	NpmBuildName                = "cli-npm-build"
	YarnBuildName               = "cli-yarn-build"
	NuGetBuildName              = "cli-nuget-build"
	PipBuildName                = "cli-pip-build"
	PipenvBuildName             = "cli-pipenv-build"
	RtBuildName1                = "cli-rt-build1"
	RtBuildName2                = "cli-rt-build2"
	RtBuildNameWithSpecialChars = "cli-rt-a$+~&^a#-build3"
	RtPermissionTargetName      = "cli-rt-pt"
	LcBuildName1                = "cli-lc-build1"
	LcBuildName2                = "cli-lc-build2"
	LcBuildName3                = "cli-lc-build3"
	LcRbName1                   = "cli-lc-rb1"
	LcRbName2                   = "cli-lc-rb2"
	LcRbName3                   = "cli-lc-rb3"
	GoPublishWithExclusionPath  = "github.com/jfrog/dependency/@v/github.com/jfrog/dependency@v1.1.1/"

	// Users
	UserName1 = "alice"
	// jfrog-ignore - not a real password
	Password1 = "A12356789z"
	UserName2 = "bob"
	// jfrog-ignore - not a real password
	Password2  = "1B234578y9"
	ProjectKey = "prj"
)

func GetTxtUploadExpectedRepo1() []string {
	return []string{
		RtRepo1 + "/cliTestFile.txt",
	}
}

func GetSimpleUploadExpectedRepo1() []string {
	return []string{
		RtRepo1 + "/test_resources/a3.in",
		RtRepo1 + "/test_resources/a1.in",
		RtRepo1 + "/test_resources/a2.in",
		RtRepo1 + "/test_resources/b2.in",
		RtRepo1 + "/test_resources/b3.in",
		RtRepo1 + "/test_resources/b1.in",
		RtRepo1 + "/test_resources/c2.in",
		RtRepo1 + "/test_resources/c1.in",
		RtRepo1 + "/test_resources/c3.in",
	}
}

func GetExpectedExcludeUpload() []string {
	return []string{
		RtRepo1 + "/a2.in",
		RtRepo1 + "/a3.in",
		RtRepo1 + "/a1.in",
		RtRepo1 + "/",
	}
}
func GetExpectedExcludeUploadPart2() []string {
	return []string{
		RtRepo1 + "/a2.in",
		RtRepo1 + "/a3.in",
		RtRepo1 + "/a1.in",
		RtRepo1 + "/b3.in",
		RtRepo1 + "/c",
		RtRepo1 + "/",
	}
}
func GetExpectedExcludeUpload2() []string {
	return []string{
		RtRepo1 + "/b3.in",
		RtRepo1 + "/a2.in",
		RtRepo1 + "/a3.in",
		RtRepo1 + "/a1.in",
		RtRepo1 + "/c",
		RtRepo1 + "/",
	}
}
func GetExpectedExcludeUploadIncludeDir() []string {
	return []string{
		RtRepo1 + "/a2.in",
		RtRepo1 + "/a3.in",
		RtRepo1 + "/a1.in",
		RtRepo1 + "/b",
		RtRepo1 + "/",
	}
}

func GetUploadLegacyPropsExpected() []string {
	return []string{
		RtRepo1 + "/data/a1.in",
		RtRepo1 + "/data/a2.in",
		RtRepo1 + "/data/a3.in",
	}
}

func GetSearchAppendedBuildNoPatternExpected() []string {
	return []string{
		RtRepo1 + "/data/a1.in",
		RtRepo1 + "/data/a2.in",
		RtRepo1 + "/data/a3.in",
	}
}

func GetSimpleWildcardUploadExpectedRepo1() []string {
	return []string{
		RtRepo1 + "/upload_simple_wildcard/github.com/github.in",
	}
}

func GetSimpleAntPatternUploadExpectedRepo1() []string {
	return []string{
		RtRepo1 + "/bitbucket.in",
		RtRepo1 + "/github.in",
	}
}

func GetAntPatternUploadWithExclusionsExpectedRepo1() []string {
	return []string{
		RtRepo1 + "/a1.in",
		RtRepo1 + "/a2.in",
		RtRepo1 + "/a3.in",
		RtRepo1 + "/c1.in",
		RtRepo1 + "/c2.in",
		RtRepo1 + "/c3.in",
	}
}

func GetAntPatternUploadWithIncludeDirsExpectedRepo1() []string {
	return []string{
		RtRepo1 + "/",
		RtRepo1 + "/folder",
		RtRepo1 + "/testdata",
		RtRepo1 + "/testdata/an",
		RtRepo1 + "/testdata/an/empty",
		RtRepo1 + "/testdata/an/empty/folder",
	}
}

func GetSimpleUploadSpecialCharNoRegexExpectedRepo1() []string {
	return []string{
		RtRepo1 + "/a1.in",
	}
}

func GetSimpleUploadFilteredRepo1() []string {
	return []string{
		RtRepo1 + "/a.txt",
	}
}

func GetSimpleUploadSpecialCharNoRegexExpectedRepo2() []string {
	return []string{
		RtRepo2 + "/a1.in",
	}
}

func GetSimpleUploadSpecialCharNoRegexExpected2filesRepo1() []string {
	return []string{
		RtRepo1 + "/a1.in",
		RtRepo1 + "/a2.in",
	}
}

func GetSimpleUploadSpecialCharNoRegexExpected2filesRepo2() []string {
	return []string{
		RtRepo2 + "/a1.in",
		RtRepo2 + "/a2.in",
	}
}

func GetUploadSpecExcludeRepo1() []string {
	return []string{
		RtRepo1 + "/a1.in",
		RtRepo1 + "/b1.in",
		RtRepo1 + "/c2.in",
		RtRepo1 + "/c3.in",
	}
}

func GetUploadDebianExpected() []string {
	return []string{
		RtDebianRepo + "/data/a1.in",
		RtDebianRepo + "/data/a2.in",
		RtDebianRepo + "/data/a3.in",
	}
}

func GetPrefixFilesCopy() []string {
	return []string{
		RtRepo2 + "/prefix-a",
		RtRepo2 + "/prefix-ab",
		RtRepo2 + "/prefix-abb",
	}
}

func GetSingleFileCopy() []string {
	return []string{
		RtRepo2 + "/path/a1.in",
	}
}

func GetSingleFileCopyFullPath() []string {
	return []string{
		RtRepo2 + "/path/inner/a1.in",
	}
}

func GetSingleInnerFileCopyFullPath() []string {
	return []string{
		RtRepo2 + "/path/path/inner/a1.in",
	}
}

func GetFolderCopyTwice() []string {
	return []string{
		RtRepo2 + "/path/inner/a1.in",
		RtRepo2 + "/path/path/inner/a1.in",
	}
}

func GetFolderCopyIntoFolder() []string {
	return []string{
		RtRepo2 + "/path/path/inner/a1.in",
	}
}

func GetSingleDirectoryCopyFlat() []string {
	return []string{
		RtRepo2 + "/inner/a1.in",
	}
}

func GetAnyItemCopy() []string {
	return []string{
		RtRepo2 + "/path/inner/a1.in",
		RtRepo2 + "/someFile",
	}
}

func GetAnyItemCopyRecursive() []string {
	return []string{
		RtRepo2 + "/a/b/a1.in",
		RtRepo2 + "/aFile",
	}
}

func GetCopyFolderRename() []string {
	return []string{
		RtRepo2 + "/newPath/inner/a1.in",
	}
}

func GetAnyItemCopyUsingSpec() []string {
	return []string{
		RtRepo2 + "/a1.in",
	}
}

func GetExplodeUploadExpectedRepo1() []string {
	return []string{
		RtRepo1 + "/a/a3.in",
		RtRepo1 + "/a/a1.in",
		RtRepo1 + "/a/a2.in",
		RtRepo1 + "/a/b/b1.in",
		RtRepo1 + "/a/b/b2.in",
		RtRepo1 + "/a/b/b3.in",
		RtRepo1 + "/a/b/c/c1.in",
		RtRepo1 + "/a/b/c/c2.in",
		RtRepo1 + "/a/b/c/c3.in",
	}
}

func GetCopyFileNameWithParentheses() []string {
	return []string{
		RtRepo2 + "/testdata/b/(/(.in",
		RtRepo2 + "/testdata/b/(b/(b.in",
		RtRepo2 + "/testdata/b/)b/)b.in",
		RtRepo2 + "/testdata/b/b(/b(.in",
		RtRepo2 + "/testdata/b/b)/b).in",
		RtRepo2 + "/testdata/b/(b)/(b).in",
		RtRepo2 + "/testdata/b/)b)/)b).in",
		RtRepo2 + "/(/b(.in",
		RtRepo2 + "/()/(b.in",
		RtRepo2 + "/()/testdata/b/(b)/(b).in",
		RtRepo2 + "/(/(.in.zip",
		RtRepo2 + "/(/in-b(",
		RtRepo2 + "/(/b(.in-up",
		RtRepo2 + "/c/(.in.zip",
	}
}

func GetUploadFileNameWithParentheses() []string {
	return []string{
		RtRepo1 + "/(.in",
		RtRepo1 + "/(b.in",
		RtRepo1 + "/)b.in",
		RtRepo1 + "/b(.in",
		RtRepo1 + "/b).in",
		RtRepo1 + "/(b).in",
		RtRepo1 + "/)b).in",
		RtRepo1 + "/(new)/testdata/b/(/(.in",
		RtRepo1 + "/(new)/testdata/b/(b/(b.in",
		RtRepo1 + "/(new)/testdata/b/b(/b(.in",
		RtRepo1 + "/new)/testdata/b/b)/b).in",
		RtRepo1 + "/new)/testdata/b/(b)/(b).in",
		RtRepo1 + "/(new/testdata/b/)b)/)b).in",
		RtRepo1 + "/(new/testdata/b/)b/)b.in",
		RtRepo1 + "/new(new)/testdata/b/b(b)/c(c).in",
		RtRepo1 + "/new(new)/testdata/b/b(b)/d.in",
	}
}

func GetUploadFileNameWithParenthesesAndRegexp() []string {
	return []string{
		RtRepo1 + "/regexp/testdata/b/b(b)/d.in",
		RtRepo1 + "/regexp/testdata/b/b(b)/c(c).in",
		RtRepo1 + "/regexp/testdata/b/(b)/(b).in",
		RtRepo1 + "/regexp/testdata/b/(b/(b.in",
	}
}

func GetUploadFileNameWithParenthesesAndPlaceholdersAndRegexp() []string {
	return []string{
		RtRepo1 + "/emptyParentheses/testdata/b/(b)/(b).in",
		RtRepo1 + "/regexpAndPlaceholder/c(c)/c(c).in",
		RtRepo1 + "/regexpAndPlaceholder/d/d.in",
		RtRepo1 + "/regexpAndPlaceholder/in/testdata/(b).in",
		RtRepo1 + "/regexpAndPlaceholder/(b)-(b)/(b).in",
		RtRepo1 + "/NoRegexpAndPlaceholder/testdata/b/(b)/(b).in",
	}
}

func GetUploadFileNameWithParenthesesAndPlaceholders() []string {
	return []string{
		RtRepo1 + "/data-in/(b).in",
		RtRepo1 + "/c(c)/c(c).in",
		RtRepo1 + "/d/d.in",
		RtRepo1 + "/in/(b.in",
	}
}

func GetUploadWorkingDirAsArchive() []string {
	return []string{
		Out,
		filepath.Join(Out, "archive"),
		filepath.Join(Out, "archive/wd"),
		filepath.Join(Out, "archive/wd/b1.in"),
		filepath.Join(Out, "archive/wd/b2.in"),
		filepath.Join(Out, "archive/wd/b3.in"),
		filepath.Join(Out, "archive/wd/c"),
		filepath.Join(Out, "archive/wd/c/c1.in"),
		filepath.Join(Out, "archive/wd/c/c2.in"),
		filepath.Join(Out, "archive/wd/c/c3.in"),
	}
}

func GetUploadAsArchive() []string {
	return []string{
		RtRepo1 + "/archive/a.zip",
		RtRepo1 + "/archive/b.zip",
	}
}

func GetDownloadArchiveAndExplode() []string {
	return []string{
		filepath.Join(Out, "archive/a/a1.in"),
		filepath.Join(Out, "archive/a/a2.in"),
		filepath.Join(Out, "archive/a/a3.in"),
		filepath.Join(Out, "archive/a/b1.in"),
		filepath.Join(Out, "archive/b/b1.in"),
		filepath.Join(Out, "archive/b/b2.in"),
		filepath.Join(Out, "archive/b/b3.in"),
	}
}

func GetDownloadArchiveAndExplodeWithIncludeDirs() []string {
	return []string{
		filepath.Join(Out, "archive/archive/c"),
		filepath.Join(Out, "archive/archive/folder"),
	}
}

func GetMoveCopySpecExpected() []string {
	return []string{
		RtRepo2 + "/copy_move_target/a1.in",
		RtRepo2 + "/copy_move_target/a2.in",
		RtRepo2 + "/copy_move_target/a3.in",
		RtRepo2 + "/copy_move_target/b/b1.in",
		RtRepo2 + "/copy_move_target/b/b2.in",
		RtRepo2 + "/copy_move_target/b/b3.in",
		RtRepo2 + "/copy_move_target/b/c/c1.in",
		RtRepo2 + "/copy_move_target/b/c/c2.in",
		RtRepo2 + "/copy_move_target/b/c/c3.in",
	}
}

func GetRepo1TestResourcesExpected() []string {
	return []string{
		RtRepo1 + "/test_resources/a1.in",
		RtRepo1 + "/test_resources/a2.in",
		RtRepo1 + "/test_resources/a3.in",
		RtRepo1 + "/test_resources/b/b1.in",
		RtRepo1 + "/test_resources/b/b2.in",
		RtRepo1 + "/test_resources/b/b3.in",
		RtRepo1 + "/test_resources/b/c/c1.in",
		RtRepo1 + "/test_resources/b/c/c2.in",
		RtRepo1 + "/test_resources/b/c/c3.in",
	}
}

func GetBuildBeforeCopyExpected() []string {
	return GetBuildBeforeMoveExpected()
}

func GetBuildCopyExpected() []string {
	return []string{
		RtRepo1 + "/data/a1.in",
		RtRepo1 + "/data/a2.in",
		RtRepo1 + "/data/a3.in",
		RtRepo1 + "/data/b1.in",
		RtRepo1 + "/data/b2.in",
		RtRepo1 + "/data/b3.in",
		RtRepo2 + "/data/a1.in",
		RtRepo2 + "/data/a2.in",
		RtRepo2 + "/data/a3.in",
	}
}

func GetBundleCopyExpected() []string {
	return []string{
		DistRepo1 + "/data/a1.in",
		DistRepo1 + "/data/a2.in",
		DistRepo1 + "/data/a3.in",
		DistRepo1 + "/data/b1.in",
		DistRepo1 + "/data/b2.in",
		DistRepo1 + "/data/b3.in",
		DistRepo2 + "/data/a1.in",
		DistRepo2 + "/data/a2.in",
		DistRepo2 + "/data/a3.in",
	}
}

func GetBundlePropsExpected() []string {
	return []string{
		DistRepo1 + "/data/b1.in",
		DistRepo1 + "/data/b2.in",
		DistRepo1 + "/data/b3.in",
	}
}

func GetBundleMappingExpected() []string {
	return []string{
		DistRepo2 + "/target/b1.in",
		DistRepo2 + "/target/b2.in",
		DistRepo2 + "/target/b3.in",
	}
}

func GetGitLfsExpected() []string {
	return []string{
		RtLfsRepo + "/objects/4b/f4/4bf4c8c0fef3f5c8cf6f255d1c784377138588c0a9abe57e440bce3ccb350c2e",
	}
}

func GetBuildBeforeMoveExpected() []string {
	return []string{
		RtRepo1 + "/data/b1.in",
		RtRepo1 + "/data/b2.in",
		RtRepo1 + "/data/b3.in",
		RtRepo1 + "/data/a1.in",
		RtRepo1 + "/data/a2.in",
		RtRepo1 + "/data/a3.in",
	}
}

func GetBuildMoveExpected() []string {
	return []string{
		RtRepo1 + "/data/b1.in",
		RtRepo1 + "/data/b2.in",
		RtRepo1 + "/data/b3.in",
		RtRepo2 + "/data/a1.in",
		RtRepo2 + "/data/a2.in",
		RtRepo2 + "/data/a3.in",
	}
}

func GetBuildCopyExclude() []string {
	return []string{
		RtRepo1 + "/data/a1.in",
		RtRepo1 + "/data/a2.in",
		RtRepo1 + "/data/a3.in",
		RtRepo1 + "/data/b1.in",
		RtRepo1 + "/data/b2.in",
		RtRepo1 + "/data/b3.in",
		RtRepo2 + "/data/a1.in",
		RtRepo2 + "/data/a2.in",
		RtRepo2 + "/data/a3.in",
	}
}

func GetBuildDeleteExpected() []string {
	return []string{
		RtRepo1 + "/data/b1.in",
		RtRepo1 + "/data/b2.in",
		RtRepo1 + "/data/b3.in",
	}
}

func GetExtractedDownload() []string {
	return []string{
		filepath.Join(Out, "randFile"),
		filepath.Join(Out, "concurrent.tar.gz"),
	}
}

func GetExtractedDownloadCurDir() []string {
	return []string{
		filepath.Join(Out, "p-modules"),
		filepath.Join(Out, "p-modules", "DownloadAndExplodeCurDirTarget"),
	}
}

func GetArchiveConcurrent() []string {
	return []string{
		filepath.Join(Out, "a.zip"),
	}
}

func GetExtractedDownloadFlatFalse() []string {
	return []string{
		filepath.Join(Out, "checkFlat", "file1"),
		filepath.Join(Out, "checkFlat", "dir", "flat.tar"),
	}
}

func GetExtractedDownloadTarFileFlatFalse() []string {
	return []string{
		filepath.Join(Out, "checkFlat", "dir", "file1"),
	}
}

func GetExtractedDownloadTarFileSpecialChars() []string {
	return []string{
		filepath.Join(Out, "dir $+~&^a# test", "dir", "file $+~&^a#1"),
		filepath.Join(Out, "dir $+~&^a# test", "file $+~&^a#1"),
	}
}

func GetExtractedDownloadConcurrent() []string {
	return []string{
		filepath.Join(Out, "a/a1.in"),
		filepath.Join(Out, "a/a2.in"),
		filepath.Join(Out, "a/a3.in"),
		filepath.Join(Out, "a/b/b1.in"),
		filepath.Join(Out, "a/b/b2.in"),
		filepath.Join(Out, "a/b/b3.in"),
		filepath.Join(Out, "a/b/c/c1.in"),
		filepath.Join(Out, "a/b/c/c2.in"),
		filepath.Join(Out, "a/b/c/c3.in"),
	}
}

func GetFileWithParenthesesDownload() []string {
	return []string{
		filepath.Join(Out, "testdata"),
		filepath.Join(Out, "testdata/b"),
		filepath.Join(Out, "testdata/b/("),
		filepath.Join(Out, "testdata/b/(/(.in"),
		filepath.Join(Out, "testdata/b/(b"),
		filepath.Join(Out, "testdata/b/(b/(b.in"),
		filepath.Join(Out, "testdata/b/(b)"),
		filepath.Join(Out, "testdata/b/(b)/(b).in"),
		filepath.Join(Out, "testdata/b/)b"),
		filepath.Join(Out, "testdata/b/)b/)b.in"),
		filepath.Join(Out, "testdata/b/)b)"),
		filepath.Join(Out, "testdata/b/)b)/)b).in"),
		filepath.Join(Out, "testdata/b/b("),
		filepath.Join(Out, "testdata/b/b(/b(.in"),
		filepath.Join(Out, "testdata/b/b)"),
		filepath.Join(Out, "testdata/b/b)/b).in"),
	}
}

func GetVirtualDownloadExpected() []string {
	return []string{
		filepath.Join(Out, "a/a1.in"),
		filepath.Join(Out, "a/a2.in"),
		filepath.Join(Out, "a/a3.in"),
		filepath.Join(Out, "a/b/b1.in"),
		filepath.Join(Out, "a/b/b2.in"),
		filepath.Join(Out, "a/b/b3.in"),
		filepath.Join(Out, "a/b/c/c1.in"),
		filepath.Join(Out, "a/b/c/c2.in"),
		filepath.Join(Out, "a/b/c/c3.in"),
	}
}

func GetExpectedSyncDeletesDownloadStep2() []string {
	localPathPrefix := filepath.Join("syncDir", "testdata", "a")
	return []string{
		filepath.Join(Out, localPathPrefix, "a1.in"),
		filepath.Join(Out, localPathPrefix, "a2.in"),
		filepath.Join(Out, localPathPrefix, "a3.in"),
		filepath.Join(Out, localPathPrefix, "b/b1.in"),
		filepath.Join(Out, localPathPrefix, "b/b2.in"),
		filepath.Join(Out, localPathPrefix, "b/b3.in"),
		filepath.Join(Out, localPathPrefix, "b/c/c1.in"),
		filepath.Join(Out, localPathPrefix, "b/c/c2.in"),
		filepath.Join(Out, localPathPrefix, "b/c/c3.in"),
	}
}

func GetExpectedSyncDeletesDownloadStep3() []string {
	return []string{
		filepath.Join(Out, "a1.in"),
		filepath.Join(Out, "a2.in"),
		filepath.Join(Out, "a3.in"),
		filepath.Join(Out, "b1.in"),
		filepath.Join(Out, "b2.in"),
		filepath.Join(Out, "b3.in"),
		filepath.Join(Out, "c1.in"),
		filepath.Join(Out, "c2.in"),
		filepath.Join(Out, "c3.in"),
	}
}

func GetExpectedSyncDeletesDownloadStep4() []string {
	return []string{
		filepath.Join(Out, "a2.in"),
		filepath.Join(Out, "b2.in"),
		filepath.Join(Out, "c2.in"),
	}
}

func GetSyncExpectedDeletesDownloadStep5() []string {
	localPathPrefix := filepath.Join("syncDir", "testdata", "a")
	return []string{
		filepath.Join(Out, localPathPrefix, "a1.in"),
		filepath.Join(Out, localPathPrefix, "a2.in"),
		filepath.Join(Out, localPathPrefix, "a3.in"),
		filepath.Join(Out, localPathPrefix, "b/b1.in"),
		filepath.Join(Out, localPathPrefix, "b/b2.in"),
		filepath.Join(Out, localPathPrefix, "b/b3.in"),
	}
}

func GetSyncExpectedDeletesDownloadStep6() []string {
	localPathPrefix := "/syncDir/testdata/archives/"
	return []string{
		RtRepo1 + localPathPrefix + "a.zip",
		RtRepo1 + localPathPrefix + "b.zip",
		RtRepo1 + localPathPrefix + "c.zip",
		RtRepo1 + localPathPrefix + "d.zip",
	}
}

func GetSyncExpectedDeletesDownloadStep7() []string {
	localPathPrefix := filepath.Join("syncDir", "testdata", "archives")
	return []string{
		filepath.Join(Out, localPathPrefix, "a.zip"),
		filepath.Join(Out, localPathPrefix, "b.zip"),
		filepath.Join(Out, localPathPrefix, "c.zip"),
		filepath.Join(Out, localPathPrefix, "d.zip"),
	}
}

func GetDownloadWildcardRepo() []string {
	return []string{
		RtRepo1 + "/path/a1.in",
		RtRepo2 + "/path/a2.in",
	}
}

func GetDownloadUnicode() []string {
	return []string{
		RtRepo1 + "/testdata/unicode/dirλrectory/文件.in",
		RtRepo1 + "/testdata/unicode/dirλrectory/aȩ.ȥ1",
		RtRepo1 + "/testdata/unicode/Ԙחלص.in",
	}
}

func GetBuildDownload() []string {
	return []string{
		Out,
		filepath.Join(Out, "download"),
		filepath.Join(Out, "download", "aql_by_build"),
		filepath.Join(Out, "download", "aql_by_build", "data"),
		filepath.Join(Out, "download", "aql_by_build", "data", "a1.in"),
		filepath.Join(Out, "download", "aql_by_build", "data", "a2.in"),
		filepath.Join(Out, "download", "aql_by_build", "data", "a3.in"),
	}
}

func GetBuildDownloadDoesntExist() []string {
	return []string{
		Out,
	}
}

func GetBuildDownloadByShaAndBuild() []string {
	return []string{
		Out,
		filepath.Join(Out, "download"),
		filepath.Join(Out, "download", "aql_by_build"),
		filepath.Join(Out, "download", "aql_by_build", "data"),
		filepath.Join(Out, "download", "aql_by_build", "data", "a10.in"),
	}
}

func GetBuildDownloadByShaAndBuildName() []string {
	return []string{
		Out,
		filepath.Join(Out, "download"),
		filepath.Join(Out, "download", "aql_by_build"),
		filepath.Join(Out, "download", "aql_by_build", "data"),
		filepath.Join(Out, "download", "aql_by_build", "data", "a11.in"),
	}
}

func GetBuildSimpleDownload() []string {
	return []string{
		Out,
		filepath.Join(Out, "download"),
		filepath.Join(Out, "download", "simple_by_build"),
		filepath.Join(Out, "download", "simple_by_build", "data"),
		filepath.Join(Out, "download", "simple_by_build", "data", "b1.in"),
	}
}

func GetBuildSimpleDownloadNoPattern() []string {
	return []string{
		Out,
		filepath.Join(Out, "download"),
		filepath.Join(Out, "download", "simple_by_build"),
		filepath.Join(Out, "download", "simple_by_build", "data"),
		filepath.Join(Out, "download", "simple_by_build", "data", "a1.in"),
		filepath.Join(Out, "download", "simple_by_build", "data", "a2.in"),
		filepath.Join(Out, "download", "simple_by_build", "data", "a3.in"),
	}
}

func GetDownloadAppendedBuild() []string {
	return []string{
		filepath.Join(Out, "download"),
		filepath.Join(Out, "download", "simple_by_build"),
		filepath.Join(Out, "download", "simple_by_build", "data"),
		filepath.Join(Out, "download", "simple_by_build", "data", "a1.in"),
		filepath.Join(Out, "download", "simple_by_build", "data", "a2.in"),
		filepath.Join(Out, "download", "simple_by_build", "data", "a3.in"),
	}
}

func GetDownloadByBuildOnlyDeps() []string {
	return []string{
		Out,
		filepath.Join(Out, "download"),
		filepath.Join(Out, "download", "download_build_only_dependencies"),
		filepath.Join(Out, "download", "download_build_only_dependencies", "b1.in"),
		filepath.Join(Out, "download", "download_build_only_dependencies", "b2.in"),
		filepath.Join(Out, "download", "download_build_only_dependencies", "b3.in"),
	}
}

func GetDownloadByBuildIncludeDeps() []string {
	return []string{
		filepath.Join(Out, "download", "download_build_with_dependencies"),
		filepath.Join(Out, "download", "download_build_with_dependencies", "a1.in"),
		filepath.Join(Out, "download", "download_build_with_dependencies", "a2.in"),
		filepath.Join(Out, "download", "download_build_with_dependencies", "a3.in"),
		filepath.Join(Out, "download", "download_build_with_dependencies", "b1.in"),
		filepath.Join(Out, "download", "download_build_with_dependencies", "b2.in"),
		filepath.Join(Out, "download", "download_build_with_dependencies", "b3.in"),
	}
}

func GetBuildExcludeDownload() []string {
	return []string{
		Out,
		filepath.Join(Out, "download"),
		filepath.Join(Out, "download", "aql_by_artifacts"),
		filepath.Join(Out, "download", "aql_by_artifacts", "data"),
		filepath.Join(Out, "download", "aql_by_artifacts", "data", "a3.in"),
		filepath.Join(Out, "download", "aql_by_artifacts", "data", "b1.in"),
		filepath.Join(Out, "download", "aql_by_artifacts", "data", "b2.in"),
		filepath.Join(Out, "download", "aql_by_artifacts", "data", "b3.in"),
		filepath.Join(Out, "download", "aql_by_artifacts", "data", "c1.in"),
		filepath.Join(Out, "download", "aql_by_artifacts", "data", "c3.in"),
	}
}

func GetBuildExcludeDownloadBySpec() []string {
	return []string{
		Out,
		filepath.Join(Out, "download"),
		filepath.Join(Out, "download", "aql_by_artifacts"),
		filepath.Join(Out, "download", "aql_by_artifacts", "data"),
		filepath.Join(Out, "download", "aql_by_artifacts", "data", "a2.in"),
		filepath.Join(Out, "download", "aql_by_artifacts", "data", "b1.in"),
		filepath.Join(Out, "download", "aql_by_artifacts", "data", "b2.in"),
		filepath.Join(Out, "download", "aql_by_artifacts", "data", "b3.in"),
		filepath.Join(Out, "download", "aql_by_artifacts", "data", "c1.in"),
		filepath.Join(Out, "download", "aql_by_artifacts", "data", "c3.in"),
	}
}

func GetCleanBuild() []string {
	return []string{
		filepath.Join(Out, "clean-build"),
		filepath.Join(Out, "clean-build", "data"),
		filepath.Join(Out, "clean-build", "data", "b1.in"),
		filepath.Join(Out, "clean-build", "data", "b2.in"),
		filepath.Join(Out, "clean-build", "data", "b3.in"),
	}
}

func GetMultipleFileSpecs() []string {
	return []string{
		RtRepo1 + "/multiple/a1.out",
		RtRepo1 + "/multiple/properties/testdata/a/b/b2.in",
	}
}

func GetSimplePlaceholders() []string {
	return []string{
		RtRepo1 + "/simple_placeholders/a-in.out",
		RtRepo1 + "/simple_placeholders/b/b-in.out",
		RtRepo1 + "/simple_placeholders/b/c/c-in.out",
	}
}

func GetSimpleDelete() []string {
	return []string{
		RtRepo1 + "/test_resources/a1.in",
		RtRepo1 + "/test_resources/a2.in",
		RtRepo1 + "/test_resources/a3.in",
	}
}

func GetDeleteFolderWithWildcard() []string {
	return []string{
		RtRepo1 + "/test_resources/a1.in",
		RtRepo1 + "/test_resources/a2.in",
		RtRepo1 + "/test_resources/a3.in",
		RtRepo1 + "/test_resources/b/b1.in",
		RtRepo1 + "/test_resources/b/b2.in",
		RtRepo1 + "/test_resources/b/b3.in",
	}
}

func GetSearchIncludeDirsFiles() []utils.SearchResult {
	return []utils.SearchResult{
		{
			Path: RtRepo1 + "/",
			Type: "folder",
			Size: 0,
		},
		{
			Path: RtRepo1 + "/data",
			Type: "folder",
			Size: 0,
		},
		{
			Path: RtRepo1 + "/data/testdata",
			Type: "folder",
			Size: 0,
		},
		{
			Path: RtRepo1 + "/data/testdata/a",
			Type: "folder",
			Size: 0,
		},
		{
			Path:   RtRepo1 + "/data/testdata/a/a1.in",
			Type:   "file",
			Size:   7,
			Sha1:   "507ac63c6b0f650fb6f36b5621e70ebca3b0965c",
			Sha256: "4eb341b5d2762a853d79cc25e622aa8b978eb6e12c3259e2d99dc9dc60d82c5d",
			Md5:    "65298e78fe5883eee82056bc6d0d7f4c",
		},
		{
			Path:   RtRepo1 + "/data/testdata/a/a2.in",
			Type:   "file",
			Size:   7,
			Sha1:   "de2f31d77e2c2b1039a806f21b0c5f3243e45588",
			Sha256: "3e3deb6628658a48cf0d280a2210211f9d977ec2e10a4619b95d5fb85cb10450",
			Md5:    "28f9732cb82a2d11760e38614246ad6d",
		},
		{
			Path:   RtRepo1 + "/data/testdata/a/a3.in",
			Type:   "file",
			Size:   7,
			Sha1:   "29d38faccfe74dee60d0142a716e8ea6fad67b49",
			Sha256: "14e3dc4749bf42df13a67a271065b0f334d0ad36bb34a74cc57c6e137f9af09e",
			Md5:    "73c046196302ff7218d47046cf3c0501",
		},
		{
			Path: RtRepo1 + "/data/testdata/a/b",
			Type: "folder",
			Size: 0,
		},
		{
			Path:   RtRepo1 + "/data/testdata/a/b/b1.in",
			Type:   "file",
			Size:   9,
			Sha1:   "954cf8f3f75c41f18540bb38460910b4f0074e6f",
			Sha256: "b06c458f2aa21bd89e75e365bf5a9227c8b8d2b0728b1116d6738d214113def2",
			Md5:    "4f5561d29422374e40bd97d28b12cf35",
		},
		{
			Path:   RtRepo1 + "/data/testdata/a/b/b2.in",
			Type:   "file",
			Size:   9,
			Sha1:   "3b60b837e037568856bedc1dd4952d17b3f06972",
			Sha256: "2f96053cc48504bca84360967659abc3c145a56a530f7679bc75d9b1a66182b2",
			Md5:    "6931271be1e5f98e36bdc7a05097407b",
		},
		{
			Path:   RtRepo1 + "/data/testdata/a/b/b3.in",
			Type:   "file",
			Size:   9,
			Sha1:   "ec6420d2b5f708283619b25e68f9ddd351f555fe",
			Sha256: "b04c0632c0f647ce07741c8a6cb5ad1d2da6e8047d0127bd4171e1201cf9bf7e",
			Md5:    "305b21db102cf3a3d2d8c3f7e9584dba",
		},
		{
			Path: RtRepo1 + "/data/testdata/a/b/c",
			Type: "folder",
			Size: 0,
		},
		{
			Path:   RtRepo1 + "/data/testdata/a/b/c/c1.in",
			Type:   "file",
			Size:   11,
			Sha1:   "063041114949bf19f6fe7508aef639640e7edaac",
			Sha256: "8b511ab4559d91c559e033d60888da1409b617db21491355386242577d651af4",
			Md5:    "e53098d3d8ee1f5eb38c2ec3c783ef3d",
		},
		{
			Path:   RtRepo1 + "/data/testdata/a/b/c/c2.in",
			Type:   "file",
			Size:   11,
			Sha1:   "a4f912be11e7d1d346e34c300e6d4b90e136896e",
			Sha256: "822505a4aa0e0ada22ae6b9a23ae88b46f718df7ca64fee629739396fdefb846",
			Md5:    "82b6d565393a3fd1cc4778b1d53c0664",
		},
		{
			Path:   RtRepo1 + "/data/testdata/a/b/c/c3.in",
			Type:   "file",
			Size:   11,
			Sha1:   "2d6ee506188db9b816a6bfb79c5df562fc1d8658",
			Sha256: "69efd5b0596c22cd7629a96ee4fe061b3020cd9078e9e192235053ab1cbdc35d",
			Md5:    "d8020b86244956f647cf1beff5acdb90",
		},
	}
}

func GetSearchNotIncludeDirsFiles() []utils.SearchResult {
	return []utils.SearchResult{
		{
			Path:   RtRepo1 + "/data/testdata/a/a1.in",
			Type:   "file",
			Size:   7,
			Sha1:   "507ac63c6b0f650fb6f36b5621e70ebca3b0965c",
			Sha256: "4eb341b5d2762a853d79cc25e622aa8b978eb6e12c3259e2d99dc9dc60d82c5d",
			Md5:    "65298e78fe5883eee82056bc6d0d7f4c",
		},
		{
			Path:   RtRepo1 + "/data/testdata/a/a2.in",
			Type:   "file",
			Size:   7,
			Sha1:   "de2f31d77e2c2b1039a806f21b0c5f3243e45588",
			Sha256: "3e3deb6628658a48cf0d280a2210211f9d977ec2e10a4619b95d5fb85cb10450",
			Md5:    "28f9732cb82a2d11760e38614246ad6d",
		},
		{
			Path:   RtRepo1 + "/data/testdata/a/a3.in",
			Type:   "file",
			Size:   7,
			Sha1:   "29d38faccfe74dee60d0142a716e8ea6fad67b49",
			Sha256: "14e3dc4749bf42df13a67a271065b0f334d0ad36bb34a74cc57c6e137f9af09e",
			Md5:    "73c046196302ff7218d47046cf3c0501",
		},
		{
			Path:   RtRepo1 + "/data/testdata/a/b/b1.in",
			Type:   "file",
			Size:   9,
			Sha1:   "954cf8f3f75c41f18540bb38460910b4f0074e6f",
			Sha256: "b06c458f2aa21bd89e75e365bf5a9227c8b8d2b0728b1116d6738d214113def2",
			Md5:    "4f5561d29422374e40bd97d28b12cf35",
		},
		{
			Path:   RtRepo1 + "/data/testdata/a/b/b2.in",
			Type:   "file",
			Size:   9,
			Sha1:   "3b60b837e037568856bedc1dd4952d17b3f06972",
			Sha256: "2f96053cc48504bca84360967659abc3c145a56a530f7679bc75d9b1a66182b2",
			Md5:    "6931271be1e5f98e36bdc7a05097407b",
		},
		{
			Path:   RtRepo1 + "/data/testdata/a/b/b3.in",
			Type:   "file",
			Size:   9,
			Sha1:   "ec6420d2b5f708283619b25e68f9ddd351f555fe",
			Sha256: "b04c0632c0f647ce07741c8a6cb5ad1d2da6e8047d0127bd4171e1201cf9bf7e",
			Md5:    "305b21db102cf3a3d2d8c3f7e9584dba",
		},
		{
			Path:   RtRepo1 + "/data/testdata/a/b/c/c1.in",
			Type:   "file",
			Size:   11,
			Sha1:   "063041114949bf19f6fe7508aef639640e7edaac",
			Sha256: "8b511ab4559d91c559e033d60888da1409b617db21491355386242577d651af4",
			Md5:    "e53098d3d8ee1f5eb38c2ec3c783ef3d",
		},
		{
			Path:   RtRepo1 + "/data/testdata/a/b/c/c2.in",
			Type:   "file",
			Size:   11,
			Sha1:   "a4f912be11e7d1d346e34c300e6d4b90e136896e",
			Sha256: "822505a4aa0e0ada22ae6b9a23ae88b46f718df7ca64fee629739396fdefb846",
			Md5:    "82b6d565393a3fd1cc4778b1d53c0664",
		},
		{
			Path:   RtRepo1 + "/data/testdata/a/b/c/c3.in",
			Type:   "file",
			Size:   11,
			Sha1:   "2d6ee506188db9b816a6bfb79c5df562fc1d8658",
			Sha256: "69efd5b0596c22cd7629a96ee4fe061b3020cd9078e9e192235053ab1cbdc35d",
			Md5:    "d8020b86244956f647cf1beff5acdb90",
		},
	}
}

func GetSearchAfterDeleteWithExcludeProps() []utils.SearchResult {
	return []utils.SearchResult{
		{
			Path:   RtRepo1 + "/a/c1.in",
			Type:   "file",
			Size:   11,
			Sha1:   "063041114949bf19f6fe7508aef639640e7edaac",
			Sha256: "8b511ab4559d91c559e033d60888da1409b617db21491355386242577d651af4",
			Md5:    "e53098d3d8ee1f5eb38c2ec3c783ef3d",
			Props: map[string][]string{
				"c": {"1"},
			},
		},
		{
			Path:   RtRepo1 + "/e/c1.in",
			Type:   "file",
			Size:   11,
			Sha1:   "063041114949bf19f6fe7508aef639640e7edaac",
			Sha256: "8b511ab4559d91c559e033d60888da1409b617db21491355386242577d651af4",
			Md5:    "e53098d3d8ee1f5eb38c2ec3c783ef3d",
			Props: map[string][]string{
				"c": {"1"},
			},
		},
	}
}

func GetThirdSearchResultSortedByAsc() utils.SearchResult {
	return utils.SearchResult{
		Path:   RtRepo1 + "/org",
		Type:   "file",
		Sha1:   "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		Sha256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		Md5:    "d41d8cd98f00b204e9800998ecf8427e",
	}

}

func GetSecondSearchResultSortedByAsc() utils.SearchResult {
	return utils.SearchResult{
		Path:   RtRepo1 + "/o",
		Type:   "file",
		Sha1:   "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		Sha256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		Md5:    "d41d8cd98f00b204e9800998ecf8427e",
	}
}

func GetFirstSearchResultSortedByAsc() utils.SearchResult {
	return utils.SearchResult{
		Path:   RtRepo1 + "/or",
		Type:   "file",
		Sha1:   "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		Sha256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		Md5:    "d41d8cd98f00b204e9800998ecf8427e",
		Props: map[string][]string{
			"k1": {"v1"},
		},
	}
}

func GetSearchPropsStep1() []utils.SearchResult {
	return []utils.SearchResult{
		{
			Path:   RtRepo1 + "/a/a3.in",
			Type:   "file",
			Size:   7,
			Sha1:   "29d38faccfe74dee60d0142a716e8ea6fad67b49",
			Sha256: "14e3dc4749bf42df13a67a271065b0f334d0ad36bb34a74cc57c6e137f9af09e",
			Md5:    "73c046196302ff7218d47046cf3c0501",
			Props: map[string][]string{
				"a": {"1"},
				"b": {"3"},
				"c": {"3"},
			},
		},
		{
			Path:   RtRepo1 + "/a/b/b2.in",
			Type:   "file",
			Size:   9,
			Sha1:   "3b60b837e037568856bedc1dd4952d17b3f06972",
			Sha256: "2f96053cc48504bca84360967659abc3c145a56a530f7679bc75d9b1a66182b2",
			Md5:    "6931271be1e5f98e36bdc7a05097407b",
			Props: map[string][]string{
				"b": {"1"},
				"c": {"3"},
			},
		},
		{
			Path:   RtRepo1 + "/a/b/b3.in",
			Type:   "file",
			Size:   9,
			Sha1:   "ec6420d2b5f708283619b25e68f9ddd351f555fe",
			Sha256: "b04c0632c0f647ce07741c8a6cb5ad1d2da6e8047d0127bd4171e1201cf9bf7e",
			Md5:    "305b21db102cf3a3d2d8c3f7e9584dba",
			Props: map[string][]string{
				"a": {"1"},
				"b": {"2"},
				"c": {"3"},
			},
		},
		{
			Path:   RtRepo1 + "/a/b/c/c2.in",
			Type:   "file",
			Size:   11,
			Sha1:   "a4f912be11e7d1d346e34c300e6d4b90e136896e",
			Sha256: "822505a4aa0e0ada22ae6b9a23ae88b46f718df7ca64fee629739396fdefb846",
			Md5:    "82b6d565393a3fd1cc4778b1d53c0664",
			Props: map[string][]string{
				"c": {"3"},
			},
		},
		{
			Path:   RtRepo1 + "/a/b/c/c3.in",
			Type:   "file",
			Size:   11,
			Sha1:   "2d6ee506188db9b816a6bfb79c5df562fc1d8658",
			Sha256: "69efd5b0596c22cd7629a96ee4fe061b3020cd9078e9e192235053ab1cbdc35d",
			Md5:    "d8020b86244956f647cf1beff5acdb90",
			Props: map[string][]string{
				"c": {"3"},
			},
		},
	}
}

func GetSearchPropsStep2() []utils.SearchResult {
	return []utils.SearchResult{
		{
			Path:   RtRepo1 + "/a/a1.in",
			Type:   "file",
			Size:   7,
			Sha1:   "507ac63c6b0f650fb6f36b5621e70ebca3b0965c",
			Sha256: "4eb341b5d2762a853d79cc25e622aa8b978eb6e12c3259e2d99dc9dc60d82c5d",
			Md5:    "65298e78fe5883eee82056bc6d0d7f4c",
			Props: map[string][]string{
				"a": {"2"},
				"b": {"3"},
			},
		},
		{
			Path:   RtRepo1 + "/a/a2.in",
			Type:   "file",
			Size:   7,
			Sha1:   "de2f31d77e2c2b1039a806f21b0c5f3243e45588",
			Sha256: "3e3deb6628658a48cf0d280a2210211f9d977ec2e10a4619b95d5fb85cb10450",
			Md5:    "28f9732cb82a2d11760e38614246ad6d",
			Props: map[string][]string{
				"a": {"1"},
			},
		},
		{
			Path:   RtRepo1 + "/a/b/b1.in",
			Type:   "file",
			Size:   9,
			Sha1:   "954cf8f3f75c41f18540bb38460910b4f0074e6f",
			Sha256: "b06c458f2aa21bd89e75e365bf5a9227c8b8d2b0728b1116d6738d214113def2",
			Md5:    "4f5561d29422374e40bd97d28b12cf35",
			Props: map[string][]string{
				"a": {"1"},
				"c": {"5"},
			},
		},
		{
			Path:   RtRepo1 + "/a/b/c/c1.in",
			Type:   "file",
			Size:   11,
			Sha1:   "063041114949bf19f6fe7508aef639640e7edaac",
			Sha256: "8b511ab4559d91c559e033d60888da1409b617db21491355386242577d651af4",
			Md5:    "e53098d3d8ee1f5eb38c2ec3c783ef3d",
			Props: map[string][]string{
				"b": {"1"},
			},
		},
	}
}

func GetSearchPropsStep3() []utils.SearchResult {
	return []utils.SearchResult{
		{
			Path:   RtRepo1 + "/a/a1.in",
			Type:   "file",
			Size:   7,
			Sha1:   "507ac63c6b0f650fb6f36b5621e70ebca3b0965c",
			Sha256: "4eb341b5d2762a853d79cc25e622aa8b978eb6e12c3259e2d99dc9dc60d82c5d",
			Md5:    "65298e78fe5883eee82056bc6d0d7f4c",
			Props: map[string][]string{
				"a": {"2"},
				"b": {"3"},
			},
		},
		{
			Path:   RtRepo1 + "/a/a2.in",
			Type:   "file",
			Size:   7,
			Sha1:   "de2f31d77e2c2b1039a806f21b0c5f3243e45588",
			Sha256: "3e3deb6628658a48cf0d280a2210211f9d977ec2e10a4619b95d5fb85cb10450",
			Md5:    "28f9732cb82a2d11760e38614246ad6d",
			Props: map[string][]string{
				"a": {"1"},
			},
		},
		{
			Path:   RtRepo1 + "/a/a3.in",
			Type:   "file",
			Size:   7,
			Sha1:   "29d38faccfe74dee60d0142a716e8ea6fad67b49",
			Sha256: "14e3dc4749bf42df13a67a271065b0f334d0ad36bb34a74cc57c6e137f9af09e",
			Md5:    "73c046196302ff7218d47046cf3c0501",
			Props: map[string][]string{
				"a": {"1"},
				"b": {"3"},
				"c": {"3"},
			},
		},
		{
			Path:   RtRepo1 + "/a/b/b1.in",
			Type:   "file",
			Size:   9,
			Sha1:   "954cf8f3f75c41f18540bb38460910b4f0074e6f",
			Sha256: "b06c458f2aa21bd89e75e365bf5a9227c8b8d2b0728b1116d6738d214113def2",
			Md5:    "4f5561d29422374e40bd97d28b12cf35",
			Props: map[string][]string{
				"a": {"1"},
				"c": {"5"},
			},
		},
		{
			Path:   RtRepo1 + "/a/b/b2.in",
			Type:   "file",
			Size:   9,
			Sha1:   "3b60b837e037568856bedc1dd4952d17b3f06972",
			Sha256: "2f96053cc48504bca84360967659abc3c145a56a530f7679bc75d9b1a66182b2",
			Md5:    "6931271be1e5f98e36bdc7a05097407b",
			Props: map[string][]string{
				"b": {"1"},
				"c": {"3"},
			},
		},
		{
			Path:   RtRepo1 + "/a/b/c/c1.in",
			Type:   "file",
			Size:   11,
			Sha1:   "063041114949bf19f6fe7508aef639640e7edaac",
			Sha256: "8b511ab4559d91c559e033d60888da1409b617db21491355386242577d651af4",
			Md5:    "e53098d3d8ee1f5eb38c2ec3c783ef3d",
			Props: map[string][]string{
				"b": {"1"},
			},
		},
		{
			Path:   RtRepo1 + "/a/b/c/c2.in",
			Type:   "file",
			Size:   11,
			Sha1:   "a4f912be11e7d1d346e34c300e6d4b90e136896e",
			Sha256: "822505a4aa0e0ada22ae6b9a23ae88b46f718df7ca64fee629739396fdefb846",
			Md5:    "82b6d565393a3fd1cc4778b1d53c0664",
			Props: map[string][]string{
				"c": {"3"},
			},
		},
		{
			Path:   RtRepo1 + "/a/b/c/c3.in",
			Type:   "file",
			Size:   11,
			Sha1:   "2d6ee506188db9b816a6bfb79c5df562fc1d8658",
			Sha256: "69efd5b0596c22cd7629a96ee4fe061b3020cd9078e9e192235053ab1cbdc35d",
			Md5:    "d8020b86244956f647cf1beff5acdb90",
			Props: map[string][]string{
				"c": {"3"},
			},
		},
	}
}

func GetSearchPropsStep4() []utils.SearchResult {
	return []utils.SearchResult{
		{
			Path:   RtRepo1 + "/a/a3.in",
			Type:   "file",
			Size:   7,
			Sha1:   "29d38faccfe74dee60d0142a716e8ea6fad67b49",
			Sha256: "14e3dc4749bf42df13a67a271065b0f334d0ad36bb34a74cc57c6e137f9af09e",
			Md5:    "73c046196302ff7218d47046cf3c0501",
			Props: map[string][]string{
				"a": {"1"},
				"b": {"3"},
				"c": {"3"},
			},
		},
		{
			Path:   RtRepo1 + "/a/b/b2.in",
			Type:   "file",
			Size:   9,
			Sha1:   "3b60b837e037568856bedc1dd4952d17b3f06972",
			Sha256: "2f96053cc48504bca84360967659abc3c145a56a530f7679bc75d9b1a66182b2",
			Md5:    "6931271be1e5f98e36bdc7a05097407b",
			Props: map[string][]string{
				"b": {"1"},
				"c": {"3"},
			},
		},
		{
			Path:   RtRepo1 + "/a/b/c/c2.in",
			Type:   "file",
			Size:   11,
			Sha1:   "a4f912be11e7d1d346e34c300e6d4b90e136896e",
			Sha256: "822505a4aa0e0ada22ae6b9a23ae88b46f718df7ca64fee629739396fdefb846",
			Md5:    "82b6d565393a3fd1cc4778b1d53c0664",
			Props: map[string][]string{
				"c": {"3"},
			},
		},
		{
			Path:   RtRepo1 + "/a/b/c/c3.in",
			Type:   "file",
			Size:   11,
			Sha1:   "2d6ee506188db9b816a6bfb79c5df562fc1d8658",
			Sha256: "69efd5b0596c22cd7629a96ee4fe061b3020cd9078e9e192235053ab1cbdc35d",
			Md5:    "d8020b86244956f647cf1beff5acdb90",
			Props: map[string][]string{
				"c": {"3"},
			},
		},
	}
}

func GetSearchPropsStep5() []utils.SearchResult {
	return make([]utils.SearchResult, 0)
}

func GetSearchPropsStep6() []utils.SearchResult {
	return []utils.SearchResult{
		{
			Path:   RtRepo1 + "/a/b/c/c1.in",
			Type:   "file",
			Size:   11,
			Sha1:   "063041114949bf19f6fe7508aef639640e7edaac",
			Sha256: "8b511ab4559d91c559e033d60888da1409b617db21491355386242577d651af4",
			Md5:    "e53098d3d8ee1f5eb38c2ec3c783ef3d",
			Props: map[string][]string{
				"b": {"1"},
			},
		},
	}
}

func GetSearchResultAfterDeleteByPropsStep1() []utils.SearchResult {
	return []utils.SearchResult{
		{
			Path:   RtRepo1 + "/a/a1.in",
			Type:   "file",
			Size:   7,
			Sha1:   "507ac63c6b0f650fb6f36b5621e70ebca3b0965c",
			Sha256: "4eb341b5d2762a853d79cc25e622aa8b978eb6e12c3259e2d99dc9dc60d82c5d",
			Md5:    "65298e78fe5883eee82056bc6d0d7f4c",
			Props: map[string][]string{
				"a": {"2"},
				"b": {"3"},
			},
		},
		{
			Path:   RtRepo1 + "/a/a2.in",
			Type:   "file",
			Size:   7,
			Sha1:   "de2f31d77e2c2b1039a806f21b0c5f3243e45588",
			Sha256: "3e3deb6628658a48cf0d280a2210211f9d977ec2e10a4619b95d5fb85cb10450",
			Md5:    "28f9732cb82a2d11760e38614246ad6d",
			Props: map[string][]string{
				"a": {"1"},
			},
		},
		{
			Path:   RtRepo1 + "/a/a3.in",
			Type:   "file",
			Size:   7,
			Sha1:   "29d38faccfe74dee60d0142a716e8ea6fad67b49",
			Sha256: "14e3dc4749bf42df13a67a271065b0f334d0ad36bb34a74cc57c6e137f9af09e",
			Md5:    "73c046196302ff7218d47046cf3c0501",
			Props: map[string][]string{
				"a": {"1"},
				"b": {"3"},
				"c": {"3"},
			},
		},
		{
			Path:   RtRepo1 + "/a/b/b2.in",
			Type:   "file",
			Size:   9,
			Sha1:   "3b60b837e037568856bedc1dd4952d17b3f06972",
			Sha256: "2f96053cc48504bca84360967659abc3c145a56a530f7679bc75d9b1a66182b2",
			Md5:    "6931271be1e5f98e36bdc7a05097407b",
			Props: map[string][]string{
				"b": {"1"},
				"c": {"3"},
				"D": {"5"},
			},
		},
		{
			Path:   RtRepo1 + "/a/b/b3.in",
			Type:   "file",
			Size:   9,
			Sha1:   "ec6420d2b5f708283619b25e68f9ddd351f555fe",
			Sha256: "b04c0632c0f647ce07741c8a6cb5ad1d2da6e8047d0127bd4171e1201cf9bf7e",
			Md5:    "305b21db102cf3a3d2d8c3f7e9584dba",
			Props: map[string][]string{
				"a": {"1"},
				"b": {"2"},
				"c": {"3"},
				"D": {"5"},
			},
		},
		{
			Path:   RtRepo1 + "/a/b/c/c2.in",
			Type:   "file",
			Size:   11,
			Sha1:   "a4f912be11e7d1d346e34c300e6d4b90e136896e",
			Sha256: "822505a4aa0e0ada22ae6b9a23ae88b46f718df7ca64fee629739396fdefb846",
			Md5:    "82b6d565393a3fd1cc4778b1d53c0664",
			Props: map[string][]string{
				"c": {"3"},
				"D": {"2"},
			},
		},
		{
			Path:   RtRepo1 + "/a/b/c/c3.in",
			Type:   "file",
			Size:   11,
			Sha1:   "2d6ee506188db9b816a6bfb79c5df562fc1d8658",
			Sha256: "69efd5b0596c22cd7629a96ee4fe061b3020cd9078e9e192235053ab1cbdc35d",
			Md5:    "d8020b86244956f647cf1beff5acdb90",
			Props: map[string][]string{
				"c": {"3"},
				"D": {"2"},
			},
		},
	}
}

func GetSearchResultAfterDeleteByPropsStep2() []utils.SearchResult {
	return []utils.SearchResult{
		{
			Path:   RtRepo1 + "/a/a1.in",
			Type:   "file",
			Size:   7,
			Sha1:   "507ac63c6b0f650fb6f36b5621e70ebca3b0965c",
			Sha256: "4eb341b5d2762a853d79cc25e622aa8b978eb6e12c3259e2d99dc9dc60d82c5d",
			Md5:    "65298e78fe5883eee82056bc6d0d7f4c",
			Props: map[string][]string{
				"a": {"2"},
				"b": {"3"},
			},
		},
		{
			Path:   RtRepo1 + "/a/a2.in",
			Type:   "file",
			Size:   7,
			Sha1:   "de2f31d77e2c2b1039a806f21b0c5f3243e45588",
			Sha256: "3e3deb6628658a48cf0d280a2210211f9d977ec2e10a4619b95d5fb85cb10450",
			Md5:    "28f9732cb82a2d11760e38614246ad6d",
			Props: map[string][]string{
				"a": {"1"},
			},
		},
		{
			Path:   RtRepo1 + "/a/a3.in",
			Type:   "file",
			Size:   7,
			Sha1:   "29d38faccfe74dee60d0142a716e8ea6fad67b49",
			Sha256: "14e3dc4749bf42df13a67a271065b0f334d0ad36bb34a74cc57c6e137f9af09e",
			Md5:    "73c046196302ff7218d47046cf3c0501",
			Props: map[string][]string{
				"a": {"1"},
				"b": {"3"},
				"c": {"3"},
			},
		},
		{
			Path:   RtRepo1 + "/a/b/b3.in",
			Type:   "file",
			Size:   9,
			Sha1:   "ec6420d2b5f708283619b25e68f9ddd351f555fe",
			Sha256: "b04c0632c0f647ce07741c8a6cb5ad1d2da6e8047d0127bd4171e1201cf9bf7e",
			Md5:    "305b21db102cf3a3d2d8c3f7e9584dba",
			Props: map[string][]string{
				"a": {"1"},
				"b": {"2"},
				"c": {"3"},
				"D": {"5"},
			},
		},
	}
}

func GetSearchResultAfterDeleteByPropsStep3() []utils.SearchResult {
	return []utils.SearchResult{
		{
			Path:   RtRepo1 + "/a/a1.in",
			Type:   "file",
			Size:   7,
			Sha1:   "507ac63c6b0f650fb6f36b5621e70ebca3b0965c",
			Sha256: "4eb341b5d2762a853d79cc25e622aa8b978eb6e12c3259e2d99dc9dc60d82c5d",
			Md5:    "65298e78fe5883eee82056bc6d0d7f4c",
			Props: map[string][]string{
				"a": {"2"},
				"b": {"3"},
			},
		},
		{
			Path:   RtRepo1 + "/a/a3.in",
			Type:   "file",
			Size:   7,
			Sha1:   "29d38faccfe74dee60d0142a716e8ea6fad67b49",
			Sha256: "14e3dc4749bf42df13a67a271065b0f334d0ad36bb34a74cc57c6e137f9af09e",
			Md5:    "73c046196302ff7218d47046cf3c0501",
			Props: map[string][]string{
				"a": {"1"},
				"b": {"3"},
				"c": {"3"},
			},
		},
	}
}

func GetDockerDeployedManifest() []string {
	return []string{
		DockerLocalPromoteRepo + "/" + DockerImageName + "promotion" + "/2/manifest.json",
	}
}

func GetMavenDeployedArtifacts() []string {
	return []string{
		MvnRepo1 + "/org/jfrog/cli-test/1.0/cli-test-1.0.jar",
		MvnRepo1 + "/org/jfrog/cli-test/1.0/cli-test-1.0.pom",
	}
}

func GetMavenMultiIncludedDeployedArtifacts() []string {
	return []string{
		MvnRepo1 + "/org/jfrog/test/multi1/3.7-SNAPSHOT/multi1-3.7-SNAPSHOT-tests.jar",
		MvnRepo1 + "/org/jfrog/test/multi/3.7-SNAPSHOT/multi-3.7-SNAPSHOT.pom",
		MvnRepo1 + "/org/jfrog/test/multi3/3.7-SNAPSHOT/multi3-3.7-SNAPSHOT.pom",
		MvnRepo1 + "/org/jfrog/test/multi1/3.7-SNAPSHOT/multi1-3.7-SNAPSHOT.pom",
	}
}

func GetGradleDeployedArtifacts() []string {
	return []string{
		GradleRepo + "/minimal-example/1.0/minimal-example-1.0.jar",
	}
}

func GetNpmDeployedScopedArtifacts(isNpm7 bool) []string {
	path := NpmRepo + "/@jscope/jfrog-cli-tests/-/@jscope/"
	path += GetNpmArtifactName(isNpm7, true)
	return []string{
		path,
	}
}

func GetNpmDeployedArtifacts(isNpm7 bool) []string {
	path := NpmRepo + "/jfrog-cli-tests/-/"
	path += GetNpmArtifactName(isNpm7, false)
	return []string{
		path,
	}
}

func GetNpmArtifactName(isNpm7, isScoped bool) string {
	if isNpm7 {
		if isScoped {
			return "jfrog-cli-tests-=1.0.0.tgz"
		} else {
			return "jfrog-cli-tests-v1.0.0.tgz"
		}
	}
	return "jfrog-cli-tests-1.0.0.tgz"
}

func GetSortAndLimit() []string {
	return []string{
		Out,
		filepath.Join(Out, "download"),
		filepath.Join(Out, "download", "sort_limit"),
		filepath.Join(Out, "download", "sort_limit", "data"),
		filepath.Join(Out, "download", "sort_limit", "data", "a1.in"),
		filepath.Join(Out, "download", "sort_limit", "data", "b"),
		filepath.Join(Out, "download", "sort_limit", "data", "b", "c"),
		filepath.Join(Out, "download", "sort_limit", "data", "b", "c", "c1.in"),
		filepath.Join(Out, "download", "sort_limit", "data", "b", "c", "c2.in"),
		filepath.Join(Out, "download", "sort_limit", "data", "b", "c", "c3.in"),
	}
}

func GetBuildDownloadByShaAndBuildNameWithSort() []string {
	return []string{
		filepath.Join(Out, "download", "sort_limit_by_build"),
		filepath.Join(Out, "download", "sort_limit_by_build", "data"),
		filepath.Join(Out, "download", "sort_limit_by_build", "data", "a11.in"),
	}
}

func GetBuildArchiveEntriesDownloadCli() []string {
	return []string{
		Out,
		filepath.Join(Out, "a.zip"),
		filepath.Join(Out, "b.zip"),
		filepath.Join(Out, "c.zip"),
	}
}

func GetBuildArchiveEntriesSpecificPathDownload() []string {
	return []string{
		Out,
		filepath.Join(Out, "b.zip"),
	}
}

func GetBuildArchiveEntriesDownloadSpec() []string {
	return []string{
		Out,
		filepath.Join(Out, "d.zip"),
	}
}

func GetWinCompatibility() []string {
	return []string{
		Out,
		filepath.Join(Out, "win"),
		filepath.Join(Out, "win", "a2.in"),
		filepath.Join(Out, "win", "b1.in"),
		filepath.Join(Out, "win", "b3.in"),
	}
}

func GetUploadExpectedRepo1SyncDeleteStep1() []string {
	return []string{
		RtRepo1 + "/syncDir/testdata/a/a3.in",
		RtRepo1 + "/syncDir/testdata/a/a1.in",
		RtRepo1 + "/syncDir/testdata/a/a2.in",
		RtRepo1 + "/syncDir/testdata/a/b/b1.in",
		RtRepo1 + "/syncDir/testdata/a/b/b2.in",
		RtRepo1 + "/syncDir/testdata/a/b/b3.in",
		RtRepo1 + "/syncDir/testdata/a/b/c/c1.in",
		RtRepo1 + "/syncDir/testdata/a/b/c/c2.in",
		RtRepo1 + "/syncDir/testdata/a/b/c/c3.in",
	}
}

func GetUploadExpectedRepo1SyncDeleteStep2() []string {
	return []string{
		RtRepo1 + "/syncDir/testdata/a/a3.in",
		RtRepo1 + "/syncDir/testdata/a/a1.in",
		RtRepo1 + "/syncDir/testdata/a/a2.in",
		RtRepo1 + "/syncDir/testdata/a/b/b1.in",
		RtRepo1 + "/syncDir/testdata/a/b/c/c1.in",
	}
}

func GetUploadExpectedRepo1SyncDeleteStep3() []string {
	return []string{
		RtRepo1 + "/syncDir/a.zip",
		RtRepo1 + "/syncDir/b.zip",
		RtRepo1 + "/syncDir/c.zip",
		RtRepo1 + "/syncDir/d.zip",
	}
}
func GetUploadExpectedRepo1SyncDeleteStep4() []string {
	return []string{
		RtRepo1 + "/syncDir/testdata/c/a/a.zip",
		RtRepo1 + "/syncDir/testdata/c/a/aaa.zip",
		RtRepo1 + "/syncDir/testdata/c/a-b/a.zip",
		RtRepo1 + "/syncDir/testdata/c/a-b/aaa.zip",
		RtRepo1 + "/syncDir/testdata/c/#a/a.zip",
	}
}

func GetUploadedFileWithDownloadedPlaceHolder() []string {
	return []string{
		RtRepo1 + "/path/b1.in",
		RtRepo1 + "/path/b2.in",
		RtRepo1 + "/path/b3.in",
		RtRepo1 + "/path/c/c1.in",
		RtRepo1 + "/path/c/c2.in",
		RtRepo1 + "/path/c/c3.in",
	}
}

func GetFileWithDownloadedPlaceHolder() []string {
	return []string{
		Out,
		filepath.Join(Out, "mypath2"),
		filepath.Join(Out, "mypath2", "b1.in"),
		filepath.Join(Out, "mypath2", "b2.in"),
		filepath.Join(Out, "mypath2", "b3.in"),
		filepath.Join(Out, "mypath2", "c"),
		filepath.Join(Out, "mypath2", "c", "c1.in"),
		filepath.Join(Out, "mypath2", "c", "c2.in"),
		filepath.Join(Out, "mypath2", "c", "c3.in"),
	}
}

func GetUploadedFileWithDownloadedDoublePlaceHolder() []string {
	return []string{
		RtRepo1 + "/path/c/c1.in",
		RtRepo1 + "/path/c/c2.in",
		RtRepo1 + "/path/c/c3.in",
	}
}

func GetFileWithDownloadedDoublePlaceHolder() []string {
	return []string{
		Out,
		filepath.Join(Out, "mypath2"),
		filepath.Join(Out, "mypath2", "c"),
		filepath.Join(Out, "mypath2", "c", "c1.in"),
		filepath.Join(Out, "mypath2", "c", "c2.in"),
		filepath.Join(Out, "mypath2", "c", "c3.in"),
	}
}

func GetUploadedFileWithDownloadedPlaceHolderlashSlashSuffix() []string {
	return []string{
		RtRepo1 + "/path/b1.in/b1.in",
		RtRepo1 + "/path/b2.in/b2.in",
		RtRepo1 + "/path/b3.in/b3.in",
		RtRepo1 + "/path/c/c1.in/c1.in",
		RtRepo1 + "/path/c/c2.in/c2.in",
		RtRepo1 + "/path/c/c3.in/c3.in",
	}
}

func GetFileWithDownloadedPlaceHolderSlashSuffix() []string {
	return []string{
		Out,
		filepath.Join(Out, "mypath2"),
		filepath.Join(Out, "mypath2", "b1.in"),
		filepath.Join(Out, "mypath2", "b2.in"),
		filepath.Join(Out, "mypath2", "b3.in"),
		filepath.Join(Out, "mypath2", "b1.in", "b1.in"),
		filepath.Join(Out, "mypath2", "b2.in", "b2.in"),
		filepath.Join(Out, "mypath2", "b3.in", "b3.in"),
		filepath.Join(Out, "mypath2", "c"),
		filepath.Join(Out, "mypath2", "c", "c1.in"),
		filepath.Join(Out, "mypath2", "c", "c1.in"),
		filepath.Join(Out, "mypath2", "c", "c2.in"),
		filepath.Join(Out, "mypath2", "c", "c1.in", "c1.in"),
		filepath.Join(Out, "mypath2", "c", "c2.in", "c2.in"),
		filepath.Join(Out, "mypath2", "c", "c3.in", "c3.in"),
	}
}

func GetExpectedUploadSummaryDetails(rtUrl string) []clientUtils.FileTransferDetails {
	path1, path2, path3 := filepath.Join("testdata", "a", "a1.in"), filepath.Join("testdata", "a", "a2.in"), filepath.Join("testdata", "a", "a3.in")
	return []clientUtils.FileTransferDetails{
		{SourcePath: path1, RtUrl: rtUrl, TargetPath: RtRepo1 + "/testdata/a/a1.in", Sha256: "4eb341b5d2762a853d79cc25e622aa8b978eb6e12c3259e2d99dc9dc60d82c5d"},
		{SourcePath: path2, RtUrl: rtUrl, TargetPath: RtRepo1 + "/testdata/a/a2.in", Sha256: "3e3deb6628658a48cf0d280a2210211f9d977ec2e10a4619b95d5fb85cb10450"},
		{SourcePath: path3, RtUrl: rtUrl, TargetPath: RtRepo1 + "/testdata/a/a3.in", Sha256: "14e3dc4749bf42df13a67a271065b0f334d0ad36bb34a74cc57c6e137f9af09e"},
	}
}

func GetReplicationConfig() []servicesUtils.ReplicationParams {
	return []servicesUtils.ReplicationParams{
		{
			Url:                      *JfrogUrl + ArtifactoryEndpoint + "targetRepo",
			Username:                 "admin",
			CronExp:                  "0 0 12 * * ?",
			RepoKey:                  RtRepo1,
			EnableEventReplication:   false,
			SocketTimeoutMillis:      15000,
			Enabled:                  true,
			SyncDeletes:              true,
			SyncProperties:           true,
			PathPrefix:               "",
			IncludePathPrefixPattern: "/my/path",
		},
	}
}

func GetExpectedPermissionTarget(repoValue string) services.PermissionTargetParams {
	return services.PermissionTargetParams{
		Name: RtPermissionTargetName,
		Repo: &services.PermissionTargetSection{
			Repositories:    []string{repoValue},
			IncludePatterns: []string{"**"},
			ExcludePatterns: []string{},
			Actions: &services.Actions{
				Groups: map[string][]string{
					"readers": {"read"},
				},
			},
		},
	}
}

func GetTerraformModulesFilesDownload() []string {
	namespace := "namespace"
	provider := "provider"
	return []string{
		filepath.Join(Out, "results"),
		filepath.Join(Out, "results", namespace),
		filepath.Join(Out, "results", namespace, "asg"),
		filepath.Join(Out, "results", namespace, "asg", provider),
		filepath.Join(Out, "results", namespace, "asg", provider, "module.json"),
		filepath.Join(Out, "results", namespace, "asg", provider, "module.tf"),
		filepath.Join(Out, "results", namespace, "asg", provider, "files"),
		filepath.Join(Out, "results", namespace, "asg", provider, "files", "f.sh"),
		filepath.Join(Out, "results", namespace, "asg", provider, "submodule"),
		filepath.Join(Out, "results", namespace, "asg", provider, "submodule", "module.tf"),
		filepath.Join(Out, "results", namespace, "awsmodule"),
		filepath.Join(Out, "results", namespace, "awsmodule", provider),
		filepath.Join(Out, "results", namespace, "awsmodule", provider, "module.json"),
		filepath.Join(Out, "results", namespace, "awsmodule", provider, "main.tf"),
		filepath.Join(Out, "results", namespace, "byok"),
		filepath.Join(Out, "results", namespace, "byok", provider),
		filepath.Join(Out, "results", namespace, "byok", provider, "module.json"),
		filepath.Join(Out, "results", namespace, "byok", provider, "module.tf"),
	}
}

func GetTransferExpectedRepo1() []string {
	return []string{
		RtRepo1 + "/testdata/a/a1.in",
		RtRepo1 + "/testdata/a/a2.in",
		RtRepo1 + "/testdata/a/a3.in",
		RtRepo1 + "/testdata/a/b/b1.in",
		RtRepo1 + "/testdata/a/b/b2.in",
		RtRepo1 + "/testdata/a/b/b3.in",
		RtRepo1 + "/testdata/a/b/c/c1.in",
		RtRepo1 + "/testdata/a/b/c/c2.in",
		RtRepo1 + "/testdata/a/b/c/c3.in",
	}
}

func GetTransferExpectedRepo2() []string {
	return []string{
		RtRepo2 + "/testdata/a/b/b1.in",
		RtRepo2 + "/testdata/a/b/b2.in",
		RtRepo2 + "/testdata/a/b/b3.in",
		RtRepo2 + "/testdata/a/b/c/c1.in",
		RtRepo2 + "/testdata/a/b/c/c2.in",
		RtRepo2 + "/testdata/a/b/c/c3.in",
	}
}

func GetTransferExpectedRepoSnapshot() []string {
	return []string{
		RtRepo1 + "/testdata/a/a1.in",
		RtRepo1 + "/testdata/a/a2.in",
		RtRepo1 + "/testdata/a/a3.in",
		RtRepo1 + "/testdata/a/b/b1.in",
		RtRepo1 + "/testdata/a/b/b2.in",
		RtRepo1 + "/testdata/a/b/b3.in",
	}
}

func GetExpectedBackwardCompatibleLifecycleArtifacts() []string {
	return []string{
		RtDevRepo + "/a1.in",
		RtDevRepo + "/a2.in",
		RtDevRepo + "/a3.in",
		RtDevRepo + "/b1.in",
		RtDevRepo + "/b2.in",
		RtDevRepo + "/b3.in",
		RtDevRepo + "/dep-file",
	}
}

func GetExpectedLifecycleCreationByArtifacts() []string {
	return []string{
		RtDevRepo + "/a2.in",
		RtDevRepo + "/b1.in",
		RtDevRepo + "/b2.in",
		RtDevRepo + "/b3.in",
	}
}

func GetExpectedLifecycleCreationByAql() []string {
	return []string{
		RtDevRepo + "/a2.in",
	}
}

func GetExpectedLifecycleArtifacts() []string {
	return []string{
		RtProdRepo1 + "/a1.in",
		RtProdRepo1 + "/a2.in",
		RtProdRepo1 + "/a3.in",
		RtProdRepo1 + "/b1.in",
		RtProdRepo1 + "/b2.in",
		RtProdRepo1 + "/b3.in",
		RtProdRepo1 + "/dep-file",
	}
}

func GetExpectedLifecycleDistributedArtifacts() []string {
	return []string{
		RtDevRepo + "/a1.in",
		RtDevRepo + "/a2.in",
		RtDevRepo + "/a3.in",
		RtDevRepo + "/b1.in",
		RtDevRepo + "/b2.in",
		RtDevRepo + "/b3.in",
		RtDevRepo + "/c1.in",
		RtDevRepo + "/c2.in",
		RtDevRepo + "/c3.in",
		RtDevRepo + "/target/a1.in",
		RtDevRepo + "/target/a2.in",
		RtDevRepo + "/target/a3.in",
		RtDevRepo + "/target/b1.in",
		RtDevRepo + "/target/b2.in",
		RtDevRepo + "/target/b3.in",
		RtDevRepo + "/target/c1.in",
		RtDevRepo + "/target/c2.in",
		RtDevRepo + "/target/c3.in",
	}
}

func GetGoPublishWithExclusionsExpectedRepoGo() []string {
	var expected = []string{
		GoRepo + "/github.com/jfrog/dependency/@v/v1.1.1.info",
		GoRepo + "/github.com/jfrog/dependency/@v/v1.1.1.mod",
		GoRepo + "/github.com/jfrog/dependency/@v/v1.1.1.zip",
	}
	return expected
}

func GetGoPublishWithExclusionsExpectedFiles1() []string {
	var expected = []string{
		GoPublishWithExclusionPath + "dir4/d.txt",
	}
	return expected
}

func GetGoPublishWithExclusionsExcludedFiles1() []string {
	var excluded = []string{
		GoPublishWithExclusionPath + "dir1/a.txt",
		GoPublishWithExclusionPath + "dir1/dir2/b.txt",
		GoPublishWithExclusionPath + "dir1/dir2/dir3/c.txt",
	}
	return excluded
}

func GetGoPublishWithExclusionsExpectedFiles2() []string {
	var expected = []string{
		GoPublishWithExclusionPath + "dir4/d.txt",
		GoPublishWithExclusionPath + "dir1/a.txt",
	}
	return expected
}

func GetGoPublishWithExclusionsExcludedFiles2() []string {
	var excluded = []string{
		GoPublishWithExclusionPath + "dir1/dir2/b.txt",
		GoPublishWithExclusionPath + "dir1/dir2/dir3/c.txt",
	}
	return excluded
}

func GetGoPublishWithExclusionsExcludedFiles3() []string {
	var excluded = []string{
		GoPublishWithExclusionPath + "dir1/a.txt",
		GoPublishWithExclusionPath + "dir1/dir2/b.txt",
		GoPublishWithExclusionPath + "dir1/dir2/dir3/c.txt",
		GoPublishWithExclusionPath + "dir4/d.txt",
	}
	return excluded
}
