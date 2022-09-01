# Release Notes

| Release notes moved to https://github.com/jfrog/jfrog-cli/releases |
|----------------------------------------------------------------------------------------------------------------------------------------------------------|


## 2.16.1 (April 21, 2022)
- Update the Go version used to build JFrog CLI to 1.17.9
- Bug fix - The JSON output of the "jf build-scan" command also includes a log message
- Bug fix - The curl installaiton of JFrog CLI may fail to copy the executable to the PATH
- Bug fix - Build-info collection and audit for npm projects nay fail due to peer dependencies conflicts

## 2.16.0 (April 15, 2022)
- Allow the execution of JFrog CLI Plugins which also include resources
- When authenticating using a PEM certificate improve the error message, in case the certificate key isn't provided
- Bug fix - Missing npm depedencies should not return an error

## 2.15.1 (April 12, 2022)
- Bug fix - Update user groups API may remove the user from all groups
- Bug fix - The 'jf audit' command should return an error if no package manager was detected
- Bug fix - The package managers config commands get stuck following a bug in go-prompt
- Improve results table utility by separating emoji from severity title

## 2.15.0 (March 31, 2022)
- Add Operational Risk Violations data to 'jf scan', 'jf build-scan' & 'jf audit' commands
- Add emojis to the output of some commands
- Upgrade jfrog-client-go & build-info-go
- Add --extended-table option to 'jf docker scan'
- Bug fix - Avoid 'jf setup' failing when CI=true
- Bug fix - Avoid adding optional npm dependencies to the build-info

## 2.14.2 (March 27, 2022)
- 'jf setup' command UX improvements
- Bug fix - When uploading to Artifactory, ANT patterns aren't translated correctly to regexp in some cases
- Bug fix - The "jf mvn" command fails with some versions of maven

## 2.14.1 (March 24, 2022)
- "jf c export" - Return an error if config is empty
- Bug fix - Avoid npm checksum calculation by "jf audit", to avoid failires for some npm projects
- Bu fix - "jf audit" for nuget may crash in some scenarios

## 2.14.0 (March 18, 2022)
- Emojis added to log messages
- Log Artifactory response after failure when encrypting the password
- Improve the error log message when using the "jf rt repo-create" command with wrong packageType or rclass
- Support IncludePathPrefixPattern param in the replication API
- Update go.mod to go 1.17
- Static code analysis badge added to README
- Improve the intro message of "jf setup" and "jf project init"
- Improve getting-started-with-jfrog-using-the-cli.md
- Bug fix - Panic in Dotnet and NuGet commands
- Bug fix - Incorrect exit code for the "jf scan" cxommand
- Bug fix - The JFROG_CLI_TRANSITIVE_DOWNLOAD_EXPERIMENTAL env var does not affect the "jf rt dl" command
- Bug fix - "jf docker scan" command ignores the --watches, --project and --repo-path options

## 2.13.0 (February 24, 2022)
- New 'jf docker push' & 'jf docker push' commands
- New support for auditing .NET projects (nuget/dotnet)
- 'jf docker scan' - New progress indicator
- Performance improvement for collecting build-info dependencies
- Update the intro message of the 'jf project init' command to include docker
- Update the go version used to build jfrog-cli to v1.17.7
- Make Zsh completion autoloadable
- Bug fix - Build-info should not be created for 'jf npm install <package name>'
- Bug fix - Limit the total for RequestedBy, to avoid out-of-memory errors
- Bug fix - 'jf project init' - Consider .csproj when detecting .NET projects

## 2.12.1 (February 8, 2022)
- Bug fix - "nil pointer dereference" on jfrog-client-go

## 2.12.0 (January 30, 2022)
- New timestamp added to the log messages
- JFrog CLI now prompts alternative options, if the command does not exist
- "jf c add" command interactive mode improvement - Replaced "Server ID:" with "Choose a server ID:"
- Bug fix - Avoid creating a redundant build-info module in some scenarios
- Bug fix - When searching or downloading from Artifactory with the `transitive` setting, validate that only one repository is requested, since only one repository search is supported with `transitive`
- Bug fix - When walking the file tree during an upload, the same file might be visited more than once, but a dir symlink should not be visited more than once.
- Bug fix - The --scan option used by the "jf mvn / gradle / npm" commands, fails the upload for every vulnerability, even if according to the Xray policy, the upload shouldn't fail
- Bug fix - The "jf audit" command for pip projects ignores the requirements.txt file
- Bug fix - The module added to the build-info by the "jf rt bad" command has no type
- Bug fix - The "jf rt bp" should not prompt the build publish URL with the --dry-run option

## 2.11.1 (January 24, 2022)
- Remove redundant "jf docker scan" flags
- Improve docker scan code and logs. 
- Create a temp folder for the Xray indexer app to run at. 
- Improve Xray indexer logs.

## 2.11.0 (January 13, 2022)
- New "jf docker scan" command

## 2.10.1 (January 4, 2022)
- New --fail-fast option added to the "jf rt build-promote" command
- "jf setup" command UX improvements

## 2.10.0 (December 31, 2021)
- Support for generating build-info for multi-platform images (fat-manifest)
- New --retry-wait-time option added to all commands supporting the --retry option
- Minor improvements to the "jf build-scan" command
- Support for the pipenv package manager
- New --fail option added to the "jf audit" and "jf scan" commands
- Bug fix - "jf pip install" may fail, if module name cannot be fetched from setup.py
- Bug fix - Remove repo from the path inside the build-info
- Bug fix - Removed unrelated options from the "jf build-scan" command

## 2.9.0 (December 13, 2021)
- Add scanType to build scan and Xray version validation.
- Add fish autocompletion.
- Fix category and hidden commands in help.
- Bug fix - getting OpenShift CLI version for v3.
- Bug fix - audit ignores JFROG_CLI_BUILD_PROJECT env.
- Bug fix - bp ignores JFROG_CLI_BUILD_PROJECT env.
- Bug fix - plugin install command. 
  
## 2.8.3 (December 6, 2021)
- UX improvements to the 'jfrog project init' command

## 2.8.2 (December 5, 2021)
- Bug fix - The JFROG_CLI_BUILD_PROJECT environment variable is ignored
- UX improvements to the 'jf setup' and 'jfrog project init' commands

## 2.8.1 (December 3, 2021)
- The "jf setup" command now set the newly created server as the default configured server
- The "jf project init" command now also displays the getting started guide
- Bug fix - The JFrog CLI's docker images size increased

## 2.8.0 (November 30, 2021)
- New "jf setup" command
- New "jf project init" command
- New "jf build-scan" command

## 2.7.0 (November 29, 2021)
- "jfrog go" - New --no-fallback option
- Bug fix - Minimum supported maven version validation fails on some operating systems

## 2.6.2 (November 25, 2021)
- All maven commands now validate that maven 3.1.0 or above are used
- Bug fix - "jfrog rt upload" with --ant may include wrong files in some scenarios
- Bug fix - "jfrog rt bp" can fail, if a previous build-info collection action left an empty cache file
- Bug fix - "jfrog rt npm-publish" may fail with some versions of npm
- Bug fix - "jfrog xr audit-mvn" and "jfrog xr audit-gradle" may skip transitive dependencies

## 2.6.1 (November 22, 2021)
- New shorten commands syntax.
- Shorten executable name to jf.
- Move the environment variables list to a new "jf options" command.
- Export default server if no args were passed to "jfrog c export" command.
- Start using the new build-info-go library.
- Bug fix - 'getMavenHome' fails on windows OS.

## 2.5.1 (November 9, 2021)
- The --scan option for the "jfrog rt mvn", "jfrog rt gradle" and "jfrog rt npm" can be now combined with --format option
to control scan output ("table" or "json").
- Bug fix - Release bundle creation error ignored.
- Bug fix - Fail to create build-info with a long build name.
- Bug fix - Release bundle recursive flag ignored.
- Bug fix - Fix npm version parsing command.
- Bug fix - Fails to collect buildinfo vcs information for repository without commits.

## 2.5.0 (October 23, 2021)
- "jfrog rt repo-template" - Support for Alpine repositories
- "jfrog rt repo-template" - Support for providing a project key
- Breaking change - When the --fail-no-op option is used, and no files are affected, the command summary status is now set to "failure" instead of "success"
- JFrog CLI is now built with go 1.17.2
- Bug fix - Avoid returning an error, in case the indexer-app scans a file which is not supported for scanning
- Bug fix - The --scan option for the "jfrog rt mvn", "jfrog rt gradle" and "jfrog rt npm" command may cause some issues to be skipped and not displayed
- Bug fix - The "jfrog rt build-append" command fails when used with the --project option
- Bug fix - When downloading an archive with the --explode option, the directories inside the archive may be extracted as files
- Bug fix - The commands summary may have missing quotes, if the response is empty

## 2.4.1 (October 4, 2021)
- Bug fix - the "jfrog xr audit-go" command alias should be "ago".
- Bug fix - Irelevant option should be removed from the "jfrog xr audit-go" command.  

## 2.4.0 (October 3, 2021)
- New "jfrog xr audit-pip" command.
- New "jfrog xr audit-go" command.
- Support for GPG validation when downloading release bundles using the "jfrog rt download" command.
- OpenShift support - new "jfrog oc start-build" command.
- Improve the "jfrog ci-setup" command with Jenkins - the command now generates a pipeline that uses the Artifactory DSL.
- Bug fix - the --project option is ignored (when used with the --build option) in many Artifactory commands.
- Bug fix - The --sync option of the "jfrog ds rbdel" command doesn't work.
- Bug fix - Uploading file to Artifactory with properties that include special characters can fail.
- Bug fix - The UI build-info link is wrong when publishing the build-info as part of a JFrog project.
- Bug fix - "jfrog xr scan" will not show an error, if a file that isn't supported by Xray is scanned. 

## 2.3.1 (August 31, 2021)
- Sign JFrog CLI's RPM package 

## 2.3.0 (August 28, 2021)
- The --server-id flag has now become optional for all the package managers' config commands. If not provided, the default server ID is used
- The M2_HOME environment variable is no longer mandatory for maven builds
- Bug fix - "jfrog rt npm-publish" may read the wrong package.json and therefore fetch the wrong package name and number
- Bug fix - The indexer-app downloaded by the "jfrog xr audit..." and "jfrog xr scan" commands cannot be used on Windows OS
- Bug fix - "jfrog rt upload" with --archive and --include-dirs may leaves out empty directories

## 2.2.1 (August 17, 2021)
- Bug fix - Error when downloading the indexer-app from Xray

## 2.2.0 (August 9, 2021)
- "jfrog rt mvn" - Support including / excluding deployed artifacts
- "jfrog rt search" - Allow searching in Artifactory by build, even if the build is included in a project
- "jfrog rt upload" - Allow storing symlinks in an archive when uploading it to Artifactory
- "jfrog xr scan", "jfrog xr audit-..." - When downloading the xray-indexer app, get the version from the app itself, and not from Xray
- Bug fix - Gradle builds which use an old version of the Gradle Artifactory Plugin may fail to deploy artifacts
- Bug fix - The build-info URL is incorrect, in case the build name and number include special characters
- Bug fix - SSH authantication with Artifactory cannot be used without a passphrase
- Bug fix - When searching and filtering by the latest build run, the latest build run isn't always returned
- Bug fix - "jfrog rt build-discard" - the --project flag is missing
- Bug fix - npm-publish may fail if package.json has pre/post pack scripts

## 2.1.1 (July 22, 2021)
- Improvements to the table and full response output of the Xray scan and audit commands
- Removed the JFROG_CLI_OUTPUT_COLORS environment variable introduced in v2.1.0
- Bug fix - Usage report is attempted even if Artifactory is not configured
- Bug fix - Xray and Distribution commands wrongly include Artifactory connection details options

## 2.1.0 (July 21, 2021)
- New "jfrog xr scan" command
- New "jfrog xr audit-npm" command
- New "jfrog xr audit-mvn" command
- New "jfrog xr audit-gradle" command
- New --scan option added to the "jfrog rt npm", "jfrog rt mvn" and "jfrog rt gradle" commands
- Bug fix - When using the --detailed-summary option, the returned upload path is incorrect for the "jfrog rt gp" and "jfrog rt mvn" commands
- Bug fix - When using the --detailed-summary option, there are additional log messages added to stdout, making it impossible to parse the summary

## 2.0.1 (July 4, 2021)
- Fix 'npm install -g jfrog-cli-v2'

## 2.0.0 (July 4, 2021)
- The default value of the --flat option is now set to false for the "jfrog rt upload" command.
- The deprecated syntax of the "jfrog rt mvn" command is no longer supported. To use the new syntax, the project needs to be first configured using the "jfrog rt mvnc" command.
- The deprecated syntax of the "jfrog rt gradle" command is no longer supported. To use the new syntax, the project needs to be first configured using the "jfrog rt gradlec" command.
- The deprecated syntax of the "jfrog rt npm" and "jfrog rt npm-ci" commands is no longer supported. To use the new syntax, the project needs to be first configured using the "jfrog rt npmc" command.
- The deprecated syntax of the "jfrog rt go" command is no longer supported. To use the new syntax, the project needs to be first configured using the "jfrog rt go-config" command.
- The deprecated syntax of the "jfrog rt nuget" command is no longer supported. To use the new syntax, the project needs to be first configured using the "jfrog rt nugetc" command.
- All Bintray commands are removed.
- The "jfrog rt config" command is removed and replaced by the "jfrog config add" command.
- The "jfrog rt use" command is removed and replaced with the "jfrog config use".
- The "props" command option and "Props" file spec property for the "jfrog rt upload" command are removed, and replaced with the "target-props" command option and "targetProps" file spec property respectively.
- The following commands are removed:

   * jfrog rt release-bundle-create
   * jfrog rt release-bundle-delete
   * jfrog rt release-bundle-distribute
   * jfrog rt release-bundle-sign
   * jfrog rt release-bundle-update

   are replaced with the following commands respectively:
   
   * jfrog ds release-bundle-create
   * jfrog ds release-bundle-delete
   * jfrog ds release-bundle-distribute
   * jfrog ds release-bundle-sign
   * jfrog ds release-bundle-update

- The "jfrog rt go-publish" command now only supports Artifactory version 6.10.0 and above. Also, the command no longer accepts the target repository as an argument. The target repository should be pre-configured using the "jfrog rt go-config-command".
- The "jfrog rt go" command no longer falls back to the VCS when dependencies are not found in Artifactory.
- The --deps, --publish-deps, --no-registry and --self options of the "jfrog rt go-publish" command are now removed.
- The API key option is now removed. The API key should now be passed as the value of the password option.
- The --exclude-patterns option is now removed, and replaced with the --exclusions option. The same is true for the excludePatterns file spec property, which is replaced with the exclusions property.
- The JFROG_CLI_JCENTER_REMOTE_SERVER and JFROG_CLI_JCENTER_REMOTE_REPO environment variables are now removed and replaced with the JFROG_CLI_EXTRACTORS_REMOTE environment variable.
- The JFROG_CLI_HOME environment variable is now removed and replaced with the JFROG_CLI_HOME_DIR environment variable.
- The JFROG_CLI_OFFER_CONFIG environment variable is now removed and replaced with the CI environment variable. Setting CI to true disables all prompts.
- The directory structure is now changed when the "jfrog rt download" command is used with placeholders and --flat=false (--flat=false is now the default). When placeholders are used, the value of the --flat option is ignored.

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
- New --list-download option added to the "jfrog bt u" command.
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
