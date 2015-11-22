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

You can copy the *bt* executable to any location on your file-system as long as you add it to your *PATH* environment variable,
so that you can access it from any path.

##### Command syntax

` ``console
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

##### The *entitlements* (ent) command

###### Function
Used to manage Entitlements and Download Keys.

###### Command options
The command uses the global options, in addition to the following command.
```console
   --id                  Entitlement ID. Used for Entitlements update.
   --access              Entitlement access. Used for Entitlements creation and update.
   --keys                Used for Entitlements creation and update. List of Download Keys in the form of \"key1\",\"key2\"...
   --path                Entitlement path. Used for Entitlements creating and update.
   --org                 Bintray organization.
   --key-id              Download Key ID (required for 'bt entitlements key show/create/update/delete'
   --key-expiry          Download Key expiry (required for 'bt entitlements key show/create/update/delete'
   --key-ex-check-url    Used for Download Key creation and update. You can optionally provide an existence check directive, in the form of a callback URL, to verify whether the source identity of the Download Key still exists.
   --key-ex-check-cache  Used for Download Key creation and update. You can optionally provide the period in seconds for the callback URK results cache.
   --key-white-cidrs     Used for Download Key creation and update. Specifying white CIDRs in the form of 127.0.0.1/22,193.5.0.1/22 will allow access only for those IPs that exist in that address range.
   --key-black-cidrs     Used for Download Key creation and update. Specifying black CIDRs in the foem of 127.0.0.1/22,193.5.0.1/22 will block access for all IPs that exist in the specified range.
```

###### Arguments
* When sending an argument in the form of subject/repo or subject/repo/package or subject/repo/package/version, a list of all entitles for the send repo, package or version is displayed.
* If the first argument sent is show, create, update or delete, it should be followed by an argument in the form of subject/repo or subject/repo/package or subject/repo/package/version to show, create, update or delete the entitlement.
* If the argument *keys* is sent, the command displayes a list of all download keys.
* If the argument *key* is sent, then it should be followed by one of the following arguments: show, create, update or delete.

###### Examples

Show all Entitlements of the swamp-repo repository.
```console
bt ent my-org/swamp-repo
```

Show all Entitlements of the green-frog package.
```console
bt ent my-org/maven/green-frog
```

Show all Entitlements of the green-frog package.
```console
bt ent my-org/maven/green-frog
```

Show all Entitlements of version 1.0 of the green-frog package.
```console
bt ent my-org/maven/green-frog/1.0
```

Create an Entitlement for the green-frog package, with rw access, the key1 and key2 Download Keys and the a/b/c path.
```console
bt ent create my-org/maven/green-frog --access=rw --keys=key1,key2 --path=a/b/c
```

TODO: bt ent show

TODO: bt ent update

TODO: bt ent delete

Show all Download Keys
```console
bt ent keys
```
Create a Download Key
```console
bt ent key create --key-id=key1 --key-expiry=7956915742000 --key-ex-check-url=http://callback.com --key-white-cidrs=127.0.0.1/22,193.5.0.1/92 --key-black-cidrs=127.0.0.1/22,193.5.0.1/92
```

Show a specific Download Key
```console
bt ent key show --key-id=key1
```

Update a Download Key
```console
bt ent key update --key-id=key1 --key-expiry=7956915742000 --key-ex-check-url=http://new-callback.com --key-white-cidrs=127.0.0.1/22,193.5.0.1/92 --key-black-cidrs=127.0.0.1/22,193.5.0.1/92
```

Delete a Download Key
```console
bt ent key delete --key-id=key1
```