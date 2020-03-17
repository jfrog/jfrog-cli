package tests

import (
	"path/filepath"

	"github.com/jfrog/jfrog-cli-go/artifactory/commands/generic"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

const (
	ArchiveEntriesDownload                 = "archive_entries_download_spec.json"
	ArchiveEntriesUpload                   = "archive_entries_upload_spec.json"
	BuildAddDepsBuildName                  = "cli-bad-test-build"
	BuildAddDepsDoubleSpec                 = "build_add_deps_double_spec.json"
	BuildAddDepsSpec                       = "build_add_deps_simple_spec.json"
	BuildDownloadSpec                      = "build_download_spec.json"
	BuildDownloadSpecNoBuildNumber         = "build_download_spec_no_build_number.json"
	BuildDownloadSpecNoBuildNumberWithSort = "build_download_spec_no_build_number_with_sort.json"
	BuildDownloadSpecNoPattern             = "build_download_spec_no_pattern.json"
	BundleDownloadSpec                     = "bundle_download_spec.json"
	BundleDownloadSpecNoPattern            = "bundle_download_spec_no_pattern.json"
	CopyByBuildPatternAllSpec              = "move_copy_delete_by_build_pattern_all_spec.json"
	CopyByBuildSpec                        = "move_copy_delete_by_build_spec.json"
	CopyByBundleSpec                       = "copy_by_bundle_spec.json"
	CopyItemsSpec                          = "copy_items_spec.json"
	CopyMoveSimpleSpec                     = "copy_move_simple.json"
	CpMvDlByBuildAssertSpec                = "copy_by_build_assert_spec.json"
	DebianTestRepositoryConfig             = "debian_test_repository_config.json"
	DebianUploadSpec                       = "upload_debian_spec.json"
	DeleteSimpleSpec                       = "delete_simple_spec.json"
	DeleteSpec                             = "delete_spec.json"
	DeleteSpecWildcardInRepo               = "delete_spec_wildcard.json"
	DelSpecExclude                         = "delete_spec_exclude.json"
	DelSpecExclusions                      = "delete_spec_exclusions.json"
	DistributionRules                      = "distribution_rules.json"
	DownloadAllRepo1TestResources          = "download_all_repo1_test_resources.json"
	DownloadEmptyDirs                      = "download_empty_dir_spec.json"
	DownloadModFileGo                      = "downloadmodfile_go.json"
	DownloadModOfDependencyGo              = "downloadmodofdependency_go.json"
	DownloadSpecExclude                    = "download_spec_exclude.json"
	DownloadSpecExclusions                 = "download_spec_exclusions.json"
	DownloadWildcardRepo                   = "download_wildcard_repo.json"
	GitLfsAssertSpec                       = "git_lfs_assert_spec.json"
	GitLfsTestRepositoryConfig             = "git_lfs_test_repository_config.json"
	GoLocalRepositoryConfig                = "go_local_repository_config.json"
	GradleConfig                           = "gradle.yaml"
	GradleServerIDConfig                   = "gradle_server_id.yaml"
	GradleServerIDUsesPluginConfig         = "gradle_server_id_uses_plugin.yaml"
	GradleUsernamePasswordTemplate         = "gradle_user_pass_template.yaml"
	HttpsProxyEnvVar                       = "PROXY_HTTPS_PORT"
	JcenterRemoteRepositoryConfig          = "jcenter_remote_repository_config.json"
	MavenConfig                            = "maven.yaml"
	MavenServerIDConfig                    = "maven_server_id.yaml"
	MavenUsernamePasswordTemplate          = "maven_user_pass_template.yaml"
	MoveCopySpecExclude                    = "move_copy_spec_exclude.json"
	MoveCopySpecExclusions                 = "move_copy_spec_exclusions.json"
	MoveRepositoryConfig                   = "move_repository_config.json"
	NpmBuildName                           = "cli-npm-test-build"
	NpmLocalRepositoryConfig               = "npm_local_repository_config.json"
	NpmRemoteRepositoryConfig              = "npm_remote_repository_config.json"
	NugetBuildName                         = "cli-nuget-test-build"
	NugetRemoteRepo                        = "jfrog-cli-tests-nuget-remote-repo"
	Out                                    = "out"
	PipBuildName                           = "cli-pip-test-build"
	PypiRemoteRepositoryConfig             = "pypi_remote_repository_config.json"
	PypiVirtualRepositoryConfig            = "pypi_virtual_repository_config.json"
	RepoDetailsUrl                         = "api/repositories/"
	RtServerId                             = "rtTestServerId"
	SearchAllRepo1                         = "search_all_repo1.json"
	SearchGo                               = "search_go.json"
	SearchRepo1ByInSuffix                  = "search_repo1_by_in_suffix.json"
	SearchRepo1TestResources               = "search_repo1_test_resources.json"
	SearchRepo2                            = "search_repo2.json"
	SearchSimplePlaceholders               = "search_simple_placeholders.json"
	SearchTargetInRepo2                    = "search_target_in_repo2.json"
	SearchTxt                              = "search_txt.json"
	SetDeletePropsSpec                     = "set_delete_props_spec.json"
	SpecsTestRepositoryConfig              = "specs_test_repository_config.json"
	SplitUploadSpecA                       = "upload_split_spec_a.json"
	SplitUploadSpecB                       = "upload_split_spec_b.json"
	Temp                                   = "tmp"
	UploadEmptyDirs                        = "upload_empty_dir_spec.json"
	UploadFileWithParenthesesSpec          = "upload_file_with_parentheses.json"
	UploadFlatNonRecursive                 = "upload_flat_non_recursive.json"
	UploadFlatRecursive                    = "upload_flat_recursive.json"
	UploadMultipleFileSpecs                = "upload_multiple_file_specs.json"
	UploadSimplePlaceholders               = "upload_simple_placeholders.json"
	UploadSpecExclude                      = "upload_spec_exclude.json"
	UploadSpecExcludeRegex                 = "upload_spec_exclude_regex.json"
	UploadTempWildcard                     = "upload_temp_wildcard.json"
	UploadWithPropsSpec                    = "upload_with_props_spec.json"
	VirtualRepositoryConfig                = "specs_virtual_repository_config.json"
	WinBuildAddDepsSpec                    = "win_simple_build_add_deps_spec.json"
	WinSimpleDownloadSpec                  = "win_simple_download_spec.json"
	WinSimpleUploadSpec                    = "win_simple_upload_spec.json"
	ReplicationTempCreate                  = "replication_push_create.json"
)

var Repo1 = "jfrog-cli-tests-repo1"
var Repo2 = "jfrog-cli-tests-repo2"
var Repo1And2 = "jfrog-cli-tests-repo*"
var VirtualRepo = "jfrog-cli-tests-virtual-repo"
var JcenterRemoteRepo = "jfrog-cli-tests-jcenter-remote"
var LfsRepo = "jfrog-cli-tests-lfs-repo"
var DebianRepo = "jfrog-cli-tests-debian-repo"
var NpmLocalRepo = "jfrog-cli-tests-npm-local-repo"
var NpmRemoteRepo = "jfrog-cli-tests-npm-remote-repo"
var GoLocalRepo = "jfrog-cli-tests-go-local-repo"
var PypiRemoteRepo = "jfrog-cli-tests-pypi-remote-repo"
var PypiVirtualRepo = "jfrog-cli-tests-pypi-virtual-repo"

func GetTxtUploadExpectedRepo1() []string {
	return []string{
		Repo1 + "/cliTestFile.txt",
	}
}

func GetSimpleUploadExpectedRepo1() []string {
	return []string{
		Repo1 + "/test_resources/a3.in",
		Repo1 + "/test_resources/a1.in",
		Repo1 + "/test_resources/a2.in",
		Repo1 + "/test_resources/b2.in",
		Repo1 + "/test_resources/b3.in",
		Repo1 + "/test_resources/b1.in",
		Repo1 + "/test_resources/c2.in",
		Repo1 + "/test_resources/c1.in",
		Repo1 + "/test_resources/c3.in",
	}
}

func GetSimpleWildcardUploadExpectedRepo1() []string {
	return []string{
		Repo1 + "/upload_simple_wildcard/github.com/github.in",
	}
}

func GetSimpleUploadSpecialCharNoRegexExpectedRepo1() []string {
	return []string{
		Repo1 + "/a1.in",
	}
}

func GetSimpleUploadSpecialCharNoRegexExpectedRepo2() []string {
	return []string{
		Repo2 + "/a1.in",
	}
}

func GetSimpleUploadSpecialCharNoRegexExpected2filesRepo1() []string {
	return []string{
		Repo1 + "/a1.in",
		Repo1 + "/a2.in",
	}
}

func GetSimpleUploadSpecialCharNoRegexExpected2filesRepo2() []string {
	return []string{
		Repo2 + "/a1.in",
		Repo2 + "/a2.in",
	}
}

func GetUploadSpecExcludeRepo1() []string {
	return []string{
		Repo1 + "/a1.in",
		Repo1 + "/b1.in",
		Repo1 + "/c2.in",
		Repo1 + "/c3.in",
	}
}

func GetUploadDebianExpected() []string {
	return []string{
		DebianRepo + "/data/a1.in",
		DebianRepo + "/data/a2.in",
		DebianRepo + "/data/a3.in",
	}
}

func GetSingleFileCopy() []string {
	return []string{
		Repo2 + "/path/a1.in",
	}
}

func GetSingleFileCopyFullPath() []string {
	return []string{
		Repo2 + "/path/inner/a1.in",
	}
}

func GetSingleInnerFileCopyFullPath() []string {
	return []string{
		Repo2 + "/path/path/inner/a1.in",
	}
}

func GetFolderCopyTwice() []string {
	return []string{
		Repo2 + "/path/inner/a1.in",
		Repo2 + "/path/path/inner/a1.in",
	}
}

func GetFolderCopyIntoFolder() []string {
	return []string{
		Repo2 + "/path/path/inner/a1.in",
	}
}

func GetSingleDirectoryCopyFlat() []string {
	return []string{
		Repo2 + "/inner/a1.in",
	}
}

func GetAnyItemCopy() []string {
	return []string{
		Repo2 + "/path/inner/a1.in",
		Repo2 + "/someFile",
	}
}

func GetAnyItemCopyRecursive() []string {
	return []string{
		Repo2 + "/a/b/a1.in",
		Repo2 + "/aFile",
	}
}

func GetCopyFolderRename() []string {
	return []string{
		Repo2 + "/newPath/inner/a1.in",
	}
}

func GetAnyItemCopyUsingSpec() []string {
	return []string{
		Repo2 + "/a1.in",
	}
}

func GetExplodeUploadExpectedRepo1() []string {
	return []string{
		Repo1 + "/a/a3.in",
		Repo1 + "/a/a1.in",
		Repo1 + "/a/a2.in",
		Repo1 + "/a/b/b1.in",
		Repo1 + "/a/b/b2.in",
		Repo1 + "/a/b/b3.in",
		Repo1 + "/a/b/c/c1.in",
		Repo1 + "/a/b/c/c2.in",
		Repo1 + "/a/b/c/c3.in",
	}
}

func GetCopyFileNameWithParentheses() []string {
	return []string{
		Repo2 + "/testsdata/b/(/(.in",
		Repo2 + "/testsdata/b/(b/(b.in",
		Repo2 + "/testsdata/b/)b/)b.in",
		Repo2 + "/testsdata/b/b(/b(.in",
		Repo2 + "/testsdata/b/b)/b).in",
		Repo2 + "/testsdata/b/(b)/(b).in",
		Repo2 + "/testsdata/b/)b)/)b).in",
		Repo2 + "/(/b(.in",
		Repo2 + "/()/(b.in",
		Repo2 + "/()/testsdata/b/(b)/(b).in",
		Repo2 + "/(/testsdata/b/(/(.in.zip",
		Repo2 + "/(/in-b(",
		Repo2 + "/(/b(.in-up",
		Repo2 + "/c/(.in.zip",
	}
}
func GetUploadFileNameWithParentheses() []string {
	return []string{
		Repo1 + "/(.in",
		Repo1 + "/(b.in",
		Repo1 + "/)b.in",
		Repo1 + "/b(.in",
		Repo1 + "/b).in",
		Repo1 + "/(b).in",
		Repo1 + "/)b).in",
		Repo1 + "/(new)/testsdata/b/(/(.in",
		Repo1 + "/(new)/testsdata/b/(b/(b.in",
		Repo1 + "/(new)/testsdata/b/b(/b(.in",
		Repo1 + "/new)/testsdata/b/b)/b).in",
		Repo1 + "/new)/testsdata/b/(b)/(b).in",
		Repo1 + "/(new/testsdata/b/)b)/)b).in",
		Repo1 + "/(new/testsdata/b/)b/)b.in",
	}
}

func GetMoveCopySpecExpected() []string {
	return []string{
		Repo2 + "/copy_move_target/a1.in",
		Repo2 + "/copy_move_target/a2.in",
		Repo2 + "/copy_move_target/a3.in",
		Repo2 + "/copy_move_target/b/b1.in",
		Repo2 + "/copy_move_target/b/b2.in",
		Repo2 + "/copy_move_target/b/b3.in",
		Repo2 + "/copy_move_target/b/c/c1.in",
		Repo2 + "/copy_move_target/b/c/c2.in",
		Repo2 + "/copy_move_target/b/c/c3.in",
	}
}

func GetRepo1TestResourcesExpected() []string {
	return []string{
		Repo1 + "/test_resources/a1.in",
		Repo1 + "/test_resources/a2.in",
		Repo1 + "/test_resources/a3.in",
		Repo1 + "/test_resources/b/b1.in",
		Repo1 + "/test_resources/b/b2.in",
		Repo1 + "/test_resources/b/b3.in",
		Repo1 + "/test_resources/b/c/c1.in",
		Repo1 + "/test_resources/b/c/c2.in",
		Repo1 + "/test_resources/b/c/c3.in",
	}
}

func GetBuildBeforeCopyExpected() []string {
	return GetBuildBeforeMoveExpected()
}

func GetBuildCopyExpected() []string {
	return []string{
		Repo1 + "/data/a1.in",
		Repo1 + "/data/a2.in",
		Repo1 + "/data/a3.in",
		Repo1 + "/data/b1.in",
		Repo1 + "/data/b2.in",
		Repo1 + "/data/b3.in",
		Repo2 + "/data/a1.in",
		Repo2 + "/data/a2.in",
		Repo2 + "/data/a3.in",
	}
}

func GetGitLfsExpected() []string {
	return []string{
		LfsRepo + "/objects/4b/f4/4bf4c8c0fef3f5c8cf6f255d1c784377138588c0a9abe57e440bce3ccb350c2e",
	}
}

func GetBuildBeforeMoveExpected() []string {
	return []string{
		Repo1 + "/data/b1.in",
		Repo1 + "/data/b2.in",
		Repo1 + "/data/b3.in",
		Repo1 + "/data/a1.in",
		Repo1 + "/data/a2.in",
		Repo1 + "/data/a3.in",
	}
}

func GetBuildMoveExpected() []string {
	return []string{
		Repo1 + "/data/b1.in",
		Repo1 + "/data/b2.in",
		Repo1 + "/data/b3.in",
		Repo2 + "/data/a1.in",
		Repo2 + "/data/a2.in",
		Repo2 + "/data/a3.in",
	}
}

func GetBuildCopyExclude() []string {
	return []string{
		Repo1 + "/data/a1.in",
		Repo1 + "/data/a2.in",
		Repo1 + "/data/a3.in",
		Repo1 + "/data/b1.in",
		Repo1 + "/data/b2.in",
		Repo1 + "/data/b3.in",
		Repo2 + "/data/a1.in",
		Repo2 + "/data/a2.in",
		Repo2 + "/data/a3.in",
	}
}

func GetBuildDeleteExpected() []string {
	return []string{
		Repo1 + "/data/b1.in",
		Repo1 + "/data/b2.in",
		Repo1 + "/data/b3.in",
	}
}

func GetExtractedDownload() []string {
	return []string{
		filepath.Join(Out, "randFile"),
		filepath.Join(Out, "concurrent.tar.gz"),
	}
}

func GetFileWithParenthesesDownload() []string {
	return []string{
		filepath.Join(Out, "testsdata"),
		filepath.Join(Out, "testsdata/b"),
		filepath.Join(Out, "testsdata/b/("),
		filepath.Join(Out, "testsdata/b/(/(.in"),
		filepath.Join(Out, "testsdata/b/(b"),
		filepath.Join(Out, "testsdata/b/(b/(b.in"),
		filepath.Join(Out, "testsdata/b/(b)"),
		filepath.Join(Out, "testsdata/b/(b)/(b).in"),
		filepath.Join(Out, "testsdata/b/)b"),
		filepath.Join(Out, "testsdata/b/)b/)b.in"),
		filepath.Join(Out, "testsdata/b/)b)"),
		filepath.Join(Out, "testsdata/b/)b)/)b).in"),
		filepath.Join(Out, "testsdata/b/b("),
		filepath.Join(Out, "testsdata/b/b(/b(.in"),
		filepath.Join(Out, "testsdata/b/b)"),
		filepath.Join(Out, "testsdata/b/b)/b).in"),
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
	localPathPrefix := filepath.Join("syncDir", "testsdata", "a")
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
	localPathPrefix := filepath.Join("syncDir", "testsdata", "a")
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
	localPathPrefix := "/syncDir/testsdata/archives/"
	return []string{
		Repo1 + localPathPrefix + "a.zip",
		Repo1 + localPathPrefix + "b.zip",
		Repo1 + localPathPrefix + "c.zip",
		Repo1 + localPathPrefix + "d.zip",
	}
}

func GetSyncExpectedDeletesDownloadStep7() []string {
	localPathPrefix := filepath.Join("syncDir", "testsdata", "archives")
	return []string{
		filepath.Join(Out, localPathPrefix, "a.zip"),
		filepath.Join(Out, localPathPrefix, "b.zip"),
		filepath.Join(Out, localPathPrefix, "c.zip"),
		filepath.Join(Out, localPathPrefix, "d.zip"),
	}
}

func GetDownloadWildcardRepo() []string {
	return []string{
		Repo1 + "/path/a1.in",
		Repo2 + "/path/a2.in",
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
		Repo1 + "/multiple/a1.out",
		Repo1 + "/multiple/properties/testsdata/a/b/b2.in",
	}
}

func GetSimplePlaceholders() []string {
	return []string{
		Repo1 + "/simple_placeholders/a-in.out",
		Repo1 + "/simple_placeholders/b/b-in.out",
		Repo1 + "/simple_placeholders/b/c/c-in.out",
	}
}

func GetSimpleDelete() []string {
	return []string{
		Repo1 + "/test_resources/a1.in",
		Repo1 + "/test_resources/a2.in",
		Repo1 + "/test_resources/a3.in",
	}
}

func GetDeleteFolderWithWildcard() []string {
	return []string{
		Repo1 + "/test_resources/a1.in",
		Repo1 + "/test_resources/a2.in",
		Repo1 + "/test_resources/a3.in",
		Repo1 + "/test_resources/b/b1.in",
		Repo1 + "/test_resources/b/b2.in",
		Repo1 + "/test_resources/b/b3.in",
	}
}

func GetSearchIncludeDirsFiles() []generic.SearchResult {
	return []generic.SearchResult{
		{
			Path:  Repo1 + "/",
			Type:  "folder",
			Props: make(map[string][]string, 0),
			Size:  0,
		},
		{
			Path:  Repo1 + "/data",
			Type:  "folder",
			Props: make(map[string][]string, 0),
			Size:  0,
		},
		{
			Path:  Repo1 + "/data/testsdata",
			Type:  "folder",
			Props: make(map[string][]string, 0),
			Size:  0,
		},
		{
			Path:  Repo1 + "/data/testsdata/a",
			Type:  "folder",
			Props: make(map[string][]string, 0),
			Size:  0,
		},
		{
			Path:  Repo1 + "/data/testsdata/a/a1.in",
			Type:  "file",
			Props: make(map[string][]string, 0),
			Size:  7,
		},
		{
			Path:  Repo1 + "/data/testsdata/a/a2.in",
			Type:  "file",
			Props: make(map[string][]string, 0),
			Size:  7,
		},
		{
			Path:  Repo1 + "/data/testsdata/a/a3.in",
			Type:  "file",
			Props: make(map[string][]string, 0),
			Size:  7,
		},
		{
			Path:  Repo1 + "/data/testsdata/a/b",
			Type:  "folder",
			Props: make(map[string][]string, 0),
			Size:  0,
		},
		{
			Path:  Repo1 + "/data/testsdata/a/b/b1.in",
			Type:  "file",
			Props: make(map[string][]string, 0),
			Size:  9,
		},
		{
			Path:  Repo1 + "/data/testsdata/a/b/b2.in",
			Type:  "file",
			Props: make(map[string][]string, 0),
			Size:  9,
		},
		{
			Path:  Repo1 + "/data/testsdata/a/b/b3.in",
			Type:  "file",
			Props: make(map[string][]string, 0),
			Size:  9,
		},
		{
			Path:  Repo1 + "/data/testsdata/a/b/c",
			Type:  "folder",
			Props: make(map[string][]string, 0),
			Size:  0,
		},
		{
			Path:  Repo1 + "/data/testsdata/a/b/c/c1.in",
			Type:  "file",
			Props: make(map[string][]string, 0),
			Size:  11,
		},
		{
			Path:  Repo1 + "/data/testsdata/a/b/c/c2.in",
			Type:  "file",
			Props: make(map[string][]string, 0),
			Size:  11,
		},
		{
			Path:  Repo1 + "/data/testsdata/a/b/c/c3.in",
			Type:  "file",
			Props: make(map[string][]string, 0),
			Size:  11,
		},
	}
}

func GetSearchNotIncludeDirsFiles() []generic.SearchResult {
	return []generic.SearchResult{
		{
			Path:  Repo1 + "/data/testsdata/a/a1.in",
			Type:  "file",
			Props: make(map[string][]string, 0),
			Size:  7,
		},
		{
			Path:  Repo1 + "/data/testsdata/a/a2.in",
			Type:  "file",
			Props: make(map[string][]string, 0),
			Size:  7,
		},
		{
			Path:  Repo1 + "/data/testsdata/a/a3.in",
			Type:  "file",
			Props: make(map[string][]string, 0),
			Size:  7,
		},
		{
			Path:  Repo1 + "/data/testsdata/a/b/b1.in",
			Type:  "file",
			Props: make(map[string][]string, 0),
			Size:  9,
		},
		{
			Path:  Repo1 + "/data/testsdata/a/b/b2.in",
			Type:  "file",
			Props: make(map[string][]string, 0),
			Size:  9,
		},
		{
			Path:  Repo1 + "/data/testsdata/a/b/b3.in",
			Type:  "file",
			Props: make(map[string][]string, 0),
			Size:  9,
		},
		{
			Path:  Repo1 + "/data/testsdata/a/b/c/c1.in",
			Type:  "file",
			Props: make(map[string][]string, 0),
			Size:  11,
		},
		{
			Path:  Repo1 + "/data/testsdata/a/b/c/c2.in",
			Type:  "file",
			Props: make(map[string][]string, 0),
			Size:  11,
		},
		{
			Path:  Repo1 + "/data/testsdata/a/b/c/c3.in",
			Type:  "file",
			Props: make(map[string][]string, 0),
			Size:  11,
		},
	}
}

func GetSearchPropsStep1() []generic.SearchResult {
	return []generic.SearchResult{
		{
			Path: Repo1 + "/a/a3.in",
			Type: "file",
			Size: 7,
			Props: map[string][]string{
				"a": {"1"},
				"b": {"3"},
				"c": {"3"},
			},
		},
		{
			Path: Repo1 + "/a/b/b2.in",
			Type: "file",
			Size: 9,
			Props: map[string][]string{
				"b": {"1"},
				"c": {"3"},
			},
		},
		{
			Path: Repo1 + "/a/b/b3.in",
			Type: "file",
			Size: 9,
			Props: map[string][]string{
				"a": {"1"},
				"b": {"2"},
				"c": {"3"},
			},
		},
		{
			Path: Repo1 + "/a/b/c/c2.in",
			Type: "file",
			Size: 11,
			Props: map[string][]string{
				"c": {"3"},
			},
		},
		{
			Path: Repo1 + "/a/b/c/c3.in",
			Type: "file",
			Size: 11,
			Props: map[string][]string{
				"c": {"3"},
			},
		},
	}
}

func GetSearchPropsStep2() []generic.SearchResult {
	return []generic.SearchResult{
		{
			Path: Repo1 + "/a/a1.in",
			Type: "file",
			Size: 7,
			Props: map[string][]string{
				"a": {"2"},
				"b": {"3"},
			},
		},
		{
			Path: Repo1 + "/a/a2.in",
			Type: "file",
			Size: 7,
			Props: map[string][]string{
				"a": {"1"},
			},
		},
		{
			Path: Repo1 + "/a/b/b1.in",
			Type: "file",
			Size: 9,
			Props: map[string][]string{
				"a": {"1"},
				"c": {"5"},
			},
		},
		{
			Path: Repo1 + "/a/b/c/c1.in",
			Type: "file",
			Size: 11,
			Props: map[string][]string{
				"b": {"1"},
			},
		},
	}
}

func GetSearchPropsStep3() []generic.SearchResult {
	return []generic.SearchResult{
		{
			Path: Repo1 + "/a/a1.in",
			Type: "file",
			Size: 7,
			Props: map[string][]string{
				"a": {"2"},
				"b": {"3"},
			},
		},
		{
			Path: Repo1 + "/a/a2.in",
			Type: "file",
			Size: 7,
			Props: map[string][]string{
				"a": {"1"},
			},
		},
		{
			Path: Repo1 + "/a/a3.in",
			Type: "file",
			Size: 7,
			Props: map[string][]string{
				"a": {"1"},
				"b": {"3"},
				"c": {"3"},
			},
		},
		{
			Path: Repo1 + "/a/b/b1.in",
			Type: "file",
			Size: 9,
			Props: map[string][]string{
				"a": {"1"},
				"c": {"5"},
			},
		},
		{
			Path: Repo1 + "/a/b/b2.in",
			Type: "file",
			Size: 9,
			Props: map[string][]string{
				"b": {"1"},
				"c": {"3"},
			},
		},
		{
			Path: Repo1 + "/a/b/c/c1.in",
			Type: "file",
			Size: 11,
			Props: map[string][]string{
				"b": {"1"},
			},
		},
		{
			Path: Repo1 + "/a/b/c/c2.in",
			Type: "file",
			Size: 11,
			Props: map[string][]string{
				"c": {"3"},
			},
		},
		{
			Path: Repo1 + "/a/b/c/c3.in",
			Type: "file",
			Size: 11,
			Props: map[string][]string{
				"c": {"3"},
			},
		},
	}
}

func GetSearchPropsStep4() []generic.SearchResult {
	return []generic.SearchResult{
		{
			Path: Repo1 + "/a/a3.in",
			Type: "file",
			Size: 7,
			Props: map[string][]string{
				"a": {"1"},
				"b": {"3"},
				"c": {"3"},
			},
		},
		{
			Path: Repo1 + "/a/b/b2.in",
			Type: "file",
			Size: 9,
			Props: map[string][]string{
				"b": {"1"},
				"c": {"3"},
			},
		},
		{
			Path: Repo1 + "/a/b/c/c2.in",
			Type: "file",
			Size: 11,
			Props: map[string][]string{
				"c": {"3"},
			},
		},
		{
			Path: Repo1 + "/a/b/c/c3.in",
			Type: "file",
			Size: 11,
			Props: map[string][]string{
				"c": {"3"},
			},
		},
	}
}

func GetSearchPropsStep5() []generic.SearchResult {
	return make([]generic.SearchResult, 0)
}

func GetSearchPropsStep6() []generic.SearchResult {
	return []generic.SearchResult{
		{
			Path: Repo1 + "/a/b/c/c1.in",
			Type: "file",
			Size: 11,
			Props: map[string][]string{
				"b": {"1"},
			},
		},
	}
}

func GetSearchResultAfterDeleteByPropsStep1() []generic.SearchResult {
	return []generic.SearchResult{
		{
			Path: Repo1 + "/a/a1.in",
			Type: "file",
			Size: 7,
			Props: map[string][]string{
				"a": {"2"},
				"b": {"3"},
			},
		},
		{
			Path: Repo1 + "/a/a2.in",
			Type: "file",
			Size: 7,
			Props: map[string][]string{
				"a": {"1"},
			},
		},
		{
			Path: Repo1 + "/a/a3.in",
			Type: "file",
			Size: 7,
			Props: map[string][]string{
				"a": {"1"},
				"b": {"3"},
				"c": {"3"},
			},
		},
		{
			Path: Repo1 + "/a/b/b2.in",
			Type: "file",
			Size: 9,
			Props: map[string][]string{
				"b": {"1"},
				"c": {"3"},
				"D": {"5"},
			},
		},
		{
			Path: Repo1 + "/a/b/b3.in",
			Type: "file",
			Size: 9,
			Props: map[string][]string{
				"a": {"1"},
				"b": {"2"},
				"c": {"3"},
				"D": {"5"},
			},
		},
		{
			Path: Repo1 + "/a/b/c/c2.in",
			Type: "file",
			Size: 11,
			Props: map[string][]string{
				"c": {"3"},
				"D": {"2"},
			},
		},
		{
			Path: Repo1 + "/a/b/c/c3.in",
			Type: "file",
			Size: 11,
			Props: map[string][]string{
				"c": {"3"},
				"D": {"2"},
			},
		},
	}
}

func GetSearchResultAfterDeleteByPropsStep2() []generic.SearchResult {
	return []generic.SearchResult{
		{
			Path: Repo1 + "/a/a1.in",
			Type: "file",
			Size: 7,
			Props: map[string][]string{
				"a": {"2"},
				"b": {"3"},
			},
		},
		{
			Path: Repo1 + "/a/a2.in",
			Type: "file",
			Size: 7,
			Props: map[string][]string{
				"a": {"1"},
			},
		},
		{
			Path: Repo1 + "/a/a3.in",
			Type: "file",
			Size: 7,
			Props: map[string][]string{
				"a": {"1"},
				"b": {"3"},
				"c": {"3"},
			},
		},
		{
			Path: Repo1 + "/a/b/b3.in",
			Type: "file",
			Size: 9,
			Props: map[string][]string{
				"a": {"1"},
				"b": {"2"},
				"c": {"3"},
				"D": {"5"},
			},
		},
	}
}

func GetSearchResultAfterDeleteByPropsStep3() []generic.SearchResult {
	return []generic.SearchResult{
		{
			Path: Repo1 + "/a/a1.in",
			Type: "file",
			Size: 7,
			Props: map[string][]string{
				"a": {"2"},
				"b": {"3"},
			},
		},
		{
			Path: Repo1 + "/a/a3.in",
			Type: "file",
			Size: 7,
			Props: map[string][]string{
				"a": {"1"},
				"b": {"3"},
				"c": {"3"},
			},
		},
	}
}

func GetMavenDeployedArtifacts() []string {
	return []string{
		Repo1 + "/org/jfrog/cli-test/1.0/cli-test-1.0.jar",
		Repo1 + "/org/jfrog/cli-test/1.0/cli-test-1.0.pom",
	}
}

func GetGradleDeployedArtifacts() []string {
	return []string{
		Repo1 + "/minimal-example/1.0/minimal-example-1.0.jar",
	}
}

func GetNpmDeployedScopedArtifacts() []string {
	return []string{
		NpmLocalRepo + "/@jscope/jfrog-cli-tests/-/jfrog-cli-tests-1.0.0.tgz",
	}
}
func GetNpmDeployedArtifacts() []string {
	return []string{
		NpmLocalRepo + "/jfrog-cli-tests/-/jfrog-cli-tests-1.0.0.tgz",
	}
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
		Repo1 + "/syncDir/testsdata/a/a3.in",
		Repo1 + "/syncDir/testsdata/a/a1.in",
		Repo1 + "/syncDir/testsdata/a/a2.in",
		Repo1 + "/syncDir/testsdata/a/b/b1.in",
		Repo1 + "/syncDir/testsdata/a/b/b2.in",
		Repo1 + "/syncDir/testsdata/a/b/b3.in",
		Repo1 + "/syncDir/testsdata/a/b/c/c1.in",
		Repo1 + "/syncDir/testsdata/a/b/c/c2.in",
		Repo1 + "/syncDir/testsdata/a/b/c/c3.in",
	}
}

func GetUploadExpectedRepo1SyncDeleteStep2() []string {
	return []string{
		Repo1 + "/syncDir/testsdata/a/a3.in",
		Repo1 + "/syncDir/testsdata/a/a1.in",
		Repo1 + "/syncDir/testsdata/a/a2.in",
		Repo1 + "/syncDir/testsdata/a/b/b1.in",
		Repo1 + "/syncDir/testsdata/a/b/c/c1.in",
	}
}

func GetUploadExpectedRepo1SyncDeleteStep3() []string {
	return []string{
		Repo1 + "/syncDir/a.zip",
		Repo1 + "/syncDir/b.zip",
		Repo1 + "/syncDir/c.zip",
		Repo1 + "/syncDir/d.zip",
	}
}

func GetReplicationConfig() []services.PushReplicationParams {
	return []services.PushReplicationParams{
		{
			URL:      *RtUrl,
			Username: *RtUser,
			Password: "",
			CommonReplicationParams: services.CommonReplicationParams{
				CronExp:                "0 0 12 * * ?",
				RepoKey:                Repo1,
				EnableEventReplication: false,
				SocketTimeoutMillis:    15000,
				Enabled:                true,
				SyncDeletes:            true,
				SyncProperties:         true,
				SyncStatistics:         false,
				PathPrefix:             "/my/path",
			},
		},
	}
}
