# 📖 Guidelines

- Ensure that your changes are covered by existing tests. If not, please add new tests.
- Create pull requests on the `dev` branch.
- Before submitting the pull request, format the code by running `go fmt ./...`.
- Before submitting the pull request, ensure the code compiles by running `go vet ./...`.

# ⚒️ Building and Testing the Sources

## Building JFrog CLI

To build JFrog CLI, first, make sure Go is installed by running the following command:

```sh
go version
```

Next, clone the project sources and navigate to the root directory:

```sh
git clone https://github.com/jfrog/jfrog-cli.git
cd jfrog-cli
```

To build the sources on Unix-based systems, run:

```sh
./build/build.sh
```

On Windows, run:

```sh
.\build\build.bat
```

After the build process completes, you will find the `jf` or `jf.exe` executable in the current directory.

### Dependencies in other JFrog modules

This project heavily depends on the following modules:

- [github.com/jfrog/jfrog-client-go](https://github.com/jfrog/jfrog-client-go)
- [github.com/jfrog/jfrog-cli-core](github.com/jfrog/jfrog-cli-core)
- [github.com/jfrog/build-info-go](github.com/jfrog/build-info-go)
- [github.com/jfrog/gofrog](github.com/jfrog/gofrog)

#### Local Development

During local development, if you come across code that needs to be modified in one of the mentioned modules, it is advisable to replace the dependency with a local clone of the module.

For instance, let's assume you wish to modify files from `jfrog-cli-core`. Clone the `jfrog-cli-core` repository (preferably your fork) to your local development machine, placing it at `/local/path/in/your/machine/jfrog-cli-core`.

To include this local dependency, modify the `go.mod` file as follows:

```
replace github.com/jfrog/jfrog-cli-core/v2 => /local/path/in/your/machine/jfrog-cli-core
```

Afterward, execute `go mod tidy` to ensure the Go module files are updated. Note that Go will automatically adjust the version in the `go.mod` file.

#### Pull Requests

Once you have completed your coding changes, it is recommended to push the modifications made to the other modules first. Once these changes are pushed, you can update this project to resolve dependencies from your GitHub fork or branch. To achieve this, modify the `go.mod` file to point the dependency to your repository and branch, as shown in the example below:

```
replace github.com/jfrog/jfrog-cli-core/v2 => github.com/galusben/jfrog-cli-core/v2 dev
```

Finally, execute `go mod tidy` to update the Go module files. Please note that Go will automatically update the version in the `go.mod` file.

## Tests

### Usage

To run tests, use the following command:

```
go test -v github.com/jfrog/jfrog-cli [test-types] [flags]
```

The available flags are:

| Flag                | Description                                                                                     |
| ------------------- | ----------------------------------------------------------------------------------------------- |
| `-jfrog.url`        | [Default: http://localhost:8081] JFrog platform URL                                             |
| `-jfrog.user`       | [Default: admin] JFrog platform username                                                        |
| `-jfrog.password`   | [Default: password] JFrog platform password                                                     |
| `-jfrog.adminToken` | [Optional] JFrog platform admin token                                                           |
| `-ci.runId`         | [Optional] A unique identifier used as a suffix to create repositories and builds in the tests. |

The available test types are:

| Type                 | Description        |
| -------------------- | ------------------ |
| `-test.artifactory`  | Artifactory tests  |
| `-test.access`       | Access tests       |
| `-test.npm`          | Npm tests          |
| `-test.maven`        | Maven tests        |
| `-test.gradle`       | Gradle tests       |
| `-test.docker`       | Docker tests       |
| `-test.dockerScan`   | Docker scan tests  |
| `-test.podman`       | Podman tests       |
| `-test.go`           | Go tests           |
| `-test.pip`          | Pip tests          |
| `-test.pipenv`       | Pipenv tests       |
| `-test.poetry`       | Poetry tests       |
| `-test.nuget`        | Nuget tests        |
| `-test.plugins`      | Plugins tests      |
| `-test.distribution` | Distribution tests |
| `-test.transfer`     | Transfer tests     |
| `-test.xray`         | Xray tests         |
| `-test.xsc`          | Xsc tests          |

When running the tests, builds and repositories with timestamps will be created, for example: `cli-rt1-1592990748` and `cli-rt2-1592990748`. The content of these repositories will be deleted once the tests are completed.

#### Artifactory tests

In addition to the [general optional flags](#Usage), you can use the following optional Artifactory flags:

| Flag                   | Description                                                                                                     |
| ---------------------- | --------------------------------------------------------------------------------------------------------------- |
| `-jfrog.sshKeyPath`    | [Optional] Path to the SSH key file. Use this flag only if the Artifactory URL format is `ssh://[domain]:port`. |
| `-jfrog.sshPassphrase` | [Optional] Passphrase for the SSH key.                                                                          |

##### Examples

To run Artifactory tests, execute the following command:

```
go test -v github.com/jfrog/jfrog-cli -test.artifactory [flags]
```

#### Npm tests

##### Requirements

- The _npm_ executables should be included in the system's `PATH` environment variable.
- The tests are compatible with npm 7 and higher.

##### Limitations

- Currently, npm integration only supports HTTP(S) connections to Artifactory using username and password.

##### Examples

To run Npm tests, execute the following command:

```
go test -v github.com/jfrog/jfrog-cli -test.npm [flags]
```

#### Maven tests

##### Requirements

- The _java_ executable should be included in the system's `PATH` environment variable. Alternatively, set the `_JAVA_HOME` environment variable.

##### Limitations

- Currently, Maven integration only supports HTTP(S) connections to Artifactory using username and password.

##### Examples

To run Maven tests, execute the following command:

```
go test -v github.com/jfrog/jfrog-cli -test.maven [flags]
```

#### Gradle tests

##### Requirements

- The _gradle_ and _java_ executables should be included in the system's `PATH` environment variable. Alternatively, set the `JAVA_HOME` environment variable.

##### Limitations

- Currently, Gradle integration only supports HTTP(S) connections to Artifactory using username and password.

##### Examples

To run Gradle tests, execute the following command:

```
go test -v github.com/jfrog/jfrog-cli -test.gradle [flags]
```

#### Docker tests

##### Requirements

- Make sure the `RTLIC` environment variable is configured with a valid license.
- You can start an Artifactory container by running the `startArtifactory.sh` script located in the `testdata/docker/artifactory` directory. Before running the tests, wait for Artifactory to finish booting up in the container.

| Flag                      | Description                         |
| ------------------------- | ----------------------------------- |
| `-test.containerRegistry` | Artifactory Docker registry domain. |

##### Examples

To run Docker tests, execute the following command (replace the missing parameters as described below):

```
go test -v github.com/jfrog/jfrog-cli -test.docker [flags]
```

#### Podman tests

| Flag                      | Description                            |
| ------------------------- | -------------------------------------- |
| `-test.containerRegistry` | Artifactory container registry domain. |

##### Examples

To run Podman tests, execute the following command (replace the missing parameters as described below):

```
go test -v github.com/jfrog/jfrog-cli -test.podman [flags]
```

#### Go commands tests

#####

Requirements

- The tests are compatible with Artifactory 6.10 and higher.
- To run Go tests, use the following command:

```
go test -v github.com/jfrog/jfrog-cli -test.go [flags]
```

#### NuGet tests

##### Requirements

- Add the NuGet executable to the system's `PATH` environment variable.
- Run the following command:

```
go test -v github.com/jfrog/jfrog-cli -test.nuget [flags]
```

#### Pip tests

##### Requirements

- Add the Python and pip executables to the system's `PATH` environment variable.
- Run the following command:

```
go test -v github.com/jfrog/jfrog-cli -test.pip [flags]
```

#### Plugins tests

To run Plugins tests, execute the following command:

```
go test -v github.com/jfrog/jfrog-cli -test.plugins
```

#### Distribution tests

To run Distribution tests, execute the following command:

```
go test -v github.com/jfrog/jfrog-cli -test.distribution [flags]
```

#### Transfer tests

##### Requirement

The Transfer tests execute `transfer-files` commands between a local Artifactory server and a remote SaaS instance. In addition to the [general optional flags](#Usage), you _must_ use the following flags:

| Flag                               | Description                                                                                                                      |
| ---------------------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| `-jfrog.targetUrl`                 | JFrog target platform URL.                                                                                                       |
| `-jfrog.targetAdminToken`          | JFrog target platform admin token.                                                                                               |
| `-jfrog.jfrogHome`                 | The JFrog home directory of the local Artifactory installation.                                                                  |
| `-jfrog.installDataTransferPlugin` | Set this flag to `true` if you want the test to automatically install the data-transfer plugin in the source Artifactory server. |

To run Transfer tests, execute the following command:

```
go test -v github.com/jfrog/jfrog-cli -test.transfer [flags]
```

### Xray tests

To run Xray tests, execute the following command:

```
go test -v github.com/jfrog/jfrog-cli -test.xray -test.dockerScan [flags]
```
