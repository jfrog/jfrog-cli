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
$ go get github.com/JFrogDev/bintray-cli-go/...
```

Navigate to the following directory
```console
$ cd $GOPATH/bin
```
#### Usage

You can copy the *bt* executable to any location on your file-system as long as you add it to your *PATH* environment variable,
so that you can access it from any path.

##### Command syntax

```console
$ bt command-name global-options command-options arguments
```

The sections below specify the available commands, their respective options and additional arguments that may be needed.
*bt* should be followed by a command name (for example, download-ver), a list of options (for example, --repo=my-bintray-repo)
and the list of arguments for the command.

##### Global options

Global options are used for all commands.
```console
   --user            [Optional] Bintray username. It can be also set using the BINTRAY_USER environment variable. If not set, the subject sent as part of the command argument is used for authentication.
   --key             [Mandatory] Bintray API key. It can be also set using the BINTRAY_KEY environment variable.
   --api-url         [Default: https://api.bintray.com] Bintray API URL. It can be also set using the BINTRAY_API_URL environment variable.
   --download-url    [Default: https://dl.bintray.com] Bintray download server URL. It can be also set using the BINTRAY_DOWNLOAD_URL environment variable.
```

##### The *download-file* (dlf) command

###### Function
Used to download a specific file from Bintray.

###### Command options

Same as the *download-ver* command.

###### Arguments
The command expects one argument in the form of *subject/repository/package/version/path*.

###### Examples
```console
bt download-file my-org/swamp-repo/froggy-package/1.0/com/jfrog/bintray/crazy-frog.zip --user=my-user --key=my-api-key
bt dlf my-org/swamp-repo/froggy-package/1.0/com/jfrog/bintray/crazy-frog.zip --user=my-user --key=my-api-key
```

##### The *download-ver* (dlv) command

###### Function
Used to download the files of a specific version from Bintray.

###### Command options
This command has no command options. It uses however the global options.

###### Arguments
The command expects one argument in the form of *subject/repository/package/version*.

###### Examples
```console
bt download-ver my-org/swamp-repo/froggy-package/1.0 --user=my-user --key=my-api-key
bt dlv my-org/swamp-repo/froggy-package/1.0 --user=my-user --key=my-api-key
```

##### The *entitlement-keys* (ent-keys) command

###### Function
Used managing Download Keys.

###### Command options
The command uses the global options, in addition to the following command options.
```console
   --org             Bintray organization.
   --expiry          Download Key expiry (required for 'bt ent-keys show/create/update/delete'
   --ex-check-url    Used for Download Key creation and update. You can optionally provide an existence check directive, in the form of a callback URL, to verify whether the source identity of the Download Key still exists.
   --ex-check-cache  Used for Download Key creation and update. You can optionally provide the period in seconds for the callback URK results cache.
   --white-cidrs     Used for Download Key creation and update. Specifying white CIDRs in the form of 127.0.0.1/22,193.5.0.1/22 will allow access only for those IPs that exist in that address range.
   --black-cidrs     Used for Download Key creation and update. Specifying black CIDRs in the foem of 127.0.0.1/22,193.5.0.1/22 will block access for all IPs that exist in the specified range.
```

###### Arguments
* With no arguments, a list of all download keys is displayed.
* When sending the show, create, update or delete argument, it should be followed by an argument indicating the download key ID to show, create, update or delete the download key.

###### Examples
Show all Download Keys
```console
bt ent-keys
```
Create a Download Key
```console
bt ent-keys create key1 
bt ent-keys create key1 --expiry=7956915742000 
```

Show a specific Download Key
```console
bt ent-keys show key1
```

Update a Download Key
```console
bt ent-keys update key1 --ex-check-url=http://new-callback.com --white-cidrs=127.0.0.1/22,193.5.0.1/92 --black-cidrs=127.0.0.1/22,193.5.0.1/92
bt ent-keys update key1 --expiry=7956915752000
```

Delete a Download Key
```console
bt ent key delete key1
```

##### The *entitlements* (ent) command

###### Function
Used to manage Entitlements.

###### Command options
The command uses the global options, in addition to the following command.
```console
   --id              Entitlement ID. Used for Entitlements update.
   --access          Entitlement access. Used for Entitlements creation and update.
   --keys            Used for Entitlements creation and update. List of Download Keys in the form of \"key1\",\"key2\"...
   --path            Entitlement path. Used for Entitlements creating and update.
```

###### Arguments
* When sending an argument in the form of subject/repo or subject/repo/package or subject/repo/package/version, a list of all entitles for the send repo, package or version is displayed.
* If the first argument sent is show, create, update or delete, it should be followed by an argument in the form of subject/repo or subject/repo/package or subject/repo/package/version to show, create, update or delete the entitlement.

###### Examples

Show all Entitlements of the swamp-repo repository.
```console
bt ent my-org/swamp-repo
```

Show all Entitlements of the green-frog package.
```console
bt ent my-org/swamp-repo/green-frog
```

Show all Entitlements of version 1.0 of the green-frog package.
```console
bt ent my-org/swamp-repo/green-frog/1.0
```

Create an Entitlement for the green-frog package, with rw access, the key1 and key2 Download Keys and the a/b/c path.
```console
bt ent create my-org/swamp-repo/green-frog --access=rw --keys=key1,key2 --path=a/b/c
```

Show a specific Entitlement on the swamp-repo repository.
```console
bt ent show my-org/swamp-repo --id=451433e7b3ec3f18110ba770c77b9a3cb5534cfc
```

Update the download keys and access of an Entitlement on the swamp-repo repository.
```console
bt ent update my-org/swamp-repo --id=451433e7b3ec3f18110ba770c77b9a3cb5534cfc --keys=key1,key2 --access=r
```

Delete an Entitlement on the my-org/swamp-repo.
```console 
bt ent delete my-org/swamp-repo --id=451433e7b3ec3f18110ba770c77b9a3cb5534cfc
```