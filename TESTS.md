# Tests

Below is a summary of the testing guidelines for JFrog-CLI.

### Status

Listed below are the current statuses of all the testing suites:
<table>
   <tr>
      <th></th>
      <th width="100">V2</th>
      <th width="100">DEV</th>
   </tr>
   <tr>
      <td><img align="center" src="./images/artifactory.png" alt="artifactory"> Artifactory</td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Artifactory%20Tests/v2?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Artifactory%20Tests/dev?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
   </tr>
   <tr>
      <td><img align="center" src="./images/xray.png" alt="xray"> Xray</td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Xray%20Tests/v2?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Xray%20Tests/dev?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
   </tr>
   <tr>
      <td><img align="center" src="./images/distribution.png" alt="distribution"> Distribution</td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Distribution%20Tests/v2?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Distribution%20Tests/dev?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
   </tr>
   </tr>
   <tr>
      <td><img align="center" src="./images/access.png" alt="access"> Access</td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Access%20Tests/v2?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Access%20Tests/dev?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
   </tr>
   </tr>
   <tr>
      <td><img align="center" src="./images/mvn.png" alt="mvn"> Maven</td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/mvn%20Tests/v2?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/mvn%20Tests/dev?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
   </tr>
   </tr>
   <tr>
      <td><img align="center" src="./images/gradle.png" alt="gradle"> Gradle</td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Gradle%20Tests/v2?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Gradle%20Tests/dev?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
   </tr>
   <tr>
      <td><img align="center" src="./images/npm.png" alt="npm"> npm</td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/npm%20Tests/v2?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/npm%20Tests/dev?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
   </tr>
   </tr>
   <tr>
      <td><img align="center" src="./images/docker.png" alt="docker"> Docker</td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Docker%20Tests/v2?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Docker%20Tests/dev?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
   </tr>
   <tr>
      <td><img align="center" src="./images/nuget.png" alt="nuget"> NuGet</td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/NuGet%20Tests/v2?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/NuGet%20Tests/dev?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
   </tr>
   </tr>
   <tr>
      <td><img align="center" src="./images/python.png" alt="python"> pip</td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/pip%20Tests/v2?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/pip%20Tests/dev?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
   </tr>
   </tr>
   <tr>
      <td><img align="center" src="./images/go.png" alt="go"> GO</td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/GO%20Tests/v2?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/GO%20Tests/dev?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
   </tr>
   </tr>
   <tr>
      <td> 📃  Scripts</td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Scripts%20Tests/v2?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Scripts%20Tests/dev?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
   </tr>
   </tr>
   <tr>
      <td> Code Analysis</td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Static%20Analysis/v2?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Static%20Analysis/dev?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
   </tr>
   <tr>
      <td> Plugins</td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Plugins%20Tests/v2?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Plugins%20Tests/dev?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
   </tr>
   </tr>
   <tr>
      <td>Lint Code</td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Lint%20Tests/v2?label=%20&style=for-the-badge" alt="1">
         </div>
      </td>
      <td>
         <div align="center">
            <img text-align="center" src="https://img.shields.io/github/workflow/status/jfrog/jfrog-cli/Lint%20Tests/dev?label=%20&style=for-the-badge" alt="1">
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

| Flag                | Description                                                                                     |
| ------------------- | ----------------------------------------------------------------------------------------------- |
| `-jfrog.url`        | [Default: <http://localhost:8081>] JFrog platform URL.                                          |
| `-jfrog.user`       | [Default: admin] JFrog platform username.                                                       |
| `-jfrog.password`   | [Default: password] JFrog platform password.                                                    |
| `-jfrog.adminToken` | JFrog platform admin token.                                                                     |
| `-ci.runId`         | [Optional] A unique identifier used as a suffix to create repositories and builds in the tests. |

The types are:

| Type                 | Description        |
| -------------------- | ------------------ |
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
| ---------------------------- | -------------------------------------------------------------- |
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

Happy coding! 👋.
