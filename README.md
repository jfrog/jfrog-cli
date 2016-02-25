# JFrog CLI

*JFrog CLI* provides a command line interface for invoking actions on *JFrog Artifactory* and *JFrog Bintray*.

## Getting Started

### Downloading the executables from Bintray

[TO DO]

### Building the command line executable

If you prefer, you may instead build the client in go.

#### Setup GO on your machine

* Make sure you have a working Go environment. [See the install instructions](http://golang.org/doc/install).
* Navigate to the directory where you want to create the *bintray-cli-go* project.
* Set the value of the GOPATH environment variable to the full path of this directory.

#### Download JFrog CLI from GitHub

Run the following command to create the *bintray-cli-go* project:
```console
$ go get github.com/jfrogdev/jfrog-cli-go/...
```

Navigate to the following directory
```console
$ cd $GOPATH/bin
```
### JFrog CLI Usage

You can copy the *frog* executable to any location on your file-system as long as you add it to your *PATH* environment variable,
so that you can access it from any path.

#### Commands structure
JFrog CLI Commands have the following structure:
1. *frog* followed by either *art* for *JFrog Artifactory* commands or *bt* for *JFrog Bintray* commands.
2. The command name.
3. Global options.
4. Command options.    
```console
$ jfrog [art | bt] command-name global-options command-options arguments
```
To display the list of available commands, run *jfrog artii* or *frog bt*.

The sections below describe the available commands, their arguments and respective options.
- [JFrog Artifactory commands](#jfrog-artifactory-commands)
- [JFrog Bintray commands](#jfrog-bintray-commands)

### Security - JFrog Artifactory
JFrog CLI lets you authenticate yourself to Artifactory either using your Artifactory user and password, or using RSA public and private keys.

#### Authentication with user and password
To use your Artifactory user and password, simply use the *--user* and *--password* command options 
as described below or configure your user and password using the [config](#a-config) command.

#### Authentication using RSA public and private keys
Artifactory supports SSH authentication using RSA public and private keys from version 4.4. 
To authenticate yourself to Artifactory using RSA keys, execute the following instructions:

* Enable SSH authentication in Artifactory as described in the [Artifactory Documentation](https://www.jfrog.com/confluence/display/RTF/SSH+Integration).
* Configure your Artifactory URL to have the following format: *ssh://[host]:[port]* using the *--url* command option or the [config](#a-config) command. Please make sure the [host] URL section does not include your Artifactory context URL, but only the host name or IP.
* Configure the path to your private SSH key file using the *--ssh-key-path* command option or the [config](#a-config) command.

<a name="jfrog-artifactory-commands"/>
### JFrog Artifactory commands

#### Global options

Global options are used for all commands.
It is recommended to use the [config](#a-config) command, so that you don't have to add the --url, --user and --password for each command. 
```console
   --url          [Mandatory] Artifactory URL.
   --user         [Mandatory] Artifactory user.
   --password     [Mandatory] Artifactory password.
```   

#### Commands list
- [config (c)](#a-config)
- [upload (u)](#a-upload)
- [download (dl)](#a-download)

<a name="a-config"/>
#### *config* (c) command

##### Function
Used to configure the Artifactory URL and authentication details, so that you don't have to send them as options
for the *upload* and *download* commands.
The configuration is saved at ~/.jfrog/jfrog-cli.conf

##### Command options
```console
   --interactive  [Default: true] Set to false if you do not wish the config command to be interactive. If true, the --url option becomes optional.
   --enc-password [Default: true] If set to false then the configured password will not be encrypted using Artifatory's encryption API.
   --url          [Optional] Artifactory URL.
   --user         [Optional] Artifactory user.
   --password     [Optional] Artifactory password.
```

##### Arguments
* If no arguments are sent, the command will configure the Artifactory URL, user and password sent through the command options
or through the command's interactive prompt.
* The *show* argument will make the command show the stored configuration.
* The *clear* argument will make the command clear the stored configuration.

###### Important Note

if your Artifactory server has [encrypted password set to required](https://www.jfrog.com/confluence/display/RTF/Configuring+Security#ConfiguringSecurity-PasswordEncryptionPolicy) you should use your API Key as your password.

##### Examples

Configure the Artifactory details through an interactive propmp.
```console
$  jfrog arti c
```

Configure the Artifactory details through the command options.

```console
$  jfrog arti c 
```

Show the configured Artifactory details.
```console
$  jfrog arti c show
```

Clear the configured Artifactory details.
```console
$  jfrog arti c clear
```

<a name="a-upload"/>
#### *upload* (u) command

##### Function
Used to upload artifacts to Artifactory.

##### Command options
The command uses the global options, in addition to the following command options.
```console
   --props        [Optional] List of properties in the form of key1=value1;key2=value2,... to be attached to the uploaded artifacts.
   --deb          [Optional] Used for Debian packages in the form of distribution/component/architecture.
   --flat         [Default: true] If not set to true, and the upload path ends with a slash, artifacts are uploaded according to their file system hierarchy.
   --recursive    [Default: true] Set to false if you do not wish to collect artifacts in sub-folders to be uploaded to Artifactory.
   --regexp       [Default: false] Set to true to use a regular expression instead of wildcards expression to collect artifacts to upload.
   --threads      [Default: 3] Number of artifacts to upload in parallel.
   --dry-run      [Default: false] Set to true to disable communication with Artifactory.
```
##### Arguments
* The first argument is the local file-system path to the artifacts to be uploaded to Artifactory.
The path can include a single file or multiple artifacts, by using the * wildcard.
**Important:** If the path is provided as a regular expression (with the --regexp=true option) then
the first regular expression appearing as part of the argument must be enclosed in parenthesis.

* The second argument is the upload path in Artifactory.
The argument should have the following format: [repository name]/[repository path]
The path can include symbols in the form of {1}, {2}, ...
These symbols are replaced with the sections enclosed with parenthesis in the first argument.

##### Examples

This example uploads the *froggy.tgz* file to the root of the *my-local-repo* repository
```console
$  jfrog arti u froggy.tgz my-local-repo/
```

This example collects all the zip artifacts located under the build directory (including sub-directories).
and uploads them to the *my-local-repo* repository, under the zipFiles folder, while keeping the artifacts original names.
```console
$  jfrog arti u build/*.zip libs-release-local/zipFiles/ 
```
And on Windows:
```console
$  jfrog arti u build\\*.zip libs-release-local/zipFiles/ 
```

<a name="a-download"/>
##### *download* (dl) command

##### Function
Used to download artifacts from Artifactory.

##### Command options
The command uses the global options, in addition to the following command options.
```console
   --props        [Optional] List of properties in the form of key1=value1;key2=value2,... Only artifacts with these properties will be downloaded.
   --flat         [Default: false] Set to true if you do not wish to have the Artifactory repository path structure created locally for your downloaded artifacts
   --recursive    [Default: true] Set to false if you do not wish to include the download of artifacts inside sub-directories in Artifactory.
   --min-split    [Default: 5120] Minimum file size in KB to split into ranges. Set to -1 for no splits.
   --split-count  [Default: 3] Number of parts to split a file when downloading. Set to 0 for no splits.
   --threads      [Default: 3] Number of artifacts to download in parallel.
```

##### Arguments
The command expects one argument - the path of artifacts to be downloaded from Artifactory.
The argument should have the following format: [repository name]/[repository path]
The path can include a single artifact or multiple artifacts, by using the * wildcard.
The artifacts are downloaded and saved to the current directory, while saving their folder structure.

##### Examples

This example downloads the *cool-froggy.zip* artifact located at the root of the *my-local-repo* repository to current directory.
```console
$  jfrog arti dl my-local-repo/cool-froggy.zip 
```

This example downloads all artifacts located in the *my-local-repo* repository under the *all-my-frogs* folder to the *all-my-frog* directory located unde the current directory.
```console
$  jfrog arti dl my-local-repo/all-my-frogs/ 
```

<a name="jfrog-bintray-commands"/>
### JFrog Bintray commands

#### Global options

Global options are used for all commands.
It is recommended to use the [config](#config) command, so that you don't have to add the --user and --key for each command.
```console
   --user            [Optional] Bintray username. It can be also set using the BINTRAY_USER environment variable. If not set, the subject sent as part of the command argument is used for authentication.
   --key             [Mandatory] Bintray API key. It can be also set using the BINTRAY_KEY environment variable.
```

#### Commands list
- [config (c)](#config)
- [upload (u)](#upload)
- [download-file (dlf)](#download-file)
- [download-ver (dlv)](#download-ver)
- [package-show (ps)](#package-show)
- [package-create (pc)](#package-create)
- [package-update (pc)](#package-update)
- [package-delete (pd)](#package-delete)
- [version-show (vs)](#version-show)
- [version-create (vc)](#version-create)
- [version-update (vu)](#version-update)
- [version-delete (vd)](#version-delete)
- [version-publish (vd)](#version-publish)
- [access-keys (acc-keys)](#access-keys)
- [entitlements (ent)](#entitlements)
- [sign-url (su)](#sign-url)
- [gpg-sign-file (gsf)](#gpg-sign-file)
- [gpg-sign-ver (gsv)](#gpg-sign-ver)

<a name="config"/>
#### *config* (c) command

##### Function
Used to configure your Bintray user and API key, so that you don't have to send them as command options.
The configuration is saved at ~/.jfrog/jfrog-cli.conf

##### Command options
```console
   --interactive  [Default: true] Set to false if you do not wish the config command to be interactive.
   --user         [Optional] Bintray user.
   --key          [Optional] Bintray key.
   --licenses     [Optional] Default package licenses in the form of Apache-2.0,GPL-3.0...
```

##### Arguments
* If no arguments are sent, the command will configure your user and API Key sent through the command options
or through the command's interactive prompt.
* The *show* argument will make the command show the stored configuration.
* The *clear* argument will make the command clear the stored configuration.

##### Examples

Configure user and API Key through an interactive propmp.
```console
$  jfrog bt c
```

Configure user and API Key through the command options.
```console
$  jfrog bt c --user=my-user --key=mybintrayapikey
```

Show the configured user and API Key.
```console
$  jfrog bt c show
```

Clear the configured user and API Key.
```console
$  jfrog bt c clear
```

<a name="upload"/>
#### *upload* (u) command

##### Function
Used to upload files to Bintray

##### Command options
The command uses the global options, in addition to the following command options.
```console
   --flat              [Default: true]   If not set to true, and the upload path ends with a slash, artifacts are uploaded according to their file system hierarchy.
   --recursive         [Default: true]   Set to false if you do not wish to collect artifacts in sub-directories to be uploaded to Bintray.
   --regexp            [Default: false]  Set to true to use a regular expression instead of wildcards expression to collect artifacts to upload.
   --publish           [Default: false]  Set to true to publish the uploaded files.
   --override          [Default: false]  Set to true to enable overriding existing published files.
   --explode           [Default: false]  Set to true to explode archived files after upload.
   --threads           [Default: 3]      Number of artifacts to upload in parallel.   
   --dry-run           [Default: false]  Set to true to disable communication with Bintray.
   --deb               [Optional] Used for Debian packages in the form of distribution/component/architecture.   
```
If the Bintray Package to which you're uploading does not exist, the CLI will try to create it.
Please send the following command options for the package creation.
```console
   --pkg-desc                    [Optional]        Package description.
   --pkg-labels                  [Optional]        Package lables in the form of lable11,lable2...
   --pkg-licenses                [Mandatory]       Package licenses in the form of Apache-2.0,GPL-3.0...
   --pkg-cust-licenses           [Optional]        Package custom licenses in the form of my-license-1,my-license-2...
   --pkg-vcs-url                 [Mandatory]       Package VCS URL.
   --pkg-website-url             [Optional]        Package web site URL.
   --pkg-issuetracker-url           [Optional]        Package Issues Tracker URL.
   --pkg-github-repo             [Optional]        Package Github repository.
   --pkg-github-rel-notes        [Optional]        Github release notes file.
   --pkg-pub-dn                  [Default: false]  Public download numbers.
   --pkg-pub-stats               [Default: false]  Public statistics
```
If the Package Version to which you're uploading does not exist, the CLI will try to create it.
Please send the following command options for the version creation.   
```console
   --ver-desc                    [Optional]   Version description.
   --ver-vcs-tag                 [Optional]   VCS tag.
   --ver-released                [Optional]   Release date in ISO8601 format (yyyy-MM-dd'T'HH:mm:ss.SSSZ).
   --ver-github-rel-notes        [Optional]   Github release notes file.
   --ver-github-tag-rel-notes    [Optional]   Set to true if you wish to use a Github tag release notes.
```

##### Arguments
* The first argument is the local file-system path to the artifacts to be uploaded to Bintray.
The path can include a single file or multiple artifacts, by using the * wildcard.
**Important:** If the path is provided as a regular expression (with the --regexp=true option) then
the first regular expression appearing as part of the argument must be enclosed in parenthesis.

* The second argument is the upload path in Bintray.
The path can include symbols in the form of {1}, {2}, ...
These symbols are replaced with the sections enclosed with parenthesis in the first argument.

##### Examples

Upload all files located under *dir/sub-dir*, with names that start with *frog*, to the root path under version *1.0* 
of the *froggy-package* package 
```console
frog bt u dir/sub-dir/frog* my-org/swamp-repo/froggy-package/1.0/ 
```

Upload all files located under *dir/sub-dir*, with names that start with *frog* to the /frog-files folder, under version *1.0* 
of the *froggy-package* package 
```console
 jfrog bt u dir/sub-dir/frog* my-org/swamp-repo/froggy-package/1.0/frog-files/ 
```

Upload all files located under *dir/sub-dir* with names that start with *frog* to the root path under version *1.0*, 
while adding the *-up* suffix to their names in Bintray.  
```console
 jfrog bt u dir/sub-dir/(frog*) my-org/swamp-repo/froggy-package/1.0/{1}-up 
```
<a name="download-file"/>
#### *download-file* (dlf) command

##### Function
Used to download a specific file from Bintray.

##### Command options
The command uses the global options, in addition to the following command option.
```console   
   --flat         [Default: false]  Set to true if you do not wish to have the Bintray path structure created locally for your downloaded file.
   --min-split    [Default: 5120]   Minimum file size in KB to split into ranges. Set to -1 for no splits.
   --split-count  [Default: 3]      Number of parts to split a file when downloading. Set to 0 for no splits.      
```

##### Arguments
The command expects one argument in the form of *subject/repository/package/version/path*.

##### Example
```console 
 jfrog bt dlf my-org/swamp-repo/froggy-package/1.0/com/jfrog/bintray/crazy-frog.zip 
```

<a name="download-ver"/>
#### *download-ver* (dlv) command

##### Function
Used to download the files of a specific version from Bintray.

##### Command options
The command uses the global options, in addition to the following command options.
```console
   --flat         [Default: false]  Set to true if you do not wish to have the Bintray path structure created locally for your downloaded files.   
   --min-split    [Default: 5120] Minimum file size in KB to split into ranges. Set to -1 for no splits.
   --split-count  [Default: 3] Number of parts to split a file when downloading. Set to 0 for no splits.
   --threads      [Default: 3] Number of artifacts to download in parallel.   
```

##### Arguments
The command expects one argument in the form of *subject/repository/package/version*.

##### Example
```console 
 jfrog bt dlv my-org/swamp-repo/froggy-package/1.0 
```

<a name="package-show"/>
#### *package-show* (ps) command

##### Function
Used for showing package details.

##### Command options
This command has no command options. It uses however the global options.

##### Arguments
The command expects one argument in the form of *subject/repository/package.

##### Examples
Show package *super-frog-package* 
```console
 jfrog bt ps my-org/swamp-repo/super-frog-package 
```

<a name="package-create"/>
#### *package-create* (pc) command

##### Function
Used for creating a package in Bintray

##### Command options
The command uses the global options, in addition to the following command options.
```console
   --desc               [Optional]        Package description.
   --labels             [Optional]        Package lables in the form of lable11,lable2...
   --licenses           [Mandatory]       Package licenses in the form of Apache-2.0,GPL-3.0...
   --cust-licenses      [Optional]        Package custom licenses in the form of my-license-1,my-license-2...
   --vcs-url            [Mandatory]       Package VCS URL.
   --website-url        [Optional]        Package web site URL.
   --issuetracker-url   [Optional]        Package Issues Tracker URL.
   --github-repo        [Optional]        Package Github repository.
   --github-rel-notes   [Optional]        Github release notes file.
   --pub-dn             [Default: false]  Public download numbers.
   --pub-stats          [Default: true]   Public statistics
```

##### Arguments
The command expects one argument in the form of *subject/repository/package.

##### Examples
Create the *super-frog-package* package 
```console
 jfrog bt pc my-org/swamp-repo/super-frog-package --licenses=Apache-2.0,GPL-3.0 --vcs-url=http://github.com/jFrogdev/coolfrog.git 
```

<a name="package-update"/>
#### *package-update* (pu) command

##### Function
Used for updating package details in Bintray

##### Command options
The command uses the same option as the *package-create* command

##### Arguments
The command expects one argument in the form of *subject/repository/package.

##### Examples
Create the *super-frog-package* package 
```console
 jfrog bt pu my-org/swamp-repo/super-frog-package --labels=label1,label2,label3 
```

<a name="package-delete"/>
#### *package-delete* (pd) command

##### Function
Used for deleting a packages in Bintray

##### Command options
The command uses the global options, in addition to the following command option.
```console
   --quiet      [Default: false]       Set to true to skip the delete confirmation message.
```

##### Arguments
The command expects one argument in the form of *subject/repository/package.

##### Examples
Delete the *froger-package* package 
```console
 jfrog bt pc my-org/swamp-repo/froger-package --licenses=Apache-2.0,GPL-3.0 --vcs-url=http://github.com/jFrogdev/coolfrog.git 
```

<a name="version-show"/>
#### *version-show* (vs) command

##### Function
Used for showing version details.

##### Command options
This command has no command options. It uses however the global options.

##### Arguments
The command expects one argument in one of the forms
* *subject/repository/package* to show the latest published version.
* *subject/repository/package/version* to show the specified version.

##### Examples
Show version 1.0.0 of package *super-frog-package* 
```console
 jfrog bt vs my-org/swamp-repo/super-frog-package/1.0 
```

Show the latest published version of package *super-frog-package* 
```console
 jfrog bt vs my-org/swamp-repo/super-frog-package 
```

<a name="version-create"/>
#### *version-create* (vc) command

##### Function
Used for creating a version in Bintray

##### Command options
The command uses the global options, in addition to the following command options.
```console
   --desc                    [Optional]   Version description.
   --vcs-tag                 [Optional]   VCS tag.
   --released                [Optional]   Release date in ISO8601 format (yyyy-MM-dd'T'HH:mm:ss.SSSZ).
   --github-rel-notes        [Optional]   Github release notes file.
   --github-tag-rel-notes    [Optional]   Set to true if you wish to use a Github tag release notes.
```

##### Arguments
The command expects one argument in the form of *subject/repository/package/version.

##### Examples
Create version 1.0.0 in package *super-frog-package* 
```console
 jfrog bt vc my-org/swamp-repo/super-frog-package/1.0.0 
```

<a name="version-update"/>
#### *version-update* (vu) command

##### Function
Used for updating version details in Bintray

##### Command options
The command uses the same option as the *version-create* command

##### Arguments
The command expects one argument in the form of *subject/repository/package/version.

##### Examples
Update the labels of version 1.0.0 in package *super-frog-package* 
```console
 jfrog bt vu my-org/swamp-repo/super-frog-package/1.0.0 --labels=jump,jumping,frog
```

<a name="version-delete"/>
#### *version-delete* (vd) command

##### Function
Used for deleting a version in Bintray

##### Command options
The command uses the global options, in addition to the following command option.
```console
   --quiet      [Default: false]       Set to true to skip the delete confirmation message.
```

##### Arguments
The command expects one argument in the form of *subject/repository/package/version.

##### Examples
Delete version 1.0.0 in package *super-frog-package* 
```console
 jfrog bt vd my-org/swamp-repo/super-frog-package/1.0.0 
```

<a name="version-publish"/>
#### *version-publish* (vp) command

##### Function
Used for publishing a version in Bintray

##### Command options
This command has no command options. It uses however the global options.

##### Arguments
The command expects one argument in the form of *subject/repository/package/version.

##### Examples
Publish version 1.0.0 in package *super-frog-package* 
```console
 jfrog bt vp my-org/swamp-repo/super-frog-package/1.0.0 
```

<a name="access-keys"/>
#### *access-keys* (acc-keys) command

##### Function
Used for managing Entitlement Access Keys.

##### Command options
The command uses the global options, in addition to the following command options.
```console
   --password        [Optional]     Access Key password..
   --org             [Optional]     Bintray organization.
   --expiry          [Optional]     Access Key expiry in milliseconds, in the form of Unix epoch time (required for 'jfrog bt acc-keys show/create/update/delete'
   --ex-check-url    [Optional]     Used for Access Key creation and update. You can optionally provide an existence check directive, in the form of a callback URL, to verify whether the source identity of the Access Key still exists.
   --ex-check-cache  [Optional]     Used for Access Key creation and update. You can optionally provide the period in seconds for the callback URK results cache.
   --white-cidrs     [Optional]     Used for Access Key creation and update. Specifying white CIDRs in the form of 127.0.0.1/22,193.5.0.1/22 will allow access only for those IPs that exist in that address range.
   --black-cidrs     [Optional]     Used for Access Key creation and update. Specifying black CIDRs in the foem of 127.0.0.1/22,193.5.0.1/22 will block access for all IPs that exist in the specified range.
```

##### Arguments
* With no arguments, a list of all Access Keys is displayed.
* When sending the show, create, update or delete argument, it should be followed by an argument indicating the Access Key ID to show, create, update or delete the Access Key.

##### Examples
Show all Access Keys
```console
 jfrog bt acc-keys
```
Create a Access Key
```console
 jfrog bt acc-keys create key1 
 jfrog bt acc-keys create key1 --expiry=7956915742000 
```

Show a specific Access Key
```console
 jfrog bt acc-keys show key1
```

Update a Access Key
```console
 jfrog bt acc-keys update key1 --ex-check-url=http://new-callback.com --white-cidrs=127.0.0.1/22,193.5.0.1/92 --black-cidrs=127.0.0.1/22,193.5.0.1/92
 jfrog bt acc-keys update key1 --expiry=7956915752000
```

Delete a Access Key
```console
 jfrog bt acc-keys delete key1
```

<a name="entitlements"/>
#### *entitlements* (ent) command

##### Function
Used for managing Entitlements.

##### Command options
The command uses the global options, in addition to the following command options.
` ``console
   --id              Entitlement ID. Used for Entitlements update.
   --access          Entitlement access. Used for Entitlements creation and update.
   --keys            Used for Entitlements creation and update. List of Access Keys in the form of key1,key2...
   --path            Entitlement path. Used for Entitlements creating and update.
```

##### Arguments
* When sending an argument in the form of subject/repo or subject/repo/package or subject/repo/package/version, a list of all entitles for the send repo, package or version is displayed.
* If the first argument sent is show, create, update or delete, it should be followed by an argument in the form of subject/repo or subject/repo/package or subject/repo/package/version to show, create, update or delete the entitlement.

##### Examples

Show all Entitlements of the swamp-repo repository.
```console
 jfrog bt ent my-org/swamp-repo
```

Show all Entitlements of the green-frog package.
```console
 jfrog bt ent my-org/swamp-repo/green-frog
```

Show all Entitlements of version 1.0 of the green-frog package.
```console
 jfrog bt ent my-org/swamp-repo/green-frog/1.0
```

Create an Entitlement for the green-frog package, with rw access, the key1 and key2 Access Keys and the a/b/c path.
```console
 jfrog bt ent create my-org/swamp-repo/green-frog --access=rw --keys=key1,key2 --path=a/b/c
```

Show a specific Entitlement on the swamp-repo repository.
```console
 jfrog bt ent show my-org/swamp-repo --id=451433e7b3ec3f18110ba770c77b9a3cb5534cfc
```

Update the Access Keys and access of an Entitlement on the swamp-repo repository.
```console
 jfrog bt ent update my-org/swamp-repo --id=451433e7b3ec3f18110ba770c77b9a3cb5534cfc --keys=key1,key2 --access=r
```

Delete an Entitlement on the my-org/swamp-repo.
```console 
 jfrog bt ent delete my-org/swamp-repo --id=451433e7b3ec3f18110ba770c77b9a3cb5534cfc
```

<a name="sign-url"/>
#### *sign-url* (su) command

##### Function
Used for Generating an anonymous, signed download URL with an expiry date.

##### Command options
The command uses the global options, in addition to the following command options.
```console
   --expiry            [Optional]    An expiry date for the URL, in Unix epoch time in milliseconds, after which the URL will be invalid. By default, expiry date will be 24 hours.
   --valid-for         [Optional]    The number of seconds since generation before the URL expires. Mutually exclusive with the --expiry option.
   --callback-id       [Optional]    An applicative identifier for the request. This identifier appears in download logs and is used in email and download webhook notifications.
   --callback-email    [Optional]    An email address to send mail to when a user has used the download URL. This requiers a callback_id. The callback-id will be included in the mail message.
   --callback-url      [Optional]    A webhook URL to call when a user has used the download URL.
   --callback-method   [Optional]    HTTP method to use for making the callback. Will use POST by default. Supported methods are: GET, POST, PUT and HEAD.
```

##### Arguments
The command expects one argument in the form of *subject/repository/file-path*.

##### Examples
Create a download URL for *froggy-file*, located under *froggy-folder* in the *swamp-repo* repository.
 ```console
 jfrog bt us my-org/swamp-repo/froggy-folder/froggy-file
```

<a name="gpg-sign-file"/>
#### *gpg-sign-file* (gsf) command

##### Function
GPG sign a file in Bintray.

##### Command options
The command uses the global options, in addition to the following command option.
```console
   --passphrase      [Optional]    GPG passphrase.
```

##### Arguments
The command expects one argument in the form of *subject/repository/file-path*.

##### Examples
GPG sign the *froggy-file* file, located under *froggy-folder* in the *swamp-repo* repository.
 ```console
 jfrog bt gsf my-org/swamp-repo/froggy-folder/froggy-file
```
or with a passphrase
```console
frog bt gsf my-org/swamp-repo/froggy-folder/froggy-file --passphrase=gpgX***yH8eKw
```

<a name="gpg-sign-ver"/>
#### *gpg-sign-ver* (gsv) command

##### Function
GPS sign all files of a specific version in Bintray.

##### Command options
The command uses the global options, in addition to the following command option.
```console
   --passphrase      [Optional]    GPG passphrase.
```

##### Arguments
The command expects one argument in the form of *subject/repository/package/version*.

##### Examples
GPG sign all files of version *1.0* of the *froggy-package* package in the *swamp-repo* repository.
 ```console
 jfrog bt gsv my-org/swamp-repo/froggy-package/1.0
```
or with a passphrase
```console
 jfrog bt gsv my-org/swamp-repo/froggy-package/1.0 --passphrase=gpgX***yH8eKw
```