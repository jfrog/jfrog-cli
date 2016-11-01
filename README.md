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
$ brew install jfrog-cli-go
````

## Building the Executable

JFrog CLI is written in the [Go programming language](https://golang.org/), so to build the CLI yourself, you first need to have Go installed and configured on your machine.

### Setup Go

To download and install `Go`, please refer to the [Go documentation](https://golang.org/doc/install).
Please download `Go 1.6` or above.

Navigate to the directory where you want to create the jfrog-cli-go project, and set the value of the GOPATH environment variable to the full path of this directory.

### Download and Build the CLI

To download the jfrog-cli-go project, execute the following command:
````
$ go get github.com/jfrogdev/jfrog-cli-go/...
````
Go will download and build the project on your machine. Once complete, you will find the JFrog CLI executable under your `$GOPATH/bin` directory.

### Integration tests
To execute integration tests, run the following commands:
````
$ go test -v github.com/jfrogdev/jfrog-cli-go/jfrog
````

By default, these tests use the following settings:

* Artifactory url: http://localhost:8081/artifatory
* User: admin
* Password: password

These settings can be customized using the *--url*, *--user*, *--password* and *--apikey* flags while executing the go test command.<br/>
For example:
````
$ go test -v github.com/jfrogdev/jfrog-cli-go/jfrog --url=http://yourArtifactoryUrl/artifactory --user=user --password=password --apikey=apikey
````
* Adding the *--apikey* flag triggers specific tests for the API Key functionality, in addition to other tests which use user and password.
* Running the tests will create two repositories: jfrog-cli-tests-repo and jfrog-cli-tests-repo1.<br/>
  Once the tests are completed, the content of these repositories will be deleted.

# Using JFrog CLI with Artifactory, Bintray and Mission Control
JFrog CLI can be used for a variety of functions with Artifactory, Bintray, Xray and Mission Control,
and has a dedicated set of commands for each product.
To learn how to use JFrog CLI, please visit the [JFrog CLI User Guide](https://www.jfrog.com/confluence/display/CLI/Welcome+to+JFrog+CLI).

