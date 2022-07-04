# Release Notes
## 1.53.4 (July 4, 2022)
- Bug fix - Upload with target-prop's value contain special characters

## 1.53.3 (May 13, 2022)
- Upgrade golang.org/x/crypto
- Bug fix - jfrog rt bp may throw "panic: runtime error: invalid memory address or nil pointer dereference"

## 1.53.2 (March 13, 2022)
- Upgrade go to 1.17.7
- Update the CentOS image used for creating jfrog-cli's rpm package
- Update dependencies

## 1.53.1 (January 31, 2022)
- Build JFrog CLI with go 1.17.2
- Bug fix - The logged build-info link is wrong when used with projects

## 1.53.0 (January 18, 2022)
- The "M2_HOME" environment variable is now optional when running the "jfrog rt mvn" command
- Bug fix - The JSON summary returned by some commands can be corrupted, in case no files are affected
- Bug fix - "jfrog rt release-bundle-create" - If the command fails, the command ends with 0 as the exit code
- Bug fix - "jfrog rt pip" - The module name for the build info isn't extracted in some cases from the setup.py file
- Bug fix - "jfrog rt pip" - Parsing the egg_info command does not always pass successfully. Instead, search for the PKG-INFO file in the known eggBase dir
- Bug fix - Uploading to Artifactory with the ANT pattern option set, may miss some of the files to be uploaded
- Bug fix - When searching files with the exclusions and transitive options set, the search may not provide accurate results

## 1.52.0 (October 10, 2021)
- Allow searching in Artifactory by build, even if the build is included in a project

## 1.51.2 (August 31, 2021)
- Sign JFrog CLI's RPM package 

## 1.51.1 (August 28, 2021)
- Bug fix - "jfrog rt npm-publish" may read the wrong package.json and therefore fetch the wrong package name and number
- Bug fix - "jfrog rt upload" with --archive and --include-dirs may leaves out empty directories
- Bug fix - Uploading to Artifactory with the archive option fails if the symlink target does not exist

## 1.51.0 (August 9, 2021)
- "jfrog rt mvn" - Support including / excluding deployed artifacts
- "jfrog rt search" - Allow searching in Artifactory by build, even if the build is included in a project
- "jfrog rt upload" - Allow storing symlinks in an archive when uploading it to Artifactory
- Bug fix - Gradle builds which use an old version of the Gradle Artifactory Plugin may fail to deploy artifacts
- Bug fix - The build-info URL is incorrect, in case the build name and number include special characters
- Bug fix - SSH authantication with Artifactory cannot be used without a passphrase
- Bug fix - When searching and filtering by the latest build run, the latest build run isn't always returned
- Bug fix - "jfrog rt build-discard" - the --project flag is missing
- Bug fix - npm-publish may fail if package.json has pre/post pack scripts

## 1.50.2 (July 14, 2021)
- Bug fix - "jfrog rt docker-push" and "jfrog rt docker-pull" commands fail

## 1.50.1 (July 14, 2021)
- Bug fix - When using the --detailed-summary option, the returned upload path is incorrect for the "jfrog rt gp" and "jfrog rt mvn" commands
- Bug fix - When using the --detailed-summary option, there are additional log messages added to stdout, making it impossible to parse the summary

## 1.50.0 (June 24, 2021)
- New --retries option for the search, set-props, delete-props, delete, copy and move commands

## 1.49.0 (June 17, 2021)
- New --detailed-summary option added to the "jfrog rt mvn", "jfrog rt gradle", "jfrog rt dp" and "jfrog rt gp" commands
- The "jfrog rt s", "jfrog rt del", "jfrog rt sp" and "jfrog rt delp" commands no longer require the pattern argument when used with the --build or --bundle options
- Bug fix - JFrog CLI's rpm package license was updated to Apache-2.0

## 1.48.1 (May 28, 2021)
- Bug fix - "jfrog rt go get" fails to collect build-info, if used with an internal module package

## 1.48.0 (May 23, 2021)
- New "jfrog ci-setup" command
- Support for yarn - new "jfrog rt yarn" command
- New --detailed-summary option added to the "jfrog rt npm-publish" command
- New --detailed-summary option added to the release-bundle create and sign commands
- Bug fix - Temp files are not deleted after download
- Bug fix - Change the permission of the npmrc file created by the "jfrog rt npmi" command

## 1.47.3 (May 15, 2021)
- Bug fix - "jfrog rt upload" - using ANT patterns fails to convert doube asteriks to directory range.
- Bug fix - "jfrog rt npm-install" can fail when .npmrc includes 'json=true'.
- Bug fix - "jfrog rt nuget" & "jfrog rt dotnet" can fail when there are multiple .net projects in the same directory.
- Bug fix - "jfrog rt build-publish" module type is missing in build-info modules.
- JFrog CLI binaries are now also published for the ppc64 and ppc64le Linux architectures.
- "jfrog config add" - New --overwrite option.

## 1.47.2 (May 5, 2021)
- Bug fix - the "jfrog rt bpr" command ignores the JFROG_CLI_BUILD_PROJECT environment variable.
- Bug fix - Unable to upload a file if its name includes semicolons.
- Bug fix - Upgrade jfrog-client-go, which includes the upgrade of go-git v4.7.1, to resolve errors that occur when collecting data from the local git repository.

## 1.47.1 (April 29, 2021)
- Bug fix - Error when unmarshalling response received from JFrog Distribution

## 1.47.0 (April 28, 2021)
- "jfrog rt bp" -  New --detailed-summary option added
- "jfrog rt u" - The --detailed-summary option now also returns sha256 of the uploaded files
- The maven and gradle extractors were upgraded
- The value of the JFROG_CLI_USER_AGENT environment variable now also controls the agent name in the build-info
- Bug fix - The dryRun option of release bundle management APIs returns an error 
- Bug fix - Cannot install a jfrog-cli plugin before uninstalling the installed version
- Bug fix - The "jfrog rt bpr" command ignores the --project option

## 1.46.4 (April 19, 2021)
- Bug fix - Download fails with panic, if filtered build does not exist
- Bug fix - Remove rt URL validation on config command
- Bug fix - 'jfrog --version' shows error an error on windows 2012
- Bug fix - Panic is thrown when providing a wrong image tag
- Bug fix - Config import can fail is some scenarios
- Ignores the transitive option when downloading, if Artifactory version is not compatible
- Modify the separator used for creating the temp dir, which stored the build-info before it is published
- Support for npm 7.7

## 1.46.3 (April 16, 2021)
- Bug fix - "jfrog rt u" can fail while reading the latest git commit message

## 1.46.2 (April 15, 2021)
- Bug fix - "jfrog rt dl" with --explode can fail on windows

## 1.46.1 (April 5, 2021)
- Bug fix - "jfrog xr curl" and "jfrog rt curl" don't recognize the --server-id option.

## 1.46.0 (April 4, 2021)
- Breaking change - The "jfrog rt go-recursive-command" is now removed
- New "jfrog xr curl" command
- New "transitive" option has been added to Artifactory's search and download commands, to expand the search to include remote repos (requires Artifactory 7.17)
- The JFROG_CLI_JCENTER_REMOTE_SERVER and JFROG_CLI_JCENTER_REMOTE_REPO environment variables are now deprecated, and replaced with the new (single) JFROG_CLI_EXTRACTORS_REMOTE environment variable
- The "jfrog rt go-publish" command now uses the configuration added by the "jfrog rt go-config" command.

## 1.45.2 (March 18, 2021)
- Bug fix - Uploading files to Artifactory with "archive=zip" causes high memory consumption.
- Add VCS commit message to buildinfo.
- NPM build-info now includes depedencies hierarchy.
- Block the usage of "excludeProps" with "aql" in file specs.
- Validate container name when pushing docker images.
- New File Spec schema added. This schema can help to build and validate File Specs.

## 1.45.1 (March 10, 2021)
- Bug fix - panic when running build-docker-create command if image name does not include slash or colon.
- Bug fix - wrong usage for the config 'add' and 'edit' commands.
- Bug fix - missing --server-id option for the "bad" command.
- Bug fix - Missing --project option for the "bce" command.

## 1.45.0 (March 9, 2021)
- New "jfrog config" command, replacing the old "jfrog rt config" command.
- Support for private JFrog CLI Plugins.
- New full-jfrog-cli docker image
- "jfrog rt build-add-dependencies" - support collecting the dependencies from Artifactory.
- Allow uploading files to Artifactory, after packing them in a zip archive.
- Allow specifying Artifactory project, when publishing build-info.
- Support for ANT patterns when uploading files to Artifactory.
- Download artifacts of all builds, including aggregated (referenced) builds.
- Support fetching VCS attached properties, when git submodules is used.
- Add new "vcs.branch" property to uploaded build artifacts.
- Bug fix - prompt for Artifactory's SSH passphrase when using "jfrog config".
- Bug fix - "jfrog rt build-add-git" - support the case where the revision no longer exists in the build log.

## 1.44.0 (January 31, 2021)
- New users-create and user-delete commands
- New group-create, group-update and group-delete commands
- Allow running pip install without setup.py
- Bug fix - docker - missing clear error message when working with a repo that doesn't exist
- Artifactory upload - the "Props" command option was deprecated and replaced with "TargetProps"
- Support installing jfrog-plugins on s390x linux arch

## 1.43.2 (January 6, 2021)
- Bug fix - docker push fails to set props for virtual repositories
- Bug fix - config import fails with refreshable token

## 1.43.1 (December 30, 2020)
- Embedded help fixes

## 1.43.0 (December 30, 2020)
- Add Podman & Kaniko support
- Improve tests and release process
- Bug fix - Access token create without username
- Bug fix - Artifactory download with the explode option fails to extract files with no r permissions
- Bug fix - Artifactory download fails to download files for '.' as target path

## 1.42.3 (December 18, 2020)
- Bug fix - Download and explode fails while inside a docker container

## 1.42.2 (December 17, 2020)
- Bug fix - Revert "Change pypi AQL query" following performance issues
- Bug fix - Download explode fail when overriding the archive filename 

## 1.42.1 (December 16, 2020)
- Bug fix - Explode after download fails, if not downloaded to the curr dir 

## 1.42.0 (December 15, 2020)
- Build-info performance improvements for npm and pip
- "jfrog rt cp" and "jfrog rt mv" are now parallel
- New "jfrog rt build-append" command
- Support VCS list in build-info
- New additional "cross action" progress bar for Artifactory uploads and downloads
- Progress bar added to "jfrog plugin install"
- Add color to log leval prefix in logs
- Nuget v3 is now used by default
- Bug fix - download with the explode option fails in some dcenarios
- Bug fix - Artifactory upload with sync-deletes can delete when shouldn't, if path includes dashes
- Bug fix - Support white-spaces in args for the "jfrog rt mvn", "jfrog rt nuget" and "jfrog rt dotnet"

## 1.41.2 (November 25, 2020)
- Bug fix - User prompts in plugins are invisible
- Bug fix - Occasional panic on TestInsecureTlsMavenBuild
- Bug fix - Avoid using "ParseArgs" for new syntax package manager commands
- Bug fix - Error when uploading / downloading 0 files 

## 1.41.1 (November 13, 2020)
- Bug fix - With multiple JFrog CLI plugins installed, the wrong plugin gets executed 

## 1.41.0 (November 12, 2020)
- New "JFrog CLI Plugins" feature
- Bug fix - "jfrog rt pip-install" fails with local repos in some scenarios

## 1.40.1 (November 2, 2020)
- Publish Linux s390x architecture binary of JFrog CLI
- Bug fix - Upload --detailed-summary output 

## 1.40.0 (October 26, 2020)
- “jfrog rt upload” - New --detailed-summary option
- Build-info support for “jfrog rt go get” with Artifactory
- “jfrog rt access-token-create” - the username argument is now optional
- “jfrog rt repo-delete” now accepts a wildcard pattern, allowing the deletion of multiple repositories
- Allow different JFrog CLI version to use different config files formats
- Upgrade maven and gradle extractors
- Bug fix - “jfrog rt mvn” - not all artifacts are downloaded from Artifactory
- Bug fix - “jfrog rt donet” does not generate build-info for projects which include vbproj files
- Bug fix - Set / delete props is always recursive
- Bug fix - The algorithm which turns wildcard pattern to AQL produces wrong results in some scenarios

## 1.39.7 (October 13, 2020)
- Add refreshable tokens exclusion for default server
- Bugfix - Docker push fails to update layers properties

## 1.39.6 (October 11, 2020)
- Bug fix - docker pull fails to collect buildinfo from a remote repo
- Bug fix - Command summary gets cut by the progress bar
- Bug fix - Bug fix - Artifactory upload - Allow escaping slashes using back-slashes as part of the build name value

## 1.39.5 (September 23, 2020)
- Bug fix - "jfrog rt mvn install" throws `Target repository cannot be empty` even when deployment is disabled.
- Bug fix - Download & Sync command fails in Windows environment.
- Bug fix - Using `--insecure-tls` does not work with automatically created access token.

## 1.39.4 (September 10, 2020)
- Bug fix - Set/Delete props unified errors output.
- Bug fix - Delete props command always returns a failure.
- Bug fix - Fail to copy files with the same prefix name.
- Bug fix - Search ignores "order by".

## 1.39.3 (September 3, 2020)
- Bug fix - missing defer close after using AQL service.

## 1.39.2 (August 29, 2020)
- Bug fix - "jfrog rt u" with --sync-deletes deletes wrong folders.

## 1.39.1 (August 24, 2020)
- Bug fix - "jfrog rt release-bundle-create" and "jfrog rt release-bundle-update" fail when used with file spec which includes "aql".

## 1.39.0 (August 16, 2020)
- New docker promote command.
- New Permission Target commands - Interactive template, create, update and delete.
- Update alpine to version 3.12.0 in docker image.
- Show command flag names sorted.
- Improved search infrastructure.

## 1.38.4 (August 10, 2020)
- Bug fix - "panic: Logger not initialized" error can be thrown when the config is empty.

## 1.38.3 (August 5, 2020)
- Bug fix - The "jfrog rt replication-template" command creates a template with a wrong target URL.

## 1.38.2 (July 20, 2020)
- Bug fix - Fix an issue with JFrog CLI's npm installer

## 1.38.1 (July 19, 2020)
- Bug fix - Username must be lowercase since version 1.38.0
- JFrog CLI build dir restructure

## 1.38.0 (June 30, 2020)
- "jfrog rt release-bundle-distribute" - New --sync and --max-wait-minutes command options.
- "Changing access tokens hourly" is now enabled by default.
- Distribution commands - new --insecure-tls option.
- "jfrog rt pip" - support for remote repositories.
- Bug fix - "jfrog rt npm-ci" - Fix command help and logging.
- Bug fix - Reintroduce --insecure-tls option to the "jfrog rt c" command.
- Bug fix - "jfrog rt build-publish --dry-run" deletes the locak build-info.
- Bug fix - "jfrog rt set-props" does not support path with special characters.

## 1.37.1 (June 15, 2020)
- "jfrog rt nuget" - Support running NuGet on Ubuntu. even if used without mono.
- New "jfrog bt maven-central-sync" command.

## 1.37.0 (June 9, 2020)
- New sensitive data encryption functionality
- "jfrog rt dotnet" - Support .Net Framework versions prior to 3.1.200
- Add curl and disable interactive prompts in the Docerfile
- Bug fix - JFrog CLI should use "mono" automatically when running NuGet on Linux.
- Bug fix - The changing tokens functionality should be disabled for external tools.
- Bug fix - Support downloading files which use non-ASCII characters in their name.
- Bug fix - Support placeholders use in repository name for download, move and copy.

## 1.36.0 (May 25, 2020)
- "jfrog rt download" - New --details-summary option
- New "jfrog rt dotnet" command supporting .net core CLI support
- "jfrog rt nuget" - support for Linux and MacOS
- New "jfrog rt access-token-create" command
- "jfrog rt download" - Added support for validateSymlinks via file spec
- "jfrog rt search" now also returns SHA1 and MD5
- "jfrog rt build-publish" - The default value of the --env-include option was extended to also include '***psw***'
- Add deployment path to the build-info
- Bug fix - "jfrog rt nuget" does not collect build for all .net project structures
- Breaking change - "pip-config", "nuget-config" - remove redundant flags: --server-id-deploy, --repo-deploy.

## 1.35.5 (May 13, 2020)
- Bug fix - Out of range for docker duplicate layers.
- Bug fix - 'jfrog rt bpr' fails with build params as env vars

## 1.35.4 (May 12, 2020)
- Bug fix - The “jfrog rt docker-push” command may duplicate artifacts added to the build-info, which can cause build-promotion to fail.
- Bug fix - The “jfrog rt docker-pull” command cannot collect build-info when pulling from a remote repo.
- Bug fix - VCS props are not added when uploading only one file.
- Bug fix - add retries to download/upload params

## 1.35.3 (Mar 30, 2020)
- Bug fix - Command arguments and options wrapped with quotes are not parsed properly.

## 1.35.2 (Mar 29, 2020)
- "jfrog rt bag" now includes a new -server-id option
- Improvement - Refreshable tokens for authentication with Artifactory
- Bug Fix - Cannot add arguments with white spaces in package manager related commands

## 1.35.1 (Mar 18, 2020)
- Bug fix - "jfrog rt npm-publish" can publish an incorrect module name.
- Bug fix - The --insecure-tls option is missing for "jfrog rt mvn",
- Bug fix - The replication-create and replication-delete commands are missing.

## 1.35.0 (Mar 17, 2020)
- New repo-create, repo-update and repo-delete commands for Artifactory.
- New replication-create and replication-delete commands for Artifactory.
- "jfriog rt mvn" and "jfrog rt gradle" now support parallel deployment.
- "jfrog rt mvn" - New --insecure-tls option.
- New commands - release-bundle-create, release-bundle-update, release-bundle-sign, release-bundle-distribute, release-bundle-delete
- "jfrog rt delete" now includes a new --threads option.
- "jfrog rt docker-pull" now supports docker fat manifest.
- Support using refreshable tokens for authentication with Artifactory.
- Bug fix - Wring exit code returned for the move, copy, set-props and delete-props commands.
- Bug fix - Docker version checks fails for specific versions.

## 1.34.1 (Mar 4, 2020)
- Improve URL masking reg exp.

## 1.34.0 (Feb 27, 2020)
- Allow filtering files by Release Bundle, when searching, downloading, moving, etc.
- 'exclude-patterns' file spec property and command option is deprecated and replaced by 'exclusions'.
- Allow non-interactive usage of the npm-config, maven-config, gradle-config, nuget-config and go-config commands.
- Disable all interactive prompts when CI=true
- New --client-cert-path and --client-cert-key-path options added to the "jfrog rt c" command.
- New --target option added to the "jfrog xr offline-update" command.
- New --list-download option added to the the "jfrog bt u" command.
- Bug fix - docker version check failing on Windows.
- Bug fix - npm-install and npm-ci commands - JSON output is used by default and cannot be disabled.
- New issues and pull request templates
- The pip-deps-tree command was removed.

## 1.33.2 (Jan 27, 2020)
- Bug fix - The "jfrog rt build-scan" command returns exit code 0 in case of a timeout.
- Bug fix - The --help option is not working for some commands.

## 1.33.1 (Jan 16, 2020)
- Bug fix - Gradle builds are failing.

## 1.33.0 (Jan 15, 2020)
- Replace Mission Control commands to support Mission Control 4.0.0 API.
- Support Artifactory authentication with client certificates.
- Sign JFrog CLI's Windows binary.
- Upgrade maben and gradle extractor versions.
- Validate the minimum supported version of the docker client.
- Bug fix - Build tools config commands do not fetch local repositories.
- Bug fix - Config created by "jfrog rt gradlec" is invalid.
- Bug fix - Do not remove all parenthesis from path on Artifactory upload.
- Bug fix - Wrong command name in mvnc help.

## 1.32.4 (Dec 23, 2019)
- Update jfrog-client-go version to v0.6.3

## 1.32.3 (Dec 19, 2019)
- Bug fix - vcs.url property value.
- Fix maven syntax deprecation message.
- 'jfrog rt mvnc' missing server ID from config file.

## 1.32.2 (Dec 16, 2019)
- Fixes to the CLI help.

## 1.32.1 (Dec 16, 2019)
- New syntax for the “rt mvn”, “rt gradle” and “rt nuget” commands.
- New “vcs.url” and “vcs.revision” properties added when uploading generic files (using”rt upload”) as part of a build.
- “rt docker-push” and “rt docker-pull” - Support for foreign docker layers.
- Bug fix - “jfrog rt mvn” - Can’t find plexus-classworlds-x.x.x.jar
- Bug fix - “jforg rt nuget” - artifacts are added into the wrong module in the build-info.

## 1.31.2 (Nov 27, 2019)
- Bug fix - Stack overflow in npm publish

## 1.31.0 (Nov 26, 2019)
- New and improved syntaxt for the "jfrog rt npm" commands.
- Bug fix - Maven and Gradle artifacts are missing "timestamp" property.

## 1.30.4 (Nov 12, 2019)
- Bug fix - Go builds - replaced packages in go.mod are not handled properly.
- Bug fix - Download with --explode fails with relative path.

## 1.30.3 (Nov 07, 2019)
- Bug fix - Optimize search by build without pattern
- Bug fix - Upload to Artifactory fails when path includes '..'"

## 1.30.2 (Nov 03, 2019)
- Bug fix - "jfrog rt c" with --interactive=false doesn't work as expected.
- Bug fix - Target should be option in download spec.
- Maven and gradle build-info extractors upgraded.

## 1.30.1 (Oct 31, 2019)
- Added support for File Specs for the "jfrog rt set-props" and "jfrog rt delete-props".
- Bug fix - "jfrog rt upload" does not upload when the path includes "..".
- Bug fix - "jfrog rt download" - sync-deletes does not delete when the path is "./"

## 1.30.0 (Oct 24, 2019)
- New --sync-deletes option added to the "jfrog rt download" command.
- The "jfrog rt search" command now also returns the size, created and modified fields.Added a new "jfrog rt pip-dep-tree" command.
- Bug fix - Cannot upload, download, copy, move or delete files from Artifactory, if the file path includes parentheses.
- Bug fix - The "jfrog rt pip-install" command requires admin permissions.
- Bug fix - Some Artifactory commands returns exit code 0 on failure.

## 1.29.2 (Oct 16, 2019)
- Bug fix - "jfrog rt go-publish" creates a zip while ignoring the content of .gitignore.

## 1.29.1 (Sep 26, 2019)
- Bug fix - "jfrog rt build-clean" command fails.

## 1.29.0 (Sep 24, 2019)
- New - "jfrog rt c import" and "jfrog rt c export" commands.
- Support configuring build-name, build-number, build-url and env-exclude from environment variables.
- The "jfrog rt gp" command now sends checksum headers to Artifactory.
- Go 1.13 compatibility.
- Support MAVEN_OPTS environment variable in maven commands.

## 1.28.0 (Sep 01, 2019)
- Support for pip build-info with "jfrog rt pip-install" command.
- New "jfrog rt pip-config" command for pypi build-info collection.
- Artifactory upload, sync-delete - changed attached property name to sync.delete.timestamp.

## 1.27.0 (Aug 16, 2019)
- Support wildcards in the repository name in Artifactory commands.
- New --sync-deletes option added to the "jfrog rt upload" command.
- New --exclude-props option added to the search, download, copy, move, delete and set-props commands.
- New JFROGCLIDEPENDENCIES_DIR environment variable added.
- JFrog CLI now Supports bash and zsh auto-complition.
- Support for ARM OS distributions.
- New --count option added to the search command.
- New --include-dirs option added to the search command.
- Bug fix - Build name and build number are not set correctly for the "jfrog rt gradle" command.
- Bug fix - The --explode option does not work, when the archive exists in the download target repo

## 1.26.3 (Jul 28, 2019)
- Bug fix - Nuget packages in Build-Info should have dependency-id with ':' as separator instead of '/'.

## 1.26.2 (Jul 09, 2019)
- Bug fix - npm commands should resolve Artifactory version using version API.

## 1.26.1 (Jun 16, 2019)
- Bug fix - "Set properties" deleting properties instead of setting

## 1.26.0 (Jun 11, 2019)
- New and improved syntaxt for the "jfrog rt go" command.
- New progress bar for "jfrog rt upload" and "jfrog rt download".
- New optional --module option, to set the build-info module name.
- File Specs - Support forward slashes on Windows.
- Support Artifactory access tokens for maven, gradle, npm, docker, nuget and cUrl.
- The JFrog CLI project structure has been changed to support "go get".
- New "jfrog rt npm-ci" command.
- New --skip-login option in the "jfrog rt pull" and "jfrog rt push" commands.
- New --threads option in the "jfrog rt npm-install" command.
- "jfrog rt go-publish" now also publishes info files from version 6.10.0 of Artifactory.
- Bug fix - Bintray Access Key - the --api-only command option is ignored.
- Optionally send usage information to Artifactory (can be disabled by setting the JFROGCLIREPORT_USAGE env var to false).

## 1.25.0 (Apr 17, 2019)
- New "jfrog rt curl" command.
- "jfrog rt nuget" - Support projects with multiple sln files.
- "jfrog rt build-promote" - Support attaching artifact properties during promotion.
- Support for insecure TLS.
- "jfrog rt download" with File Spec which includes more than one group - improved parallelism.
- Breaking change - "jfrog rt go" - Publish missing go depedencies only with the publish-deps option.
- Bug fix - NPM installation of JFrog CLI - httpsproxy fix.
- Bug fix - "jfrog rt nuget" - Do not add disabled projects to the build-info.
- Bug fix - "jfrog rt config" faiuls with access tokens, if URL has no trailing slash.
- Bug fix - "jfrog rt gradle" and "jfrog rt mvn" - The HTTPPROXY env var is ignored.
- Bug fix - JFROGCLITEMP_DIR env does not take effect for all commands.

## 1.24.3 (Mar 05, 2019)
- Build JFrog CLI with CGO_ENABLED=0 and -ldflags '-w -extldflags "-static"'
- Bug fix - "jfrog rt go" fails when go.mod includes a replace block.
- Bug fix - "jfrog rt go" should encode module versions (tags) and not not only module names.
- Bug fix - Authentication with Artifactory fails, when the value of the --user option include white-spaces.

## 1.24.2 (Feb 24, 2019)
- Bug fix - Support values containing white-spaces in config commands.
- Bug fix - Support using user-name containing white-spaces in docker commands.

## 1.24.1 (Feb 11, 2019)
- Bug fix - "jfrog rt build-add-git" retrieves the wrong build in Artifactory.

## 1.24.0 (Feb 10, 2019)
- "jfrog rt build-add-git" can now add issues to the build-info
- New command - "jfrog rt go-recursive-publish"
- Breaking change - "jfrog rt go" - The --recursive-tidy and --recursive-tidy-overwrite options were-  removed.
- "jfrog rt go" - If package is not found in VCS, it is searched in Artifactory.

## 1.23.2 (Jan 27, 2019)
- Bug fix - Artifactory SSH login - relogin when token expires.
- Bug fix - Connection retry is not performed in the case of EOF.
- Bug fix - Connection may hang in case of a network issue.
- Bug fix - Bintray tests may fail since no connection retries are performed.

## 1.23.1 (Dec 26, 2018)
- Bug fix - "jfrog rt go-publish" may not create the zip properly on Windows.
- Bug fix - "jfrog rt go-publish" mod files of modules with quotes are not uploaded properly to Artifactory.
- Bug fix - Support for a Go API fix in Artifactory 6.6.0 (The fix also requires Artifactory 6.6.1)
- Bug fix - Go - go.mod can be uploaded incorrectly (as an empty file) when using Artifactory 6.6.0
- Bug fix - "jfrog rt docker-push" does not work on Alpine.
- Bug fix - "jfrog mc attach-lic" --license-path does not work properly.

## 1.23.0 (Dec 17, 2018)
- "jfrog rt config" - Support for access tokens.
- File Specs - the "pattern" property is no longer required when using the "build" property.
- Fix for https://www.jfrog.com/jira/browse/RTFACT-17644 (included in Artifactory 6.6).
- Breaking change - In the "jfrog rt go" command, the --deps-tidy option has been replaced with the new --recursive-tidy option.
- "jfrog rt go" - Two new command options added: --recursive-tidy and --recursive-tidy-overwrite.

## 1.22.1 (Dec 03, 2018)
"jfrog rt mnv" and "jfrog rt gradle" can be configured to download the build-info-extractor jars from Artifactory.
The tmp dir used by JFrog CLI cam be configured using the new JFROGCLITEMP_DIR environment variable.

##1.22.0 (Nov 20, 2018)
- Go - Support for complete mod files creation.
- Bug fix - Possible panic when using the --spec-vars command option.

## 1.21.1 (Nov 14, 2018)
- Artifactory SSH authentication behaviour update.
- Xray scan now returns exit code 3 when 'Fail Build' rule is matched by Xray.
- Bug fix - Support docker-pull from remote repositories.
- Bug fix - Wrong source for artifact name in build info.
- Bug fix - Panic when --spec-vars ends with semicolon.
- Bug fix - Make build on freebsd.
- Bug fix - NuGet skip packages with type project.

##  1.21.0 (Nov 06, 2018)
- JFrog CLI is now built with Go modules.
- 'jfrog-client-go' has been moved out to a separate project.
- New docker-pull command for pulling docker images from Artifactory.
- New environment variable 'JFROG_CLI_HOME_DIR' replaces 'JFROG_CLI_HOME'.
- New docker image for JFrog CLI.
- Bug fix - Artifactory downloads returns exit code 1.
- Bug fix - Allow uploading broken symlinks to Artifactory.
- Bug fix - Allow NuGet build-info collection on missing package, assets or package.json.
- Bug fix - Cannot use placeholders when downloading from virtual repositories.

## 1.20.2 (Oct 10, 2018)
- Bug fix - Build properties are not added for go builds in some scenarios when using Nginx. The fix also requires Artifactory 6.5.0 or above

## 1.20.1 (Sep 30, 2018)
- Bug fix - "jfrog rt move/copy" ignore self-signed certificates.

## 1.20.0 (Sep 17, 2018)
- Go builds - The --no-registry option is no longer required, due to automatic fallback to github.
- New --retries option for "jfrog rt upload".
- "jfrog rt docker-push" command now performs "docker login" before execution.
- Go builds bug fixes.
- Bug fix - the sources can now be built with go 1.11.
- Bug fix - "jfrog rt build-scan" ignores self-signed certificates.

## 1.19.1 (Aug 28, 2018)
- Build-info support for the "jfrog rt go" command.

## 1.19.0 (Aug 19, 2018)
- New Discard Builds command for Artifactory
- New Delete Properties command for Artifactory
- Bug fix - Concurrent execution of the config commands.
- "jfrog rt config" - API Key cannot be configured without a Username.

## 1.18.0 (Aug 08, 2018)
- The "jfrog rt go" command now uses "go" instead of "vgo"
- Breaking change: The "jfrog rt search" did not show properties with multiple values. As a a result of this fix, the prop values are now returned as a list.
- Fix for the "jfrog rt search" command - properties were not returned if "sortBy" is used.
- "jfrog rt download" - Fix downloaded file permissions.
- Allow HTTP redirect for POST requests.

## 1.17.1 (Jul 11, 2018)
- Fix comptability issue with VGO that changed the dependency directory from v/cache to mod/cache/download

## 1.17.0 (Jul 04, 2018)
- Support for NuGet build info.
- Add Gradle 4.8 compatibility.
- Add new --archive-entries option for Artifactory search, download, move, copy, delete, set properties.
- Bug fix - Overwrite cli configuration file instead of recreating it.

## 1.16.2 (Jun 13, 2018)
- Xray offline-update bug fix.
- Bug fixes and improvements for vgo.
- Bug fix - Range downloads from Artifactory do not follow redirect.
- Bug fix - "jfrog rt use" exists with exit code 0 even if server ID does not exist.
- Nodified github repository from jfrogdev to jfrog.

## 1.16.1 (May 31, 2018)
- Bug fix - jfrog rt npm-install fails with non-admin user

## 1.16.0 (May 16, 2018)
- Mission Control commands modified to support Mission Control 3.0. Earlier versions are no longer supported.
- Checksum validation for Artifactory download command
- JFrog CLI is now available as an NPM package - https://www.npmjs.com/package/jfrog-cli-go
- Bug fixes

## 1.15.1 (May 03, 2018)
- Bug fix - jfrog rt docker-push fails to work with proxyless configuration
- Bug fix - jfrog bt download-file can return a successful response when the download fails

## 1.15.0 (Apr 30, 2018)
- New Artifactory commands to support Golang
- Added properties to Artifactory's search command response
- Bug fix - Boolean command options must be declared either as "--name=val" or "--name" but not "--name val"
- Bug fixes

## 1.14.0 (Feb 15, 2018)
- Docker Build-Info
- Allow exploding downloaded files from Artifactory
- New build-add-dependencies command
- New "fail-no-op" option for some the Artifactory commands
- New "build-url" option for the build-publish command
- Xray Offline-Update bug fix
- Structural refactor
- Bug fixes

## 1.13.1 (Dec 28, 2017)
- Artifactory gradle builds failure fix.

## 1.13.0 (Dec 28, 2017)
- npm build-info
- Unified JSON output for the Artifactory commands
- Redirect all logging, except JSON output, into std err.
- New "offset" option for the download, copy, move and delete Artifactory commands.
- Bug fixes

## 1.12.1 (Nov 09, 2017)
- Bug fix - When uploading bad files/symlinks, the cli fails even with exclude option enabled
- Bug fix - jfrog rt cp command fails on invalid URL escape

## 1.12.0 (Nov 02, 2017)
- Sort and limit options for Artifactory commands.
- SSH agents and SSH key passphrase support.
- New Xray scan build command for Artifactory
- Xray offline-update command performance improvements.
- In Artifactory's build-publish command, the env-include vale should be case insensitive.
- Fix include-dirs flag for set-props command.

## 1.11.2 (Oct 08, 2017)
- Bug fix - SSH authentication causes nil pointer dereference

## 1.11.1 (Oct 03, 2017)
- Fixes jfrog certificates directory path

## 1.11.0 (Sep 28, 2017)
- Exclude Patterns in File Specs
- Maven and Gradle integration for Artifactory
- Artifactory download retries
- HTTP Proxy Support for Artifactory through httpproxy env var.
- Self-Signed certificates integration tests
- New JFROGCLI_HOME env var
- AQL optimizations
- Improved interactive config command for the Artifactory servers
- CLI help improvements

## 1.10.3 (Aug 31, 2017)
- build-add-git command prompts ".git: is a directory" message

## 1.10.2 (Aug 17, 2017)
- https://github.com/JFrogDev/jfrog-cli-go/issues/74
- https://github.com/JFrogDev/jfrog-cli-go/issues/75

## 1.10.1 (Jul 18, 2017)
- Fix for SSH authentication with Artifactory issue.

## 1.10.0 (Jul 11, 2017)
- New Help documentation - extend the information displayed by using the --help option.
- New Artifactory set-props command.
- Add new --api_only option for bintray access key creation command.
- Improved Xray offline-update error handling.
- Improved files checksum calculation.
- Support specs replace macro using a new --spec-vars command option.
- Add new --version option for Xray offline-update command.
- Extend build-info to support principal field.
- Bug fixes.

##1.9.0 (May 23, 2017)
- Added Git LFS GC support.
- https://github.com/JFrogDev/jfrog-cli-go/issues/23
- Bug fixes.

## 1.8.0 (Mar 30, 2017)
- Support folders upload/download.
- Extend build-info capabilities to collect git details.
- Support multiple Artifactory instances configuration.
- Bug fixes.

## 1.7.1 (Feb 08, 2017)
- Added the "symlinks" option to the Artifactory Download command.

## 1.7.0 (Feb 06, 2017)
- Support symlinks in uploads and downloads to and from Artifactory.
- When deleting paths from Artifactory, empty folders are also removed.
- Added the option of extracting archives after they are uploaded to Artifactory.
- Improvements to the Artifactory upload and download concurrency mechanism.
- Bug fix - Uploads to Artifactory using "~" do not work properly.

## 1.6.0 (Dec 12, 2016)
- Artifactory build promotion command.
- Artifactory build distribution command.
- Improved logging and return values.

## 1.5.2 (Dec 11, 2016)
- Support Bintray stream.

## 1.5.1 (Nov 02, 2016)
- Support prune empty directories during recursive delete
- Support for system ssl certificates
- Integration tests infrastructure

## 1.5.0 (Sep 22, 2016)
- Support build info
- Specs support
- Xray-offline updates
- Bug fixes

## 1.4.1 (Jul 31, 2016)
- Added the JFROGCLILOG_LEVEL environment variable.

## 1.4.0 (Jul 28, 2016)
- Add Artifactory search command.
- Allow using tilde '~' when uploading to Artifactory.
- Bug fixes.

## 1.3.2 (Jun 13, 2016)
- added JFROGCLIOFFER_CONFIG environment variable

## 1.3.1 (Jun 06, 2016)
- JFrog CLI build system fixes.

## 1.3.0 (Jun 06, 2016)
- Added support for Artifactory API key authentication.
- Bypass interactive offer for creating configuration using offer-config global flag.
- Copy and Move commands support props flag for properties.
- Various bug fixes and performance optimizations.

## 1.2.1 (May 18, 2016)
- "deploy" option was added to the "jfrog mc attach-lic" command.

## 1.2.0 (May 05, 2016)
- JFrog Artifactory - Added copy, move and delete.
- JFrog Mission Control - artifactory-instances command.
- Bug fixes and improvements.

## 1.1.0 (Apr 11, 2016)
- Artifactory - support for self sign certificates.
- Bugs fixes.

## 1.0.1 (Mar 15, 2016)
- Changed Artifactory command name from "arti" to "rt".
- Fixed the expected path for the dlf command.

## 1.0.0 (Mar 07, 2016)
- Initial release.
