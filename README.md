[![JFrog CLI](images/jfrog-cli-intro.png)](#readme)

<div align="center">

# JFrog CLI

[![Go Report Card](https://goreportcard.com/badge/github.com/jfrog/jfrog-cli)](https://goreportcard.com/report/github.com/jfrog/jfrog-cli)
[![license](https://img.shields.io/badge/License-Apache_2.0-blue.svg?style=flat)](https://raw.githubusercontent.com/jfrog/jfrog-cli/v2/LICENSE) [![](https://img.shields.io/badge/Docs-%F0%9F%93%96-blue)](https://www.jfrog.com/confluence/display/CLI/JFrog+CLI)

</div>

| Branch |                                                                                                                                                                              Status                                                                                                                                                                              |
|:------:|:----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------:|
|   v2   |     [![JFrog CLI Tests](https://github.com/jfrog/jfrog-cli/actions/workflows/tests.yml/badge.svg?branch=v2)](https://github.com/jfrog/jfrog-cli/actions/workflows/tests.yml) [![Static Analysis](https://github.com/jfrog/jfrog-cli/actions/workflows/analysis.yml/badge.svg?branch=v2)](https://github.com/jfrog/jfrog-cli/actions/workflows/analysis.yml)      |
|  dev   | [![JFrog CLI Tests](https://github.com/jfrog/jfrog-cli/actions/workflows/tests.yml/badge.svg?branch=dev)](https://github.com/jfrog/jfrog-cli/actions/workflows/tests.yml)        [![Static Analysis](https://github.com/jfrog/jfrog-cli/actions/workflows/analysis.yml/badge.svg?branch=dev)](https://github.com/jfrog/jfrog-cli/actions/workflows/analysis.yml) |
|   v1   |                                                                                             [![JFrog CLI Tests](https://github.com/jfrog/jfrog-cli/actions/workflows/tests.yml/badge.svg?branch=v1)](https://github.com/jfrog/jfrog-cli/actions/workflows/tests.yml)                                                                                             |
| dev-v1 |                                                                                           [![JFrog CLI Tests](https://github.com/jfrog/jfrog-cli/actions/workflows/tests.yml/badge.svg?branch=dev-v1)](https://github.com/jfrog/jfrog-cli/actions/workflows/tests.yml)                                                                                           |

# Table of Contents

- [Overview](#overview)
- [Download and Installation](#download-and-installation)
- [Building the Executable](#building-the-executable)
- [Tests](#tests)
- [Code Contributions](#code-contributions)
- [Using JFrog CLI](#using-jfrog-cli)
- [JFrog CLI Plugins](#jfrog-cli-plugins)
- [Release Notes](#release-notes)

# Overview

JFrog CLI is a compact and smart client that provides a simple interface that automates access to _Artifactory_ and _Mission Control_ through their respective REST APIs.
By using the JFrog CLI, you can greatly simplify your automation scripts making them more readable and easier to maintain.
Several features of the JFrog CLI makes your scripts more efficient and reliable:

- Multi-threaded upload and download of artifacts make builds run faster
- Checksum optimization reduces redundant file transfers
- Wildcards and regular expressions give you an easy way to collect all the artifacts you wish to upload or download.
- "Dry run" gives you a preview of file transfer operations before you actually run them

# Download and Installation

You can either install JFrog CLI using one of the supported installers or download its executable directly. Visit the [Install JFrog CLI Page](https://jfrog.com/getcli/) for details.

# Building the Executable

JFrog CLI is written in the [Go programming language](https://golang.org/), so to build the CLI yourself, you first need to have Go installed and configured on your machine.

## Install Go

To download and install `Go`, please refer to the [Go documentation](https://golang.org/doc/install).
Please download `Go 1.14.x` or above.

## Download and Build the CLI

Navigate to a directory where you want to create the jfrog-cli project, **outside** the `$GOPATH` tree.

If the `GOPATH` variable is unset, it's default value is the go folder under the user home.

Verify that the `GO111MODULE` variable is either unset, or explicitly set to `auto`.

Clone the jfrog-cli project by executing the following command:

```
git clone https://github.com/jfrog/jfrog-cli
```

Build the project by navigating to the jfrog folder and executing the following commands.
On Unix based systems run:

```
cd jfrog-cli
build/build.sh
```

On Windows run:

```
cd jfrog-cli
build\build.bat
```

Once completed, you will find the JFrog CLI executable at your current directory.

# Tests

### Usage

```
go test -v github.com/jfrog/jfrog-cli [test-types] [flags]
```

The flags are:

| Flag                | Description                                                                                     |
| ------------------- | ----------------------------------------------------------------------------------------------- |
| `-jfrog.url`        | [Default: http://localhost:8081] JFrog platform URL.                                            |
| `-jfrog.user`       | [Default: admin] JFrog platform username.                                                       |
| `-jfrog.password`   | [Default: password] JFrog platform password.                                                    |
| `-jfrog.adminToken` | JFrog platform admin token.                                                                     |
| `-ci.runId`         | [Optional] A unique identifier used as a suffix to create repositories and builds in the tests. |

The types are:

| Type                 | Description        |
|----------------------|--------------------|
| `-test.artifactory`  | Artifactory tests  |
| `-test.access`       | Access tests       |
| `-test.npm`          | Npm tests          |
| `-test.maven`        | Maven tests        |
| `-test.gradle`       | Gradle tests       |
| `-test.docker`       | Docker tests       |
| `-test.go`           | Go tests           |
| `-test.pip`          | Pip tests          |
| `-test.pipenv`       | Pipenv tests       |
| `-test.nuget`        | Nuget tests        |
| `-test.plugins`      | Plugins tests      |
| `-test.distribution` | Distribution tests |
| `-test.xray`         | Xray tests         |

- Running the tests will create builds and repositories with timestamps,
  for example: `cli-rt1-1592990748` and `cli-rt2-1592990748`.<br/>
  Once the tests are completed, the content of these repositories will be deleted.

#### Artifactory tests

In addition to [general optional flags](#Usage) you can use the following optional artifactory flags.

| Flag                   | Description                                                                                             |
| ---------------------- | ------------------------------------------------------------------------------------------------------- |
| `-jfrog.sshKeyPath`    | [Optional] Ssh key file path. Should be used only if the Artifactory URL format is ssh://[domain]:port. |
| `-jfrog.sshPassphrase` | [Optional] Ssh key passphrase.                                                                          |

##### Examples

To run artifactory tests execute the following command.

```
go test -v github.com/jfrog/jfrog-cli -test.artifactory [flags]
```

#### Npm tests

##### Requirements

- The _npm_ executables should be included as part of the _PATH_ environment variable.
- The tests are compatible with npm 7 and higher.

##### Limitation

- Currently, npm integration support only http(s) connections to Artifactory using username and password.

##### Examples

To run npm tests execute the following command.

```
go test -v github.com/jfrog/jfrog-cli -test.npm [flags]
```

#### Maven tests

##### Requirements

- The _java_ executable should be included as part of the _PATH_ environment variable. Alternatively, set the _JAVA_HOME_ environment variable.

##### Limitation

- Currently, maven integration support only http(s) connections to Artifactory using username and password.

##### Examples

To run maven tests execute the following command.

```
go test -v github.com/jfrog/jfrog-cli -test.maven [flags]
```

#### Gradle tests

##### Requirements

- The _gradle_ executables should be included as part of the _PATH_ environment variable.
- The _java_ executable should be included as part of the _PATH_ environment variable. Alternatively, set the _JAVA_HOME_ environment variable.

##### Limitation

- Currently, gradle integration support only http(s) connections to Artifactory using username and password.

##### Examples

To run gradle tests execute the following command.

```
go test -v github.com/jfrog/jfrog-cli -test.gradle [flags]
```

#### Docker tests

In addition to [general optional flags](#Usage) you _must_ use the following docker flags.

##### Requirements

- On Linux machines, [Podman](https://podman.io/) tests will be running, so make sure it's available in the local path.

| Flag                         | Description                                                    |
|------------------------------|----------------------------------------------------------------|
| `-rt.dockerRepoDomain`       | Artifactory Docker registry domain.                            |
| `-rt.dockerVirtualRepo`      | Artifactory Docker virtual repository name.                    |
| `-rt.dockerRemoteRepo`       | Artifactory Docker remote repository name.                     |
| `-rt.dockerLocalRepo`        | Artifactory Docker local repository name.                      |
| `-rt.dockerPromoteLocalRepo` | Artifactory Docker local repository name - Used for promotion. |

##### Examples

To run docker tests execute the following command (fill out the missing parameters as described below).

```
go test -v github.com/jfrog/jfrog-cli -test.docker -rt.dockerRepoDomain=DOCKER_DOMAIN -rt.DockerLocalRepo=DOCKER_LOCAL_REPO [flags]
```

#### Go commands tests

##### Requirement

- The tests are compatible with Artifactory 6.10 and higher.
- To run go tests run the following command:

```
go test -v github.com/jfrog/jfrog-cli -test.go [flags]
```

#### NuGet tests

##### Requirement

- Add NuGet executable to the system search path (PATH environment variable).
- Run the following command:

```
go test -v github.com/jfrog/jfrog-cli -test.nuget [flags]
```

#### Pip tests

##### Requirement

- Add Python and pip executables to the system search path (PATH environment variable).
- Run the following command:

```
go test -v github.com/jfrog/jfrog-cli -test.pip [flags]
```

#### Plugins tests

To run Plugins tests execute the following command:

```
go test -v github.com/jfrog/jfrog-cli -test.plugins
```

### Distribution tests

To run Distribution tests execute the following command:

```
go test -v github.com/jfrog/jfrog-cli -test.distribution [flags]
```

### Xray tests

To run Xray tests execute the following command:

```
go test -v github.com/jfrog/jfrog-cli -test.xray [flags]
```

# Code Contributions

We welcome code contributions through pull requests from the community.

## Pull Requests Guidelines

- If the existing tests do not already cover your changes, please add tests..
- Pull requests should be created on the _dev_ branch.
- Please use [gofmt](https://golang.org/cmd/gofmt/) for formatting the code before submitting the pull request.

## Dependencies in other JFrog modules

This project heavily depends on:
- github.com/jfrog/jfrog-client-go
- github.com/jfrog/build-info-go
- github.com/jfrog/jfrog-cli-core

### Local Development
During local development, when you encounter code that needs to be changed from one of the above modules, it is 
recommended to replace the dependency to work with a local clone of the dependency.

For example, assuming you would like to change files from jfrog-cli-core.
Clone jfrog-cli-core (preferably your fork) to your local development machine 
(assuming it will be cloned to `/repos/jfrog-cli-core`).

Change go.mod to include the following:
```
replace github.com/jfrog/jfrog-cli-core/v2 => /repos/jfrog-cli-core
```

### Pull Requests
Once done with your coding, you should push the changes you made to the other modules first. Once pushed, you can change this 
project to resolve the dependencies from your github fork / branch.
This is done by pointing the dependency in go.mod to your repository and branch. For example:
```
replace github.com/jfrog/jfrog-cli-core/v2 => github.com/galusben/jfrog-cli-core/v2 dev
```

Then run `go mod tidy`

Notice that go will change the version in the go.mod file.

# Using JFrog CLI

JFrog CLI can be used for a variety of functions with Artifactory, Xray and Mission Control,
and has a dedicated set of commands for each product.
To learn how to use JFrog CLI, please visit the [JFrog CLI User Guide](https://www.jfrog.com/confluence/display/CLI/Welcome+to+JFrog+CLI).

# JFrog CLI Plugins

JFrog CLI plugins support enhancing the functionality of JFrog CLI to meet the specific user and organization needs. The source code of a plugin is maintained as an open source Go project on GitHub. All public plugins are registered in JFrog CLI's Plugins Registry, which is hosted in the [jfrog-cli-plugins-reg](https://github.com/jfrog/jfrog-cli-plugins-reg) GitHub repository. We encourage you, as developers, to create plugins and share them publically with the rest of the community. Read more about this in the [JFrog CLI Plugin Developer Guide](guides/jfrog-cli-plugins-developer-guide.md).

# Release Notes

The release notes are available [here](https://github.com/jfrog/jfrog-cli/releases).
