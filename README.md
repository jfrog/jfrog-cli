## Bintray CLI

Bintray CLI provides a command line interface for invoking actions on Bintray.

### Getting Started

#### Downloading the executables from Bintray

[TO DO]

#### Building the command line executable

If you prefer, you may instead build the client in go.

##### Setup GO on your machine

* Make sure you have a working Go environment. [See the install instructions](http://golang.org/doc/install).
* Navigate to the directory where you want to create the *bintray-cli-go* project.
* Set the value of the GOPATH environment variable to the full path of this directory.

##### Download Bintray CLI from GitHub

Run the following command to create the *artifactory-cli-go* project:
```console
$ go get github.com/JFrogDev/bintray-cli-go
```

Navigate to the following directory
```console
$ cd $GOPATH/bin
```
#### Usage

You can copy the *btr* executable to any location on your file-system as long as you add it to your *PATH* environment variable,
so that you can access it from any path.

##### Command syntax

```console
$ btr command-name options arguments
```

The sections below specify the available commands, their respective options and additional arguments that may be needed.
*btr* should be followed by a command name (for example, download-ver), a list of options (for example, --repo=my-bintray-repo)
and the list of arguments for the command.

##### The *download-ver* (dv) command

###### Function
Used to download the files of a specific version from Bintray.

###### Command options
```console
   --user            [Mandatory] Bintray username. It can be also set using the BINTRAY_USER environment variable.
   --key             [Mandatory] Bintray API key. It can be also set using the BINTRAY_KEY environment variable.
   --org             [Optional] Bintray organization. It can be also set using the BINTRAY_ORG environment variable.
   --api-url         [Default: https://api.bintray.com] Bintray API URL. It can be also set using the BINTRAY_API_URL environment variable.
   --download-url    [Default: https://dl.bintray.com] Bintray download server URL. It can be also set using the BINTRAY_DOWNLOAD_URL environment variable.
   --repo            [Mandatory] Bintray repository.
   --package         [Mandatory] Bintray package.
```

###### Arguments
The command expects one argument - the package version.

###### Examples
```console
btr download-ver 3.8 --repo=my-swamp --package=green-frog --user=my-user --key=my-api-key
```

##### The *download-file* (df) command

###### Function
Used to download a specific file from Bintray.

###### Command options

Same as the *download-ver* command.

###### Arguments
The command expects one argument - the Bintray file path.

###### Example
```console
btr download-file /com/jfrog/bintray/example/3.8/froggy.zip --repo=my-swamp --package=green-frog --user=my-user --key=my-api-key
```