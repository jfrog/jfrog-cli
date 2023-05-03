# JFrog CLI 

JFrog CLI is a compact and smart client that provides a simple interface that automates access to JFrog products simplifying your automation scripts making them more readable and easier to maintain. JFrog CLI works with JFrog Artifactory, Xray, Distribution, Pipelines and Mission Control, (through their respective REST APIs) making your scripts more efficient and reliable in several ways:

## Parallel uploads and downloads

JFrog CLI allows you to upload and download artifacts concurrently by a configurable number of threads that help your automated builds run faster. For big artifacts, you can define a number of chunks to split files for parallel download.

## Checksum optimization

JFrog CLI optimizes both upload and download operations by skipping artifacts that already exist in their target location. Before uploading an artifact, JFrog CLI queries Artifactory with the artifact's checksum. If it already exists in Artifactory's storage, the CLI skips sending the file, and if necessary, Artifactory only updates its database to reflect the artifact upload. Similarly, when downloading an artifact from Artifactory if the artifact already exists in the same download path, it will be skipped. With checksum optimization, long upload and download operations can be paused in the middle, and then be continued later where they were left off.

## Flexible uploads and downloads

JFrog CLI supports uploading files to Artifactory using wildcard patterns, regular expressions and ANT patterns,  giving you an easy way to collect all the files you wish to upload. You can also download files using wildcard patterns.

## Upload and download preview

All upload and download operations can be used with the `--dry-run` option to give you a preview of all the files that would be uploaded with the current command.

Read More

* [CLI for JFrog Artifactory](https://jfrog.com/help/r/jfrog-cli/cli-for-jfrog-artifactory)
* [CLI for JFrog Xray](https://jfrog.com/help/r/jfrog-cli/cli-for-jfrog-xray)
* [CLI for JFrog Mission Control](https://jfrog.com/help/r/jfrog-cli/cli-for-jfrog-mission-control)
* [CLI for JFrog Distribution](https://jfrog.com/help/r/jfrog-cli/cli-for-jfrog-distribution)

* * *

## JFrog CLI v2
### Overview

JFrog CLI v2 was launched in July 2021. It includes changes to the functionality and usage of some of the legacy JFrog CLI commands. The changes are the result of feedback we received from users over time through GitHub, making the usage and functionality easier and more intuitive. For example, some of the default values changed, and are now more consistent across different commands. We also took this opportunity for improving and restructuring the code, as well as replacing old and deprecated functionality.

Most of the changes included in v2 are breaking changes compared to the v1 releases. We therefore packaged and released these changes under JFrog CLI v2, allowing users to migrate to v2 only when they are ready.

New enhancements to JFrog CLI are planned to be introduced as part of V2 only. V1 receives very little development attention nowadays. We therefore encourage users who haven't yet migrated to V2, to do so.

### List of changes in JFrog CLI v2

1.  The default value of the _**--flat**_ option is now set to false for the _**jfrog rt upload**_ command.
2.  The deprecated syntax of the _**jfrog rt mvn**_ command is no longer supported. To use the new syntax, the project needs to be first configured using the **jfrog rt mvnc** command.
3.  The deprecated syntax of the _**jfrog rt gradle**_ command is no longer supported. To use the new syntax, the project needs to be first configured using the _**jfrog rt gradlec**_ command.
4.  The deprecated syntax of the **jfrog rt npm** and _**jfrog rt npm-ci**_ commands is no longer supported. To use the new syntax, the project needs to be first configured using the _**jfrog rt npmc**_ command.
5.  The deprecated syntax of the _**jfrog rt go**_ command is no longer supported. To use the new syntax, the project needs to be first configured using the _**jfrog rt go-config**_ command.
6.  The deprecated syntax of the _**jfrog rt nuget**_ command is no longer supported. To use the new syntax, the project needs to be first configured using the _**jfrog rt nugetc**_ command.
7.  All Bintray commands are removed.
8.  The _**jfrog rt config**_ command is removed and replaced by the _**jfrog config add**_ command.
9.  The _**jfrog rt use**_ command is removed and replaced with the _**jfrog config use**_.
10. The _**--props**_ command option and _**props**_ file spec property for the _**jfrog rt upload**_ command are removed, and replaced with the _**--target-props**_ command option and _**targetProps**_ file spec property respectively.
11. The following commands are removed 
    ```
    jfrog rt release-bundle-create
    jfrog rt release-bundle-delete
    jfrog rt release-bundle-distribute
    jfrog rt release-bundle-sign
    jfrog rt release-bundle-update
    ```
    and replaced with the following commands respectively 
    ```
    jfrog ds release-bundle-create
    jfrog ds release-bundle-delete
    jfrog ds release-bundle-distribute
    jfrog ds release-bundle-sign
    jfrog ds release-bundle-update
    ```
12. The _**jfrog rt go-publish**_ command now only supports Artifactory version 6.10.0 and above. Also, the command no longer accepts the target repository as an argument. The target repository should be pre-configured using the _**jfrog rt go-config**_ command.
13. The _**jfrog rt go**_ command no longer falls back to the VCS when dependencies are not found in Artifactory.
14. The _**--deps**_, _**--publish-deps**_, _**--no-registry**_ and _**--self**_ options of the _**jfrog rt go-publish**_ command are now removed.
15. The _**--apiKey**_ option is now removed. The API key should now be passed as the value of the _**--password**_ option.
16. The _**--exclude-patterns**_ option is now removed, and replaced with the _**--exclusions**_ option. The same is true for the _**excludePatterns**_ file spec property, which is replaced with the _**exclusions**_ property.
17. The _**JFROG\_CLI\_JCENTER\_REMOTE\_SERVER**_ and _**JFROG\_CLI\_JCENTER\_REMOTE\_REPO**_ environment variables are now removed and replaced with the _**JFROG\_CLI\_EXTRACTORS_REMOTE**_ environment variable.
18. The _**JFROG\_CLI\_HOME**_ environment variable is now removed and replaced with the _**JFROG\_CLI\_HOME_DIR**_ environment variable.
19. The _**JFROG\_CLI\_OFFER_CONFIG**_ environment variable is now removed and replaced with the _**CI**_ environment variable. Setting CI to true disables all prompts.
20. The directory structure is now changed when the _**jfrog rt download**_ command is used with placeholders and -_**-flat=false**_ (--flat=false is now the default). When placeholders are used, the value of the _**--flat**_ option is ignored.
21. When the **jfrog rt upload** command now uploads symlinks to Atyifctory, the target file referenced by the symlink is uploaded to Artifactory with the symlink name. If the **--symlink** options is used, the symlink itself (not the referenced file) is uploaded, with the referenced file as a property attached to the file.

  

## Download and installation
### General

To download the executable, please visit the  [JFrog CLI Download Site](https://www.jfrog.com/getcli/).

You can also download the sources from the  [JFrog CLI Project](https://github.com/JFrog/jfrog-cli-go) on GitHub where you will also find instructions on how to build JFrog CLI.

The legacy name of JFrog CLI's executable is _**jfrog**_. In an effort to make the CLI usage easier and more convenient, we recently exposed a series of new installers, which install JFrog CLI with the new _**jf**_ executable name. For backward compatibility, the old installers will remain available. We recommend however migrating to the newer _**jf**_ executable name.

### JFrog CLI v2 "jf" installers

The following installers are available for JFrog CLI v2. These installers make JFrog CLI available through the _**jf**_ executable.

**Debian**
```
wget -qO - https://releases.jfrog.io/artifactory/jfrog-gpg-public/jfrog\_public\_gpg.key | sudo apt-key add -
echo "deb https://releases.jfrog.io/artifactory/jfrog-debs xenial contrib" | sudo tee -a /etc/apt/sources.list;
apt update;
apt install -y jfrog-cli-v2-jf;
```

**RPM**
```
echo "\[jfrog-cli\]" > jfrog-cli.repo;
echo "name=jfrog-cli" >> jfrog-cli.repo;
echo "baseurl=https://releases.jfrog.io/artifactory/jfrog-rpms" >> jfrog-cli.repo;
echo "enabled=1" >> jfrog-cli.repo;
rpm --import https://releases.jfrog.io/artifactory/jfrog-gpg-public/jfrog\_public\_gpg.key
sudo mv jfrog-cli.repo /etc/yum.repos.d/;
yum install -y jfrog-cli-v2-jf;
```

**Homebrew**
```
brew install jfrog-cli
```

**Install with cUrl**
```
curl -fL https://install-cli.jfrog.io | sh
```

**Download with cUrl**
```
curl -fL https://getcli.jfrog.io/v2-jf | sh
```

**NPM**
```
npm install -g jfrog-cli-v2-jf
```

**Docker**
```
Slim:
docker run releases-docker.jfrog.io/jfrog/jfrog-cli-v2-jf jf -v

Full:
docker run releases-docker.jfrog.io/jfrog/jfrog-cli-full-v2-jf jf -v
```

**Powershell**
```
powershell "Start-Process -Wait -Verb RunAs powershell '-NoProfile iwr https://releases.jfrog.io/artifactory/jfrog-cli/v2-jf/\[RELEASE\]/jfrog-cli-windows-amd64/jf.exe -OutFile $env:SYSTEMROOT\\system32\\jf.exe'"
```

**Chocolatey**
```
choco install jfrog-cli-v2-jf
```
  

### JFrog CLI v2 "jfrog" installers

The following installers are available for JFrog CLI v2. These installers make JFrog CLI available through the _**jfrog**_ executable.

**Debian**
```
wget -qO - https://releases.jfrog.io/artifactory/jfrog-gpg-public/jfrog\_public\_gpg.key | sudo apt-key add -
echo "deb https://releases.jfrog.io/artifactory/jfrog-debs xenial contrib" | sudo tee -a /etc/apt/sources.list;
apt update;
apt install -y jfrog-cli-v2;
```

**RPM**
```
echo "\[jfrog-cli\]" > jfrog-cli.repo;
echo "name=jfrog-cli" >> jfrog-cli.repo;
echo "baseurl=https://releases.jfrog.io/artifactory/jfrog-rpms" >> jfrog-cli.repo;
echo "enabled=1" >> jfrog-cli.repo;
rpm --import https://releases.jfrog.io/artifactory/jfrog-gpg-public/jfrog\_public\_gpg.key
sudo mv jfrog-cli.repo /etc/yum.repos.d/;
yum install -y jfrog-cli-v2;
```

**Homebrew**
```
brew install jfrog-cli
```

**Download with Curl**
```
curl -fL https://getcli.jfrog.io/v2 | sh
```

**NPM**
```
npm install -g jfrog-cli-v2
```

**Docker**
```
Slim:
docker run releases-docker.jfrog.io/jfrog/jfrog-cli-v2 jfrog -v

Full:
docker run releases-docker.jfrog.io/jfrog/jfrog-cli-full-v2 jfrog -v
```

**Chocolatey**
```
choco install jfrog-cli
```

### JFrog CLI v1 (legacy) installers

The following installations are available for JFrog CLI v1. These installers make JFrog CLI available through the _**jfrog**_ executable.

**Debian**
```
wget -qO - https://releases.jfrog.io/artifactory/jfrog-gpg-public/jfrog\_public\_gpg.key | sudo apt-key add -
echo "deb https://releases.jfrog.io/artifactory/jfrog-debs xenial contrib" | sudo tee -a /etc/apt/sources.list;
apt update;
apt install -y jfrog-cli;
```

**RPM**
```
echo "\[jfrog-cli\]" > jfrog-cli.repo;
echo "name=jfrog-cli" >> jfrog-cli.repo;
echo "baseurl=https://releases.jfrog.io/artifactory/jfrog-rpms" >> jfrog-cli.repo;
echo "enabled=1" >> jfrog-cli.repo;
rpm --import https://releases.jfrog.io/artifactory/jfrog-gpg-public/jfrog\_public\_gpg.key
sudo mv jfrog-cli.repo /etc/yum.repos.d/;
yum install -y jfrog-cli;
```

**Download with cUrl**
```
curl -fL https://getcli.jfrog.io | sh
```

**NPM**
```
npm install -g jfrog-cli-go
```

**Docker**
```
Slim:
docker run releases-docker.jfrog.io/jfrog/jfrog-cli jfrog -v

Full:
docker run releases-docker.jfrog.io/jfrog/jfrog-cli-full jfrog -v
```

**Go**
```
GO111MODULE=on go get github.com/jfrog/jfrog-cli; 
if \[ -z "$GOPATH" \] 
then binPath="$HOME/go/bin"; 
else binPath="$GOPATH/bin"; 
fi; 
mv "$binPath/jfrog-cli" "$binPath/jfrog"; 
echo "$($binPath/jfrog -v) is installed at $binPath";
```
  
## System Requirements

JFrog CLI runs on any modern OS that fully supports the [Go programming language](https://golang.org/).



## Usage

To use the CLI, [install](https://jfrog.com/getcli/) it on your local machine, or [download](https://jfrog.com/getcli/) its executable, place it anywhere in your file system and add its location to your `PATH` environment variable. 

### Environment Variables

The **jf options** command displays all the supported environment variables.

JFrog CLI makes use of the following environment variables:

|                                |                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
|--------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Variable Name**              | **Description**                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| **JFROG\_CLI\_LOG_LEVEL**      | \[Default: INFO\]<br><br>This variable determines the log level of the JFrog CLI.  <br>Possible values are: INFO, ERROR, and DEBUG.  <br>If set to ERROR, JFrog CLI logs error messages only. It is useful when you wish to read or parse the JFrog CLI output and do not want any other information logged.                                                                                                                                          |
| **JFROG\_CLI\_LOG_TIMESTAMP**  | \[Default: TIME\]<br><br>Controls the log messages timestamp format. Possible values are: TIME, DATE\_AND\_TIME, and OFF.                                                                                                                                                                                                                                                                                                                             |
| **JFROG\_CLI\_HOME_DIR**       | \[Default: ~/.jfrog\]<br><br>Defines the JFrog CLI home directory.                                                                                                                                                                                                                                                                                                                                                                                    |
| **JFROG\_CLI\_TEMP_DIR**       | \[Default: The operating system's temp directory\]<br><br>Defines the temp directory used by JFrog CLI.                                                                                                                                                                                                                                                                                                                                               |
| **JFROG\_CLI\_PLUGINS_SERVER** | \[Default: Official JFrog CLI Plugins registry\]<br><br>Configured Artifactory server ID from which to download JFrog CLI Plugins.                                                                                                                                                                                                                                                                                                                    |
| **JFROG\_CLI\_PLUGINS_REPO**   | \[Default: 'jfrog-cli-plugins'\]<br><br>Can be optionally used with the JFROG\_CLI\_PLUGINS_SERVER environment variable. Determines the name of the local repository to use.                                                                                                                                                                                                                                                                          |
| **JFROG\_CLI\_RELEASES_REPO**  | Configured Artifactory repository name from which to download the jar needed by the mvn/gradle command.<br> This environment variable's value format should be `<server ID configured by the 'jf c add' command>/<repo name>`.<br> The repository should proxy https://releases.jfrog.io.<br> This environment variable is used by the 'jf mvn' and 'jf gradle' commands, and also by the 'jf audit' command, when used for maven or gradle projects. |
| **CI**                         | \[Default: false\]<br><br>If true, disables interactive prompts and progress bar.                                                                                                                                                                                                                                                                                                                                                                     |

## JFrog Platform Configuration

### Adding and Editing Configured Servers

The **config add** and **config edit** commands are used to add and edit JFrog Platform server configuration, stored in JFrog CLI's configuration storage. These configured servers can be used by the other commands. The configured servers' details can be overridden per command by passing in alternative values for the URL and login credentials. The values configured are saved in file under the JFrog CLI home directory.

|     |     |
| --- | --- |
| Command name | config add / config edit |
| Abbreviation | c add / c edit |
| Command options |     |
| --access-token | \[Optional\]<br><br>Access token. |
| --artifactory-url | \[Optional\]<br><br>Artifactory URL. |
| --basic-auth-only | \[Default: false\]<br><br>Used for Artifactory authentication. Set to true to disable replacing username and password/API key with automatically created access token that's refreshed hourly. Username and password/API key will still be used with commands which use external tools or the JFrog Distribution service. Can only be passed along with username and password/API key options. |
| --client-cert-key-path | \[Optional\]<br><br>Private key file for the client certificate in PEM format. |
| --client-cert-path | \[Optional\]<br><br>Client certificate file in PEM format. |
| --dist-url | \[Optional\]<br><br>Distribution URL. |
| --enc-password | \[Default: true\]<br><br>If true, the configured password will be encrypted using Artifactory's[encryption API](https://www.jfrog.com/confluence/display/RTF/Artifactory+REST+API#ArtifactoryRESTAPI-GetUserEncryptedPassword)before being stored. If false, the configured password will not be encrypted. |
| --insecure-tls | Default: false\]<br><br>Set to true to skip TLS certificates verification, while encrypting the Artifactory password during the config process. |
| --interactive | \[Default: true, unless $CI is true\]<br><br>Set to false if you do not want the config command to be interactive. |
| --mission-control-url | \[Optional\]<br><br>Mission Control URL. |
| --password | \[Optional\]<br><br>JFrog Platform password. |
| --pipelines-url | \[Optional\]<br><br>Pipelines URL. |
| --ssh-key-path | \[Optional\]<br><br>For authentication with Artifactory. SSH key file path. |
| --url | \[Optional\]<br><br>JFrog platform URL. |
| --user | \[Optional\]<br><br>JFrog Platform username. |
| --xray-url | \[Optional\] Xray URL. |
| --overwrite | \[Available for _config add_ only\]<br><br>\[Default: false\]<br><br>Overwrites the instance configuration if an instance with the same ID already exists. |
| Command arguments |     |
| server ID | A unique ID for the server configuration. |

### Removing Configured Servers

The _config remove_ command is used to remove JFrog Platform server configuration, stored in JFrog CLI's configuration storage.

|     |     |
| --- | --- |
| Command name | config remove |
| Abbreviation | c rm |
| Command options |     |
| --quiet | \[Default: $CI\]<br><br>Set to true to skip the delete confirmation message. |
| Command arguments |     |
| server ID | The server ID to remove. If no argument is sent, all configured servers are removed. |

### Showing the Configured Servers

The _config show_ command shows the stored configuration. You may show a specific server's configuration by sending its ID as an argument to the command.

|     |     |
| --- | --- |
| Command name | config show |
| Abbreviation | c s |
| Command arguments |     |
| server ID | The ID of the server to show. If no argument is sent, all configured servers are shown. |

### Setting a Server as Default

The _config use_ command sets a configured server as default. The following commands will use this server.

|     |     |
| --- | --- |
| Command name | config use |
| Command arguments |     |
| server ID | The ID of the server to set as default. |

### Exporting and Importing Configuration

The _config export_ command generates a token, which stores the server configuration. This token can be used by the _config import_ command, to import the configuration stored in the token, and save it in JFrog CLI's configuration storage.

#### Export

|     |     |
| --- | --- |
| Command name | config export |
| Abbreviation | c ex |
| Command arguments |     |
| server ID | The ID of the server to export |

#### Import

|     |     |
| --- | --- |
| Command name | config import |
| Abbreviation | c im |
| Command arguments |     |
| server token | The token to import |

## Setting up a CI Pipeline

The **ci-setup** command allows setting up a basic CI pipeline with the JFrog Platform, while automatically configuring the JFrog Platform to serve the pipeline. It is an interactive command, which prompts you with a series for questions, such as your source control details, your build tool, build command and your CI provider. The command then uses this information to do following:

* Create the repositories in JFrog Artifactory, to be used by the pipeline to resolve dependencies.
* Configure JFrog Xray to scan the build.
* Generate a basic CI pipeline, which builds and scans your code.

You can use the generated CI pipeline as a working starting point and then expand it as needed.

The command currently supports the following package managers:

* Maven
* Gradle
* npm.

and the following CI providers:

* JFrog Pipelines
* Jenkins
* GitHub Actions.

Usage:
```
jf ci-setup
```
  

## Proxy Support

JFrog CLI supports using an HTTP/S proxy. All you need to do is set HTTP_PROXY or HTTPS_PROXY environment variable with the proxy URL.

HTTP_PROXY, HTTPS_PROXY and NO_PROXY are the industry standards for proxy usages.

|     |     |
| --- | --- |
| **Variable Name** | **Description** |
| HTTP_PROXY | Determines a URL to an HTTP proxy. |
| HTTPS_PROXY | Determines a URL to an HTTPS proxy. |
| NO_PROXY | Use this variable to bypass the proxy to IP addresses, subnets or domains. This may contain a comma-separated list of hostnames or IPs without protocols and ports. A typical usage may be to set this variable to Artifactory’s IP address. |

## Shell Auto-Completion

If you're using JFrog CLI from a bash, zsh, or fish shells, you can install JFrog CLI's auto-completion scripts.

### Install JFrog CLI with Homebrew?
If you're installing JFrog CLI using Homebrew, the bash, zsh, or fish auto-complete scripts are automatically installed by Homebrew. Please make sure that your _.bash_profile_ or _.zshrc_ are configured as described in the [Homebrew Shell Completion documentation](https://docs.brew.sh/Shell-Completion).

### Using Oh My Zsh?
With your favourite text editor, open $HOME/.zshrc and add **jfrog** to the plugin list.

For example:
```
plugins=(git mvn npm sdk jfrog)
```

To install auto-completion for **bash**, run the following command and follow the instructions to complete the installation:
```
jf completion bash --install
```

To install auto-completion for **zsh**, run the following command and follow the instructions to complete the installation:
```
jf completion zsh --install
```

To install auto-completion for **fish**, run the following command:
```
jf completion fish --install
```

## Sensitive Data Encryption

Since version 1.37.0, JFrog CLI supports encrypting the sensitive data stored in JFrog CLI's config. To enable encryption, follow these steps.

* Create a random 32 character master key. Make sure that the key size is exactly 32 characters. For example _f84hc22dQfhe9f8ydFwfsdn48!wejh8A_
* Create a file named **security.yaml** under **~/.jfrog/security**.

> If you modified the default JFrog CLI home directory by setting JFROG\_CLI\_HOME_DIR environment variable, then the **security/security.yaml** file should br created under the configured home directory.
    
* Add the master key you generated to security.yaml. The file content should be:
    
```
version: 1
masterKey: "your master key"
```
* Make sure that the only permission security.yaml has is read for the user running JFrog CLI. 

The configuration will be encrypted the next time JFrog CLI attempts to access the config.

> Warning: When upgrading JFrog CLI from a version prior to 1.37.0 to version 1.37.0 or above, JFrog CLI automatically makes changes to the content of the ~/`_.jfrog_` directory, to support the new functionality introduced in version 1.37.0. Before making these changes, the content of the `_~/.jfrog_` directory is backed up inside the ~/`_.jfrog/backup_` directory. Therefore, after enabling sensitive data encryption, it is recommended to remove the `_backup_` directory, to ensure no sensitive data is left unencrypted.


## JFrog CLI Plugins
### General

JFrog CLI Plugins allow enhancing the functionality of JFrog CLI to meet the specific user and organization needs. The source code of a plugin is maintained as an open source Go project on GitHub. All public plugins are registered in [JFrog CLI's Plugins Registry](https://github.com/jfrog/jfrog-cli-plugins-reg). We encourage you, as developers, to create plugins and share them publicly with the rest of the community. When a plugin is included in the registry, it becomes publicly available and can be installed using JFrog CLI. Read the [JFrog CLI Plugins Developer Guide](https://github.com/jfrog/jfrog-cli/blob/dev/guides/jfrog-cli-plugins-developer-guide.md) if you wish to create and publish your own plugins.

### Installing Plugins

A plugin which is included [JFrog CLI's Plugins Registry](https://github.com/jfrog/jfrog-cli-plugins-reg) can be installed using the following command.
```
$ jf plugin install the-plugin-name
```
This command will install the plugin from the official public registry by default. You can also install a plugin from a private JFrog CLI Plugin registry, as described in the _Private Plugins Registries_ section.

### Private Plugins Registries

In addition to the public official JFrog CLI Plugins Registry, JFrog CLI supports publishing and installing plugins to and from private JFrog CLI Plugins Registries. A private registry can be hosted on any Artifactory server. It uses a local generic Artifactory repository for storing the plugins.

To create your own private plugins registry, follow these steps.

* On your Artifactory server, create a local generic repository named _jfrog-cli-plugins_.
* Make sure your Artifactory server is included in JFrog CLI's configuration, by running the _jf c show_ command.
* If needed, configure your Artifactory instance using the _jfrog c add_ command.
* Set the ID of the configured server as the value of the JFROG\_CLI\_PLUGINS_SERVER environment variable.
* If you wish the name of the plugins repository to be different than jfrog-cli-plugins, set this name as the value of the JFROG\_CLI\_PLUGINS_REPO environment variable.

The **jf plugin install** command will now install plugins stored in your private registry.

To publish a plugin to the private registry, run the following command, while inside the root of the plugin's sources directory. This command will build the sources of the plugin for all the supported operating systems. All binaries will be uploaded to the configured registry.
```
jf plugin publish the-plugin-name the-plugin-version
```

## Release Notes
* [Release notes](https://github.com/jfrog/jfrog-cli/releases) for [JFrog CLI v2](https://jfrog.com/help/r/jfrog-cli/jfrog-cli-v2)
* [Release notes](https://github.com/jfrog/jfrog-cli/blob/v1/RELEASE.md#release-notes) for the legacy releases of JFrog CLI

  
