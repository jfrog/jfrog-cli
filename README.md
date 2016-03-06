
# Download and Installation

You can download the **JFrog CLI TBD link** executable directly from Bintray, or you can download the source files and build it yourself.

## Building the Executable

JFrog CLI is written in the [Go programming language](https://golang.org/), so to build the CLI yourself, you first need to have Go installed and configured on your machine.

### Setup Go

To download and install `Go`, please refer to the [Go documentation](https://golang.org/doc/install).

Navigate to the directory where you want to create the jfrog-cli-go **<TBD Link>** project, and set the value of the GOPATH environment variable to the full path of this directory.

### Download and Build the CLI

To download the jfrog-cli-go project, execute the following command:
````
$ go get github.com/JFrogDev/jfrog-cli-go/...
````
Go will download and build the project on your machine. Once complete, you will find the JFrog CLI executable under your `$GOPATH/bin` directory.

# Using JFrog CLI with Artifactory and Bintray
JFrog CLI can be used for quick and easy file management with both Artifactory and Bintray, and has a dedicated set of commands for each product. To learn how to use JFrog CLI, please refer to the relevant documentation through the corresponding link below: 
* [Using JFrog CLI with Artifactory](https://www.jfrog.com/confluence/display/RTF/JFrog+CLI)
* [Using JFrog CLI with Bintray](https://bintray.com/docs/usermanual/cli/cli_jfrogcli.html)