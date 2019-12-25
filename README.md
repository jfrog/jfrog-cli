<p align="center">
  <a href="https://jfrog.com/">
    <img src="https://github.com/jfrog/jfrog-cli-go/blob/master/npm/assets/jfrog.jpg?raw=true" alt="JFrog logo" width="200">
  </a>
</p>

# Status

[![Build status](https://img.shields.io/appveyor/ci/jfrog-ecosystem/jfrog-cli-go/master?label=build%40master&logo=appveyor)](https://ci.appveyor.com/project/jfrog-ecosystem/jfrog-cli-go/branch/master)
[![Build status](https://img.shields.io/appveyor/ci/jfrog-ecosystem/jfrog-cli-go/dev?label=build%40dev&logo=appveyor)](https://ci.appveyor.com/project/jfrog-ecosystem/jfrog-cli-go/branch/dev)
[![Release status](https://img.shields.io/github/v/release/jfrog/jfrog-cli?color=blue)](https://github.com/jfrog/jfrog-cli/releases)
[![npm version](https://img.shields.io/npm/v/jfrog-cli-go.svg?color=blue)](https://www.npmjs.com/package/jfrog-cli-go)
[![brew version](https://img.shields.io/homebrew/v/jfrog-cli-go?color=blue)](https://formulae.brew.sh/formula/jfrog-cli-go)

# OS Support

[![Linux-386](https://img.shields.io/bintray/v/jfrog/jfrog-cli-go/jfrog-cli-linux-386?color=lightgrey&label=linux-386&logo=Linux)](https://bintray.com/jfrog/jfrog-cli-go/jfrog-cli-linux-386/_latestVersion)
[![Linux-amd64](https://img.shields.io/bintray/v/jfrog/jfrog-cli-go/jfrog-cli-linux-amd64?color=lightgrey&label=linux-amd64&logo=Linux) ](https://bintray.com/jfrog/jfrog-cli-go/jfrog-cli-linux-amd64/_latestVersion)
[![Linux-arm](https://img.shields.io/bintray/v/jfrog/jfrog-cli-go/jfrog-cli-linux-arm?color=lightgrey&label=linux-arm&logo=Linux) ](https://bintray.com/jfrog/jfrog-cli-go/jfrog-cli-linux-arm/_latestVersion)
[![Linux-arm64](https://img.shields.io/bintray/v/jfrog/jfrog-cli-go/jfrog-cli-linux-arm64?color=lightgrey&label=linux-arm64&logo=Linux) ](https://bintray.com/jfrog/jfrog-cli-go/jfrog-cli-linux-arm64/_latestVersion)
[![Mac-386](https://img.shields.io/bintray/v/jfrog/jfrog-cli-go/jfrog-cli-mac-386?color=lightgrey&label=mac&logo=Apple) ](https://bintray.com/jfrog/jfrog-cli-go/jfrog-cli-mac-386/_latestVersion)
[![Windows-amd64](https://img.shields.io/bintray/v/jfrog/jfrog-cli-go/jfrog-cli-windows-amd64?color=lightgrey&label=windows&logo=windows) ](https://bintray.com/jfrog/jfrog-cli-go/jfrog-cli-windows-amd64/_latestVersion)

# Overview

JFrog CLI is a compact and smart client that provides a simple interface that automates access to [Artifactory](https://jfrog.com/artifactory/), [Bintray](https://bintray.com/) and [Mission Control](https://jfrog.com/mission-control/) through their respective REST APIs.
By using the JFrog CLI, you can greatly simplify your automation scripts making them more readable and easier to maintain.
Several features of the JFrog CLI makes your scripts more efficient and reliable:

- Multi-threaded upload and download of artifacts make builds run faster
- Checksum optimization reduces redundant file transfers
- Wildcards and regular expressions give you an easy way to collect all the artifacts you wish to upload or download.
- "Dry run" gives you a preview of file transfer operations before you actually run them

# Download and Installation

You can download the executable directly using the [JFrog CLI Download Page](https://www.jfrog.com/getcli/), or install it with npm, homebrew or docker.
## NPM
````
npm install jfrog-cli-go
````
## Homebrew
````
brew install jfrog-cli-go
````
## Docker
````
docker run docker.bintray.io/jfrog/jfrog-cli-go:latest jfrog <COMMAND>
````

# Building the Executable

JFrog CLI is written in the [Go programming language](https://golang.org/), so to build the CLI yourself, you first need to have Go installed and configured on your machine.

## Install Go

To download and install `Go`, please refer to the [Go documentation](https://golang.org/doc/install).
Please download `Go 1.12.7` or above.

## Download and Build the CLI

Navigate to a directory where you want to create the jfrog-cli-go project, **outside** the `$GOPATH` tree.

If the `GOPATH` variable is unset, it's default value is the go folder under the user home.

Verify that the `GO111MODULE` variable is either unset, or explicitly set to `auto`.

Clone the jfrog-cli project by executing the following command:
````
git clone https://github.com/jfrog/jfrog-cli
````
Build the project by navigating to the jfrog folder and executing the following commands.
On Unix based systems run:
````
cd jfrog-cli
./build.sh
````
On Windows run:
````
cd jfrog-cli
build.bat
````
Once completed, you will find the JFrog CLI executable at your current directory.

# Tests
### Usage
On Unix based systems run:
````
./test.sh -v github.com/jfrog/jfrog-cli-go [test-types] [flags]
````
On Windows run:
````
test.bat -v github.com/jfrog/jfrog-cli-go [test-types] [flags]
````

The flags are:

| Flag | Description |
| --- | --- |
| `-rt.url` | [Default: http://localhost:8081/artifactory] Artifactory URL. |
| `-rt.user` | [Default: admin] Artifactory username.|
| `-rt.password` | [Default: password] Artifactory password. |
| `-rt.apikey` | Artifactory API key. |
| `-rt.accessToken` | Artifactory access token. |

The types are:

| Type | Description |
| --- | --- |
| `-test.artifactory` | Artifactory tests |
| `-test.npm` | Npm tests |
| `-test.maven` | Maven tests |
| `-test.gradle` | Gradle tests |
| `-test.docker` | Docker tests |
| `-test.go` | Go tests |
| `-test.pip` | Pip tests |
| `-test.nuget` | Nuget tests |

* Running the tests will create two repositories: `jfrog-cli-tests-repo` and `jfrog-cli-tests-repo1`.<br/>
Once the tests are completed, the content of these repositories will be deleted.
 
#### Artifactory tests
In addition to [general optional flags](#Usage) you can use the following optional artifactory flags.

| Flag | Description |
| --- | --- |
| `-rt.sshKeyPath` | [Optional] Ssh key file path. Should be used only if the Artifactory URL format is ssh://[domain]:port. |
| `-rt.sshPassphrase` | [Optional] Ssh key passphrase. |


##### Examples
To run artifactory tests execute the following command.
On Unix based systems run:
````
./test.sh -v github.com/jfrog/jfrog-cli-go -test.artifactory [flags]
````
On Windows run:
````
test.bat -v github.com/jfrog/jfrog-cli-go -test.artifactory [flags]
````

#### Npm tests
##### Requirement
* The *npm* executables should be included as part of the *PATH* environment variable.

##### Limitation
* Currently, npm integration support only http(s) connections to Artifactory using username and password.

##### Examples
To run npm tests execute the following command.

On Unix based systems run:
````
./test.sh -v github.com/jfrog/jfrog-cli-go -test.npm [flags]
````
On Windows run:
````
test.bat -v github.com/jfrog/jfrog-cli-go -test.npm [flags]
````

#### Maven tests
##### Requirements
* The *M2_HOME* environment variable should be set to the local maven installation path.
* The *java* executable should be included as part of the *PATH* environment variable. Alternatively, set the *JAVA_HOME* environment variable.

##### Limitation
* Currently, maven integration support only http(s) connections to Artifactory using username and password.

##### Examples
To run maven tests execute the following command.
On Unix based systems run:
````
./test.sh -v github.com/jfrog/jfrog-cli-go -test.maven [flags]
````
On Windows run:
````
test.bat -v github.com/jfrog/jfrog-cli-go -test.maven [flags]
````

#### Gradle tests
##### Requirements
* The *gradle* executables should be included as part of the *PATH* environment variable.
* The *java* executable should be included as part of the *PATH* environment variable. Alternatively, set the *JAVA_HOME* environment variable.

##### Limitation
* Currently, gradle integration support only http(s) connections to Artifactory using username and password.

##### Examples
To run gradle tests execute the following command.

On Unix based systems run:
````
./test.sh -v github.com/jfrog/jfrog-cli-go -test.gradle [flags]
````
On Windows run:
````
test.bat -v github.com/jfrog/jfrog-cli-go -test.gradle [flags]
````

#### Docker tests
In addition to [general optional flags](#Usage) you *must* use the following docker flags.

| Flag | Description |
| --- | --- |
| `-rt.dockerRepoDomain` | Artifactory Docker registry domain. |
| `-rt.dockerTargetRepo` | Artifactory Docker repository name. |

##### Examples
To run docker tests execute the following command (fill out the missing parameters as described below).

On Unix based systems run:
````
./test.sh -v github.com/jfrog/jfrog-cli-go -test.docker -rt.dockerRepoDomain=DOCKER_DOMAIN -rt.dockerTargetRepo=DOCKER_TARGET_REPO [flags]
````
On Windows run:
````
test.bat -v github.com/jfrog/jfrog-cli-go -test.docker -rt.dockerRepoDomain=DOCKER_DOMAIN -rt.dockerTargetRepo=DOCKER_TARGET_REPO [flags]
````

#### Go commands tests
##### Examples
To run go tests run the following command.

On Unix based systems run:
````
./test.sh -v github.com/jfrog/jfrog-cli-go -test.go [flags]
````
On Windows run:
````
test.bat -v github.com/jfrog/jfrog-cli-go -test.go [flags]
````

#### NuGet tests
##### Requirement
* Add NuGet executable to the system search path (PATH environment variable).
* Create a remote repository named jfrog-cli-tests-nuget-remote-repo.
* Run the following command.

##### Examples
On Unix based systems run:
````
./test.sh -v github.com/jfrog/jfrog-cli-go -test.nuget [flags]
````
On Windows run:
````
test.bat -v github.com/jfrog/jfrog-cli-go -test.nuget [flags]
````

#### Pip tests
##### Requirement
* Add Python and pip executables to the system search path (PATH environment variable).
* Pip tests must run inside a clean pip-environment. You can either activate a virtual-environment and execute the tests from within, or provide the path to your virtual-environment using the -rt.pipVirtualEnv flag.
* Run the following command:

In addition to [general optional flags](#Usage) you can use the following optional pip flags

| Flag | Description |
| --- | --- |
| `-rt.pipVirtualEnv` | [Optional] Path to the directory of a clean pip virtual-environment. Make sure to provide the binaries directory (in unix: */bin*, in windows: *\Scripts*). |

##### Examples
On Unix based systems run:
````
./test.sh -v github.com/jfrog/jfrog-cli-go -test.pip [flags]
````
On Windows run:
````
test.bat -v github.com/jfrog/jfrog-cli-go -test.pip [flags]
````

### Bintray tests
Bintray tests credentials are taken from the CLI configuration. If non configured or not passed as flags, the tests will fail.

To run Bintray tests execute the following command: 
````
go test -v github.com/jfrog/jfrog-cli-go -test.bintray
````
Flags:

| Flag | Description |
| --- | --- |
| `-bt.user` | [Mandatory if not configured] Bintray username. |
| `-bt.key` | [Mandatory if not configured] Bintray API key. |
| `-bt.org` | [Optional] Bintray organization. If not configured, *-bt.user* is used as the organization name. |

* Running the tests will create a repository `jfrog-cli-tests-repo1` in bintray.<br/>
  Once the tests are completed, the repository will be deleted.

# Pull Requests [![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/jfrog/jfrog-cli/blob/master/CONTRIBUTING.md) [![GitHub license](https://img.shields.io/github/license/jfrog/jfrog-cli)](https://github.com/jfrog/jfrog-cli/blob/master/LICENSE) 
We welcome pull requests from the community.
## Guidelines
* Before creating your first pull request, please join our contributors community by signing [JFrog's CLA](https://secure.echosign.com/public/hostedForm?formid=5IYKLZ2RXB543N).
* If the existing tests do not already cover your changes, please add tests..
* Pull requests should be created on the *dev* branch.
* Please use [gofmt](https://golang.org/cmd/gofmt/) for formatting the code before submitting the pull request.

# Using JFrog CLI
JFrog CLI can be used for a variety of functions with Artifactory, Bintray, Xray and Mission Control,
and has a dedicated set of commands for each product.
To learn how to use JFrog CLI, please visit the [JFrog CLI User Guide](https://www.jfrog.com/confluence/display/CLI/Welcome+to+JFrog+CLI).

## Using JFrog CLI Docker Image
The docker image of JFrog CLI can be pulled from Bintray by running the following command:
````
docker pull docker.bintray.io/jfrog/jfrog-cli-go:latest
````
Run a JFrog CLI command using docker as follows:
````
docker run docker.bintray.io/jfrog/jfrog-cli-go:latest jfrog <COMMAND>
````

# Release Notes
The release are available on [Bintray](https://bintray.com/jfrog/jfrog-cli-go/jfrog-cli-linux-amd64#release).
