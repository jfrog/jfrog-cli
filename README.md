# Table of Contents

- [Overview](#overview)
- [Download and Installation](#download-and-installation)
- [Building the Executable](#building-the-executable)
- [](#tests) [Tests](./TESTS.md)
- [Code Contributions](#code-contributions)
- [Using JFrog CLI](#using-jfrog-cli)
- [JFrog CLI Plugins](#jfrog-cli-plugins)
- [Release Notes](#release-notes)

# Overview

JFrog CLI is a compact and smart client that provides a simple interface that automates access to _Artifactory_, _Bintray_ and _Mission Control_ through their respective REST APIs.
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
