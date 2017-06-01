# Overview
JFrog CLI is a compact and smart client that provides a simple interface that automates access to *Artifactory*, *Bintray* and *Mission Control* through their respective REST APIs.
By using the JFrog CLI, you can greatly simplify your automation scripts making them more readable and easier to maintain.
Several features of the JFrog CLI makes your scripts more efficient and reliable:

- Multi-threaded upload and download of artifacts make builds run faster
- Checksum optimization reduces redundant file transfers
- Wildcards and regular expressions give you an easy way to collect all the artifacts you wish to upload or download.
- "Dry run" gives you a preview of file transfer operations before you actually run them

# Download and Installation

You can get the executable directly from the [JFrog CLI Download Page](https://www.jfrog.com/getcli/), or you can download the source files from this GitHub project and build it yourself.

On Mac you can run:
````
brew install jfrog-cli-go
````

# Building the Executable

JFrog CLI is written in the [Go programming language](https://golang.org/), so to build the CLI yourself, you first need to have Go installed and configured on your machine.

## Setup Go

To download and install `Go`, please refer to the [Go documentation](https://golang.org/doc/install).
Please download `Go 1.7` or above.

Navigate to the directory where you want to create the jfrog-cli-go project, and set the value of the GOPATH environment variable to the full path of this directory.

## Download and Build the CLI

To download the jfrog-cli-go project, execute the following command:
````
go get github.com/jfrogdev/jfrog-cli-go/jfrog
````
Go will download and build the project on your machine. Once complete, you will find the JFrog CLI executable under your `$GOPATH/bin` directory.

## Build a modified version hosted in another git repo
- Checkout your repo e.g. to "...\Repos\jfrog-cli-go\src\github.com\jfrogdev\jfrog-cli-go\"
- Do your modifications.
- cd into the jfrog dir and do a "go install", this will create a jfrog.exe in "Repos\jfrog-cli-go\bin"



# Tests

### Artifactory Integration tests
To run Artifactory integration tests execute the following command: 
````
go test -v github.com/jfrogdev/jfrog-cli-go/jfrog -test.artifactory=true -test.bintray=false
````
Optional flags:

| Flag | Description |
| --- | --- |
| `-rt.url` | [Default: http://localhost:8081/artifatory] Artifactory URL. |
| `-rt.user` | [Default: admin] Artifactory username. |
| `-rt.password` | [Default: password] Artifactory password. |
| `-rt.apikey` | [Optional] Artifactory API key. |

* Artifactory url: http://localhost:8081/artifatory
* User: admin
* Password: password

* Running the tests will create two repositories: `jfrog-cli-tests-repo` and `jfrog-cli-tests-repo1`.<br/>
  Once the tests are completed, the content of these repositories will be deleted.

### Bintray Integration tests
Bintray tests credentials are taken from the CLI configuration. If non configured or not passed as flags, the tests will fail.

To run Bintray tests execute the following command: 
````
go test -v github.com/jfrogdev/jfrog-cli-go/jfrog -test.artifactory=false -test.bintray=true
````
Flags:

| Flag | Description |
| --- | --- |
| `-bt.user` | [Mandatory if not configured] Bintray username. |
| `-bt.key` | [Mandatory if not configured] Bintray API key |

* Running the tests will create a repository `jfrog-cli-tests-repo1` in bintray.<br/>
  Once the tests are completed, the repository will be deleted.

### Unit tests
To execute all the JFrog CLI unit tests run the following command:
#### Windows
````
jfrogdev\jfrog-cli-go> for /f "" %G in ('go list ./... ^| find /i /v "/vendor/" ^| find /i /v "jfrog-cli-go/jfrog"') do @go test %G
````

#### Unix
```
jfrogdev/jfrog-cli-go$ go test $(go list ./... | grep -v vendor | grep -v jfrog-cli-go/jfrog)
```

# Pull Requests
We welcome pull requests.

# Using JFrog CLI
JFrog CLI can be used for a variety of functions with Artifactory, Bintray, Xray and Mission Control,
and has a dedicated set of commands for each product.
To learn how to use JFrog CLI, please visit the [JFrog CLI User Guide](https://www.jfrog.com/confluence/display/CLI/Welcome+to+JFrog+CLI).

