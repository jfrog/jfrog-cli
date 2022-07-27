# Tests

Below is a summary of the testing guidelines for JFrog-CLI.

### Status

Listed below are the current statuses of all the testing suites:

<table>
  <tr>
    <th></th>
    <th width="100" >V1</th>
    <th width="100" >DEV-V1</th>
  </tr>
  <tr>
    <td><img align="center" src="./images/artifactory.png" alt="artifactory"> Artifactory</td>
    <td>
        <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Artifactory%20Tests/v1?label=%20&style=for-the-badge" alt="1">
        </div>
    </td>
    <td>
        <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Artifactory%20Tests/dev-v1?label=%20&style=for-the-badge" alt="1">
        </div>
    </td>
  </tr>
  <tr>
    <td><img align="center" src="./images/xray.png" alt="xray"> Xray</td>
    <td>
        <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Xray%20Tests/v1?label=%20&style=for-the-badge" alt="1">
        </div>
    </td>
    <td>
        <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Xray%20Tests/dev-v1?label=%20&style=for-the-badge" alt="1">
        </div>
    </td>
  </tr>
  <tr>
    <td><img align="center" src="./images/distribution.png" alt="distribution"> Distribution</td>
    <td>
        <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Distribution%20Tests/v1?label=%20&style=for-the-badge" alt="1">
        </div>
    </td>
    <td>
        <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Distribution%20Tests/dev-v1?label=%20&style=for-the-badge" alt="1">
        </div>
    </td>
  </tr>
  </tr>
  <tr>
    <td><img align="center" src="./images/access.png" alt="access"> Access</td>
    <td>
        <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Access%20Tests/v1?label=%20&style=for-the-badge" alt="1">
        </div>
    </td>
    <td>
        <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Access%20Tests/dev-v1?label=%20&style=for-the-badge" alt="1">
        </div>
    </td>
  </tr>
  </tr>
  <tr>
    <td><img align="center" src="./images/mvn.png" alt="mvn"> Maven</td>
    <td>
        <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/mvn%20Tests/v1?label=%20&style=for-the-badge" alt="1">
        </div>
    </td>
    <td>
        <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/mvn%20Tests/dev-v1?label=%20&style=for-the-badge" alt="1">
        </div>
    </td>
  </tr>
  </tr>
  <tr>
    <td><img align="center" src="./images/gradle.png" alt="gradle"> Gradle</td>
    <td>
        <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Gradle%20Tests/v1?label=%20&style=for-the-badge" alt="1">
        </div>
    </td>
    <td>
        <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Gradle%20Tests/dev-v1?label=%20&style=for-the-badge" alt="1">
        </div>
    </td>
  </tr>

  <tr>
    <td><img align="center" src="./images/npm.png" alt="npm"> npm</td>
    <td>
        <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/npm%20Tests/v1?label=%20&style=for-the-badge" alt="1">
        </div>
    </td>
    <td>
        <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/npm%20Tests/dev-v1?label=%20&style=for-the-badge" alt="1">
        </div>
    </td>
  </tr>
  </tr>
  <tr>
    <td><img align="center" src="./images/docker.png" alt="docker"> Docker</td>
    <td>
        <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Docker%20Tests/v1?label=%20&style=for-the-badge" alt="1">
        </div>
    </td>
    <td>
        <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Docker%20Tests/dev-v1?label=%20&style=for-the-badge" alt="1">
        </div>
    </td>
  </tr>
  <tr>
    <td><img align="center" src="./images/nuget.png" alt="nuget"> NuGet</td>
    <td>
        <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/NuGet%20Tests/v1?label=%20&style=for-the-badge" alt="1">
        </div>
    </td>
    <td>
        <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/NuGet%20Tests/dev-v1?label=%20&style=for-the-badge" alt="1">
        </div>
    </td>
  </tr>
  </tr>
  <tr>
    <td><img align="center" src="./images/python.png" alt="python"> pip</td>
    <td>
        <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/pip%20Tests/v1?label=%20&style=for-the-badge" alt="1">
        </div>
    </td>
    <td>
        <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/pip%20Tests/dev-v1?label=%20&style=for-the-badge" alt="1">
        </div>
    </td>
  </tr>
  </tr>
  <tr>
    <td><img align="center" src="./images/go.png" alt="go"> GO</td>
    <td>
        <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/GO%20Tests/v1?label=%20&style=for-the-badge" alt="1">
        </div>
    </td>
    <td>
        <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/GO%20Tests/dev-v1?label=%20&style=for-the-badge" alt="1">
        </div>
    </td>
  </tr>
  <tr>
    <td>Plugins</td>
    <td>
        <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Plugins%20Tests/v1?label=%20&style=for-the-badge" alt="1">
        </div>
    </td>
    <td>
        <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Plugins%20Tests/dev-v1?label=%20&style=for-the-badge" alt="1">
        </div>
    </td>
  </tr>
  </tr>
</table>

### Usage

```
go test -v github.com/jfrog/jfrog-cli [test-types] [flags]
```

The flags are:

| Flag              | Description                                                                                     |
| ----------------- | ----------------------------------------------------------------------------------------------- |
| `-rt.url`         | [Default: http://localhost:8081/artifactory] Artifactory URL.                                   |
| `-rt.user`        | [Default: admin] Artifactory username.                                                          |
| `-rt.password`    | [Default: password] Artifactory password.                                                       |
| `-rt.apikey`      | Artifactory API key.                                                                            |
| `-rt.accessToken` | Artifactory access token.                                                                       |
| `-ci.runId`       | [Optional] A unique identifier used as a suffix to create repositories and builds in the tests. |

The types are:

| Type                | Description       |
| ------------------- | ----------------- |
| `-test.artifactory` | Artifactory tests |
| `-test.npm`         | Npm tests         |
| `-test.maven`       | Maven tests       |
| `-test.gradle`      | Gradle tests      |
| `-test.docker`      | Docker tests      |
| `-test.go`          | Go tests          |
| `-test.pip`         | Pip tests         |
| `-test.nuget`       | Nuget tests       |
| `-test.plugins`     | Plugins tests     |

- Running the tests will create builds and repositories with timestamps,
  for example: `cli-rt1-1592990748` and `cli-rt2-1592990748`.<br/>
  Once the tests are completed, the content of these repositories will be deleted.

#### Artifactory tests

| Flag                | Description                                                                                             |
| ------------------- | ------------------------------------------------------------------------------------------------------- |
| `-rt.sshKeyPath`    | [Optional] Ssh key file path. Should be used only if the Artifactory URL format is ssh://[domain]:port. |
| `-rt.sshPassphrase` | [Optional] Ssh key passphrase.                                                                          |

##### Examples

To run artifactory tests execute the following command.

```
go test -v github.com/jfrog/jfrog-cli -test.artifactory [flags]
```

#### Npm tests

##### Requirement

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

- The _M2_HOME_ environment variable should be set to the local maven installation path.
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

| Flag                    | Description                                 |
| ----------------------- | ------------------------------------------- |
| `-rt.dockerRepoDomain`  | Artifactory Docker registry domain.         |
| `-rt.dockerVirtualRepo` | Artifactory Docker virtual repository name. |
| `-rt.dockerRemoteRepo`  | Artifactory Docker remote repository name.  |
| `-rt.DockerLocalRepo`   | Artifactory Docker local repository name.   |

##### Examples

To run docker tests execute the following command (fill out the missing parameters as described below).

```
go test -v github.com/jfrog/jfrog-cli -test.docker -rt.dockerRepoDomain=DOCKER_DOMAIN -rt.DockerLocalRepo=DOCKER_LOCAL_REPO [flags]
```

#### Go commands tests

##### Examples

To run go tests run the following command.

```
go test -v github.com/jfrog/jfrog-cli -test.go [flags]
```

#### NuGet tests

##### Requirement

- Add NuGet executable to the system search path (PATH environment variable).
- Create a remote repository named jfrog-cli-tests-nuget-remote-repo.
- Run the following command.

##### Examples

```
go test -v github.com/jfrog/jfrog-cli -test.nuget [flags]
```

#### Pip tests

##### Requirement

- Add Python and pip executables to the system search path (PATH environment variable).
- Pip tests must run inside a clean pip-environment. You can either activate a virtual-environment and execute the tests from within, or provide the path to your virtual-environment using the -rt.pipVirtualEnv flag.
- Run the following command:

In addition to [general optional flags](#Usage) you can use the following optional pip flags

| Flag                | Description                                                                                                                                                 |
| ------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-rt.pipVirtualEnv` | [Optional] Path to the directory of a clean pip virtual-environment. Make sure to provide the binaries directory (in unix: _/bin_, in windows: _\Scripts_). |

##### Examples

```
go test -v github.com/jfrog/jfrog-cli -test.pip [flags]
```

#### Plugins tests

- To run Plugins tests execute the following command:

````
go test -v github.com/jfrog/jfrog-cli -test.plugins
```

### Bintray tests
Bintray tests credentials are taken from the CLI configuration. If non configured or not passed as flags, the tests will fail.

To run Bintray tests execute the following command:
````

go test -v github.com/jfrog/jfrog-cli -test.bintray

````
Flags:

| Flag | Description |
| --- | --- |
| `-bt.user` | [Mandatory if not configured] Bintray username. |
| `-bt.key` | [Mandatory if not configured] Bintray API key. |
| `-bt.org` | [Optional] Bintray organization. If not configured, *-bt.user* is used as the organization name. |

* Running the tests will create a repository named `cli-tests-bintray-<timestamp>` in bintray.<br/>
  Once the tests are completed, the repository will be deleted.

### Distribution tests
In addition to [general optional flags](#Usage) you can use the following flags:

| Flag | Description |
| --- | --- |
| `-rt.distUrl` | [Mandatory] JFrog Distribution URL. |
| `-rt.distAccessToken` | [Optional] Distribution access token. |

To run distribution tests run the following command:
```
go test -v github.com/jfrog/jfrog-cli -test.distribution [flags]
```

# Code Contributions

We welcome code contributions through pull requests from the community.

## Pull Requests Guidelines

* If the existing tests do not already cover your changes, please add tests..
* Pull requests should be created on the *dev* branch.
* Please use [gofmt](https://golang.org/cmd/gofmt/) for formatting the code before submitting the pull request.

# Using JFrog CLI
JFrog CLI can be used for a variety of functions with Artifactory, Bintray, Xray and Mission Control,
and has a dedicated set of commands for each product.
To learn how to use JFrog CLI, please visit the [JFrog CLI User Guide](https://www.jfrog.com/confluence/display/CLI/Welcome+to+JFrog+CLI).

# JFrog CLI Plugins

JFrog CLI plugins support enhancing the functionality of JFrog CLI to meet the specific user and organization needs. The source code of a plugin is maintained as an open source Go project on GitHub. All public plugins are registered in JFrog CLI's Plugins Registry, which is hosted in the [jfrog-cli-plugins-reg](https://github.com/jfrog/jfrog-cli-plugins-reg) GitHub repository. We encourage you, as developers, to create plugins and share them publically with the rest of the community. Read more about this in the [JFrog CLI Plugin Developer Guide](guides/jfrog-cli-plugins-developer-guide.md).

# Release Notes

The release notes are available [here](RELEASE.md#release-notes).
````
