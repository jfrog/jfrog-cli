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
