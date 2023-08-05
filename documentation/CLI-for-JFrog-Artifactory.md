# JFrog CLI : CLI for JFrog Artifactory

## Overview

This page describes how to use JFrog CLI with JFrog Artifactory.

Read more about JFrog CLI [here](https://jfrog.com/help/r/jfrog-cli/environment-variables).

## Environment Variables

The Artifactory upload command makes use of the following environment variable:

|                                                  |                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
|--------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Variable Name**                                | **Description**                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| **JFROG_CLI_MIN\_CHECKSUM\_DEPLOY\_SIZE\_KB**    | \[Default: 10\]<br><br>Minimum file size in KB for which JFrog CLI performs checksum deploy optimization.                                                                                                                                                                                                                                                                                                                                             |
| **JFROG_CLI_RELEASES_REPO**                      | Configured Artifactory repository name from which to download the jar needed by the mvn/gradle command.<br> This environment variable's value format should be `<server ID configured by the 'jf c add' command>/<repo name>`.<br> The repository should proxy https://releases.jfrog.io.<br> This environment variable is used by the 'jf mvn' and 'jf gradle' commands, and also by the 'jf audit' command, when used for maven or gradle projects. |
| **JFROG_CLI_DEPENDENCIES_DIR**                   | \[Default: $JFROG_CLI_HOME_DIR/dependencies\]<br><br>Defines the directory to which JFrog CLI's internal dependencies are downloaded.                                                                                                                                                                                                                                                                                                                 |
| **JFROG_CLI_REPORT_USAGE**                       | \[Default: true\]<br><br>Set to false to block JFrog CLI from sending usage statistics to Artifactory.                                                                                                                                                                                                                                                                                                                                                |
| **JFROG_CLI_SERVER_ID**                          | Server ID configured using the config command, unless sent as a command argument or option.                                                                                                                                                                                                                                                                                                                                                           |
| **JFROG_CLI_BUILD_NAME**                         | Build name to be used by commands which expect a build name, unless sent as a command argument or option.                                                                                                                                                                                                                                                                                                                                             |
| **JFROG_CLI_BUILD_NUMBER**                       | Build number to be used by commands which expect a build number, unless sent as a command argument or option.                                                                                                                                                                                                                                                                                                                                         |
| **JFROG_CLI_BUILD_PROJECT**                      | JFrog project key to be used by commands which expect build name and build number. Determines the project of the published build.                                                                                                                                                                                                                                                                                                                     |
| **JFROG_CLI_BUILD_URL**                          | Sets the CI server build URL in the build-info. The "jf rt build-publish" command uses the value of this environment variable, unless the --build-url command option is sent.                                                                                                                                                                                                                                                                         |
| **JFROG_CLI_ENV_EXCLUDE**                        | \[Default: \*password\*;\*secret\*;\*key\*;\*token\*\]<br><br> List of case insensitive patterns in the form of "value1;value2;...". Environment variables match those patterns will be excluded. This environment variable is used by the "jf rt build-publish" command, in case the --env-exclude command option is not sent.                                                                                                                       |
| **JFROG_CLI_TRANSITIVE\_DOWNLOAD\_EXPERIMENTAL** | \[Default: false\]<br><br>Used by the "jf rt download" command. Set to true to download artifacts also from remote repositories. This feature is experimental and available on Artifactory version 7.17.0 or higher.`                                                                                                                                                                                                                                 |

---
**Note**
> Read about additional environment variables at the [Welcome to JFrog CLI](https://jfrog.com/help/r/jfrog-cli/environment-variables) page.
---
  

## Authentication

When used with Artifactory, JFrog CLI offers several means of authentication: JFrog CLI does not support accessing  Artifactory without authentication.

### Authenticating with Username and Password / API Key

To authenticate yourself using your JFrog login credentials, either configure your credentials once using the **jf c add** command or provide the following option to each command.

| Command option | Description                                                           |
|----------------|-----------------------------------------------------------------------|
| --url          | JFrog Artifactory API endpoint URL. It usually ends with /artifactory |
| --user         | JFrog username                                                        |
| --password     | JFrog password or API key                                             |

For enhanced security, when JFrog CLI is configured to use a username and password / API key, it automatically generates an access token to authenticate with Artifactory. The generated access token is valid for one hour only. JFrog CLI automatically refreshed the token before it expires. The **jfrog c add** command allows disabling this functionality. This feature is currently not supported by commands which use external tools or package managers or work with JFrog Distribution.

### Authenticating with an Access Token

To authenticate yourself using an Artifactory Access Token, either configure your Access Token once using the **jf c add** command or provide the following option to each command.

| Command option | Description                                                           |
|----------------|-----------------------------------------------------------------------|
| --url          | JFrog Artifactory API endpoint URL. It usually ends with /artifactory |
| --access-token | JFrog access token                                                    |

### Authenticating with RSA Keys

---
**Note**
> Currently, authentication with RSA keys is not supported when working with external package managers and build tools (Maven, Gradle, Npm, Docker, Go and NuGet) or with the cUrl integration.
---

From version 4.4, Artifactory supports SSH authentication using RSA public and private keys. To authenticate yourself to Artifactory using RSA keys, execute the following instructions:

* Enable SSH authentication as described in [Configuring SSH](https://jfrog.com/help/r/jfrog-platform-administration-Documentation/Managing-Ssh-Keys).
* Configure your Artifactory URL to have the following format: `ssh://[host]:[port]  
    `There are two ways to do this:  
    
    * For each command, use the `--url` command option.
    * Specify the Artifactory URL in the correct format using the **jfrog c add** command.
    
    ---
    **Warning** <br><br>
    **Don't include your Artifactory context URL**
    
    > Make sure that the \[host\] component of the URL only includes the hostname or the IP, but not your Artifactory context URL.
    ---
    
* Configure the path to your SSH key file. There are two ways to do this:
    * For each command, use the `--ssh-key-path` command option.
    * Specify the path using the **jfrog c add** command.

### Authenticating using Client Certificates (mTLS)

From Artifactory release 7.38.4, you can authenticate users using a client certificate ([mTLS](https://en.wikipedia.org/wiki/Mutual_authentication#mTLS)). To do so will require a reverse proxy and some setup on the front reverse proxy (Nginx). Read about how to set this up [here](https://jfrog.com/help/r/jfrog-artifactory-documentation/Http-Settings).

To authenticate with the proxy using a client certificate, either configure your certificate once using the **jf c add** command or use the --`client-cert-path` and`--client-cert-ket-path` command options with each command.

---
**Note**
> Authentication using client certificates (mTLS) is not supported by commands which integrate with package managers. 
---
  

Not Using a Public CA (Certificate Authority)?

This section is relevant for you if you're not using a public CA (Certificate Authority) to issue the SSL certificate used to connect to your Artifactory domain. You may not be using a public CA either because you're using self-signed certificates or you're running your own PKI services in-house (often by using a Microsoft CA).

In this case, you'll need to make those certificates available for JFrog CLI, by placing them inside the **security/certs** directory, which is under JFrog CLI's home directory. By default, the home directory is **~/.jfrog**, but it can be also set using the **JFROG_CLI_HOME_DIR** environment variable.

**Note**
1.  The supported certificate format is PEM.
2.  Some commands support the **--insecure-tls** option, which skips the TLS certificates verification.
3.  Before version 1.37.0, JFrog CLI expected the certificates to be located directly under the **security** directory. JFrog CLI will automatically move the certificates to the new directory when installing version 1.37.0 or above. Downgrading back to an older version requires replacing the configuration directory manually. You'll find a backup if the old configuration under **.jfrog/backup** 


## Storing Symlinks in Artifactory

JFrog CLI lets you upload and download artifacts from your local file system to Artifactory, this also includes uploading symlinks (soft links).

Symlinks are stored in Artifactory as files with a zero size, with the following properties:  
**symlink.dest** - The actual path on the original filesystem to which the symlink points  
**symlink.destsha1** - the SHA1 checksum of the value in the **symlink.dest** property

To upload symlinks, the `jf rt upload` command should be executed with the `--symlinks` option set to true.

When downloading symlinks stored in Artifactory, the CLI can verify that the file to which the symlink points actually exists and that it has the correct SHA1 checksum. To add this validation, you should use the `--validate-symlinks` option with the `jf rt download` command.

* * *

## Using Placeholders

The JFrog CLI offers enormous flexibility in how you **download, upload**, **copy**, or **move** files through the use of wildcard or regular expressions with placeholders.

Any wildcard enclosed in parentheses in the source path can be matched with a corresponding placeholder in the target path to determine the name of the artifact once uploaded.

#### Examples

##### **Example 1: Upload all files to the target repository**

For each .tgz file in the source directory, create a corresponding directory with the same name in the target repository and upload it there. For example, a file named **froggy.tgz** should be uploaded to **my-local-rep/froggy**. **froggy** will be created in a folder in Artifactory).
```
jf rt u "(*).tgz" my-local-repo/{1}/ --recursive=false
```

##### **Example 2: Upload all files sharing the same prefix to the target repository**

Upload all files whose name begins with "frog" to folder **frogfiles** in the target repository, but append its name with the text "-up". For example, a file called **froggy.tgz** should be renamed **froggy.tgz-up**.
```
jf u "(frog*)" my-local-repo/frogfiles/{1}-up --recursive=false
```

##### **Example 3: Upload all files to corresponding directories according to extension type**

Upload all files in the current directory to the **my-local-repo** repository and place them in directories that match their file extensions.
```
jf rt u "(*).(*)" my-local-repo/{2}/{1}.{2} --recursive=false
```

##### **Example 4: Copy all zip files to target repository and append with an extension.**

Copy all zip files under /rabbit in the **source-frog-repo** repository into the same path in the **target-frog-repo** repository and append the copied files' names with ".cp".
```
jf rt cp "source-frog-repo/rabbit/(*.zip)" target-frog-repo/rabbit/{1}.cp
```

## General Commands

The following sections describe the commands available in the JFrog CLI for use with Artifactory.

### Verifying Artifactory is Accessible

This command can be used to verify that Artifactory is accessible by sending an applicative ping to Artifactory.

|                   |                                                                                                                                               |
|-------------------|-----------------------------------------------------------------------------------------------------------------------------------------------|
| Command name      | rt ping                                                                                                                                       |
| Abbreviation      | rt p                                                                                                                                          |
|                   |                                                                                                                                               |
| Command options   |                                                                                                                                               |
| --url             | \[Optional\]<br><br>Artifactory URL.                                                                                                          |
| --server-id       | \[Optional\]<br><br>Server ID configured using the **jf c add** command. If not specified, the default configured Artifactory server is used. |
| --insecure-tls    | \[Default: false\]<br><br>Set to true to skip TLS certificates verification.                                                                  |
| Command arguments | The command accepts no arguments.                                                                                                             |

#### **Examples**

##### **Example 1**

Ping the configured default Artifactory server.
```
jf rt ping
```
  
##### **Example 2**

Ping the configured Artifactory server with ID **rt-server-1**.
```
jf rt ping --server-id=rt-server-1
```

##### **Example 3**

Ping the Artifactory server. accessible through the specified URL.
```
jf rt ping --url=https://my-rt-server.com/artifactory
```

### Uploading Files

This command is used to upload files to Artifactory.

|                    |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
|--------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command name       | rt upload                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| Abbreviation       | rt u                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| Command options    | **Warning**<br><br> When using the * or ; characters in the upload command options or arguments, make sure to wrap the whole options or arguments string in quotes (") to make sure the * or ; characters are not interpreted as literals.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| --archive          | \[Optional\]<br><br>Set to "zip" to pack and deploy the files to Artifactory inside a ZIP archive. Currently, the only packaging format supported is zip.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| --server-id        | \[Optional\]<br><br>Server ID configured using the **jf c add** command. If not specified, the default configured Artifactory server is used.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| --spec             | \[Optional\]<br><br>Path to a file spec. For more details, please refer to [Using File Specs](https://jfrog.com/help/r/jfrog-cli/using-file-specs).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| --spec-vars        | \[Optional\]<br><br>List of variables in the form of "key1=value1;key2=value2;..." to be replaced in the File Spec. In the File Spec, the variables should be used as follows: ${key1}.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| --build-name       | \[Optional\]<br><br>Build name. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| --build-number     | \[Optional\]<br><br>Build number. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| --project          | \[Optional\]<br><br>JFrog project key.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| --module           | \[Optional\]<br><br>Optional module name for the build-info.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| --target-props     | \[Optional\]<br><br>A list of Artifactory [properties](https://jfrog.com/help/r/jfrog-artifactory-documentation/Working-With-Jfrog-Properties) specified as "key=value" pairs separated by a semi-colon ( ; ) to be attached to the uploaded files. If any key can take several values, then each value is separated by a comma ( , ). For example, "key1=value1;key2=value21,value22;key3=value3".                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| --deb              | \[Optional\]<br><br>Used for Debian packages only. Specifies the distribution/component/architecture of the package. If the the value for distribution, component or architecture include a slash. the slash should be escaped with a back-slash.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| --flat             | \[Default: false\]<br><br>If true, files are uploaded to the exact target path specified and their hierarchy in the source file system is ignored.<br><br>If false, files are uploaded to the target path while maintaining their file system hierarchy.<br><br>If [Using Placeholders](https://jfrog.com/help/r/jfrog-cli/using-placeholders) are used, the value of this option is ignored.<br><br>**Note**<br><br>**JFrog CLI v1**<br><br>In JFrog CLI v1, the default value of the --flat option is true.                                                                                                                                                                                                                                                                                                                                                                                                                 |
| --recursive        | \[Default: true\]<br><br>If true, files are also collected from sub-folders of the source directory for upload .<br><br>If false, only files specifically in the source directory are uploaded.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| --regexp           | \[Default: false\]<br><br>If true, the command will interpret the first argument, which describes the local file-system path of artifacts to upload, as a regular expression.<br><br>If false, it will interpret the first argument as a wild-card expression.<br><br>The above also applies for the --exclusions option.<br><br>If you have specified that you are using regular expressions, then the beginning of the expression must be enclosed in parenthesis. For example: **a/b/c/(.*)/file.zip**                                                                                                                                                                                                                                                                                                                                                                                                                     |
| --ant              | \[Default: false\]<br><br>If true, the command will interpret the first argument, which describes the local file-system path of artifacts to upload, as an ANT pattern.<br><br>If false, it will interpret the first argument as a wildcards expression.<br><br>The above also applies for the --exclusions option.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| --threads          | \[Default: 3\]<br><br>The number of parallel threads that should be used to upload where each thread uploads a single artifact at a time.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| --dry-run          | \[Default: false\]<br><br>If true, the command only indicates which artifacts would have been uploaded<br><br>If false, the command is fully executed and uploads artifacts as specified                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| --symlinks         | \[Default: false\]<br><br>If true, the command will preserve the soft links structure in Artifactory. The `symlink` file representation will contain the symbolic link and checksum properties.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| --explode          | \[Default: false\]<br><br>If true, the command will extract an archive containing multiple artifacts after it is deployed to Artifactory, while maintaining the archive's file structure.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| --include-dirs     | \[Default: false\]<br><br>If true, the source path applies to bottom-chain directories and not only to files. Bottom-chain directories are either empty or do not include other directories that match the source path.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| --exclusions       | \[Optional\]<br><br>A list of Semicolon-separated exclude patterns. Allows using wildcards, regular expressions or ANT patterns, according to the value of the **--regexp** and **--ant** options. Please read the **--regexp** and **--ant** options description for more information.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| --sync-deletes     | \[Optional\]<br><br>Specific path in Artifactory, under which to sync artifacts after the upload. After the upload, this path will include only the artifacts uploaded during this upload operation. The other files under this path will be deleted.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| --quiet            | \[Default: false\]<br><br>If true, the delete confirmation message is skipped.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| --fail-no-op       | \[Default: false\]<br><br>Set to true if you'd like the command to return exit code 2 in case of no files are affected.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| --retries          | \[Default: 3\]<br><br>Number of upload retries.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| --retry-wait-time  | \[Default: 0s\]<br><br>Number of seconds or milliseconds to wait between retries. The numeric value should either end with s for seconds or ms for milliseconds.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| --detailed-summary | \[Default: false\]<br><br>Set to true to include a list of the affected files as part of the command output summary.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| --insecure-tls     | \[Default: false\]<br><br>Set to true to skip TLS certificates verification.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| Command arguments  | The command takes two arguments.<br><br>In case the --spec option is used, the commands accept no arguments.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| Source path        | The first argument specifies the local file system path to artifacts that should be uploaded to Artifactory. You can specify multiple artifacts by using wildcards or a regular expression as designated by the **--regexp** command option. Please read the **--regexp** option description for more information.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| Target path        | The second argument specifies the target path in Artifactory in the following format: `[repository name]/[repository path]`<br><br>If the target path ends with a slash, the path is assumed to be a folder. For example, if you specify the target as "repo-name/a/b/", then "b" is assumed to be a folder in Artifactory into which files should be uploaded. If there is no terminal slash, the target path is assumed to be a file to which the uploaded file should be renamed. For example, if you specify the target as "repo-name/a/b", the uploaded file is renamed to "b" in Artifactory.<br><br>For flexibility in specifying the upload path, you can include placeholders in the form of {1}, {2} which are replaced by corresponding tokens in the source path that are enclosed in parenthesis. For more details, please refer to [Using Placeholders](https://jfrog.com/help/r/jfrog-cli/using-placeholders). |

#### Examples

##### **Example 1**

Upload a file called **froggy.tgz** to the root of the **my-local-repo** repository.
```
jf rt u froggy.tgz my-local-repo
```

##### **Example 2**

Collect all the zip files located under the **build** directory (including subdirectories), and upload them to the **my-local-repo** repository, under the **zipFiles** folder, while maintaining the original names of the files.
```
jf rt u "build/*.zip" my-local-repo/zipFiles/
```

##### **Example 3**

Collect all the zip files located under the **build** directory (including subdirectories), and upload them to the **my-local-repo** repository, under the **zipFiles** folder, while maintaining the original names of the files. Also delete all files in the **my-local-repo** repository, under the **zipFiles** folder, except for the files which were uploaded by this command.
```
jf rt u "build/*.zip" my-local-repo/zipFiles/ --sync-deletes="my-local-repo/zipFiles/"
```

##### **Example 4**

Collect all files located under the **build** directory (including subdirectories), and upload them to the **my-release-local **repository, under the **files** folder, while maintaining the original names of the artifacts. Exclude (do not upload) files, which include **install** as part of their path, and have the **pack** extension. This example uses a wildcard pattern. See **Example 5**, which uses regular expressions instead.
```
jf rt u "build/" my-release-local/files/ --exclusions="\*install\*pack*"
```

##### **Example 5**

Collect all files located under the **build** directory (including subdirectories), and upload them to the **my-release-local** repository, under the **files** folder, while maintaining the original names of the artifacts. Exclude (do not upload) files, which include **install** as part of their path, and have the **pack** extension. This example uses a regular expression. See **Example 4**, which uses a wildcard pattern instead.
```
jf rt u "build/" my-release-local/files/ --regexp --exclusions="(.*)install.*pack$"
```

##### **Example 6**

Collect all files located under the **build** directory and match the **/*.zip** ANT pattern, and upload them to the **my-release-local** repository, under the **files** folder, while maintaining the original names of the artifacts.
```
jf rt u "build/**/*.zip" my-release-local/files/ --ant
```

##### **Example 7**

Package all files located under the **build** directory (including subdirectories) into a zip archive named **archive.zip** , and upload the archive to the **my-local-repo** repository,
```
jf rt u "build/" my-local-repo/my-archive.zip --archive zip
```

### Downloading Files

This command is used to download files from Artifactory.

> Download from Remote Repositories: <br><br>By default, the command only downloads files that are cached on the current Artifactory instance. It does not download files located on remote Artifactory instances, through remote or virtual repositories. To allow the command to download files from remote Artifactory instances, which are proxied by the use of remote repositories, set the **JFROG_CLI_TRANSITIVE_DOWNLOAD_EXPERIMENTAL** environment variable to **true**. This functionality requires version 7.17 or above of Artifactory. The remote download functionality is supported only on remote repositories which proxy repositories on remote Artifactory instances. Downloading through a remote repository that proxies non-Artifactory repositories is not supported. 

|                     |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
|---------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command name        | rt download                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| Abbreviation        | rt dl                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| Command options     | **Warning** <br><br>When using the * or ; characters in the download command options or arguments, make sure to wrap the whole options or arguments string in quotes (") to make sure the * or ; characters are not interpreted as literals.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| --server-id         | \[Optional\]<br><br>Server ID configured using the config command. If not specified, the default configured Artifactory server is used.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| --build-name        | \[Optional\]<br><br>Build name. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| --build-number      | \[Optional\]<br><br>Build number. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| --project           | \[Optional\]<br><br>JFrog project key.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| --module            | \[Optional\]<br><br>Optional module name for the build-info.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| --spec              | \[Optional\]<br><br>Path to a file spec. For more details, please refer to [Using File Specs](https://jfrog.com/help/r/jfrog-cli/using-file-specs).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| --spec-vars         | \[Optional\]<br><br>List of variables in the form of "key1=value1;key2=value2;..." to be replaced in the File Spec. In the File Spec, the variables should be used as follows: ${key1}.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| --props             | \[Optional\]<br><br>A list of Artifactory [properties](https://jfrog.com/help/r/jfrog-artifactory-documentation/Working-With-Jfrog-Properties) specified as "key=value" pairs separated by a semi-colon (for example, "key1=value1;key2=value2;key3=value3"). Only artifacts with **all** of the specified properties names and values will be downloaded.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| --exclude-props     | \[Optional\]<br><br>A list of Artifactory [properties](https://jfrog.com/help/r/jfrog-artifactory-documentation/Working-With-Jfrog-Properties) specified as "key=value" pairs separated by a semi-colon (for example, "key1=value1;key2=value2;key3=value3"). Only artifacts **without all** of the specified properties names and values will be downloaded.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| --build             | \[Optional\]<br><br>If specified, only artifacts of the specified build are matched. The property format is build-name/build-number. If you do not specify the build number, the artifacts are filtered by the latest build number.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| --bundle            | \[Optional\]<br><br>If specified, only artifacts of the specified bundle are matched. The value format is bundle-name/bundle-version.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| --flat              | \[Default: false\]<br><br>If true, artifacts are downloaded to the exact target path specified and their hierarchy in the source repository is ignored.<br><br>If false, artifacts are downloaded to the target path in the file system while maintaining their hierarchy in the source repository.<br><br>If [Using Placeholders](https://jfrog.com/help/r/jfrog-cli/using-placeholders) are used, and you would like the local file system (download path) to be determined by placeholders only, or in other words, avoid concatenating the Artifactory folder hierarchy local, set to false.                                                                                                                                                                                                                                                       |
| --recursive         | \[Default: true\]<br><br>If true, artifacts are also downloaded from sub-paths under the specified path in the source repository.<br><br>If false, only artifacts in the specified source path directory are downloaded.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| --threads           | \[Default: 3\]<br><br>The number of parallel threads that should be used to download where each thread downloads a single artifact at a time.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| --split-count       | \[Default: 3\]<br><br>The number of segments into which each file should be split for download (provided the artifact is over `--min-split` in size). To download each file in a single thread, set to 0.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| --retries           | \[Default: 3\]<br><br>Number of download retries.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| --retry-wait-time   | \[Default: 0s\]<br><br>Number of seconds or milliseconds to wait between retries. The numeric value should either end with s for seconds or ms for milliseconds.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| --min-split         | \[Default: 5120\]<br><br>The minimum size permitted for splitting. Files larger than the specified number will be split into equally sized `--split-count` segments. Any files smaller than the specified number will be downloaded in a single thread. If set to -1, files are not split.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| --dry-run           | \[Default: false\]<br><br>If true, the command only indicates which artifacts would have been downloaded.<br><br>If false, the command is fully executed and downloads artifacts as specified.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| --explode           | \[Default: false\]<br><br>Set to true to extract an archive after it is downloaded from Artifactory.<br><br>Supported compression formats: br, bz2, gz, lz4, sz, xz, zstd.<br><br>Supported archive formats: zip, tar (including any compressed variants like tar.gz), rar.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| --validate-symlinks | \[Default: false\]<br><br>If true, the command will validate that **symlinks** are pointing to existing and unchanged files, by comparing their sha1. Applicable to files and not directories.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| --include-dirs      | \[Default: false\]<br><br>If true, the source path applies to bottom-chain directories and not only to files. Bottom-chain directories are either empty or do not include other directories that match the source path.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| --exclusions        | A list of Semicolon-separated exclude patterns. Allows using wildcards.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| --sync-deletes      | \[Optional\]<br><br>Specific path in the local file system, under which to sync dependencies after the download. After the download, this path will include only the dependencies downloaded during this download operation. The other files under this path will be deleted.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| --quiet             | \[Default: false\]<br><br>If true, the delete confirmation message is skipped.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| --sort-by           | \[Optional\]<br><br>A list of semicolon-separated fields to sort by. The fields must be part of the 'items' AQL domain. For more information read the [AQL documentation](https://jfrog.com/help/r/jfrog-artifactory-documentation/artifactory-query-language)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| --sort-order        | \[Default: asc\]<br><br>The order by which fields in the 'sort-by' option should be sorted. Accepts 'asc' or 'desc'.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| --limit             | \[Optional\]<br><br>The maximum number of items to fetch. Usually used with the 'sort-by' option.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| --offset            | \[Optional\]<br><br>The offset from which to fetch items (i.e. how many items should be skipped). Usually used with the 'sort-by' option.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| --fail-no-op        | \[Default: false\]<br><br>Set to true if you'd like the command to return exit code 2 in case of no files are affected.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| --archive-entries   | \[Optional\]<br><br>If specified, only archive artifacts containing entries matching this pattern are matched. You can use wildcards to specify multiple artifacts.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| --detailed-summary  | \[Default: false\]<br><br>Set to true to include a list of the affected files as part of the command output summary.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| --insecure-tls      | \[Default: false\]<br><br>Set to true to skip TLS certificates verification.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| --gpg-key           | \[Optional\]<br><br>Path to the public GPG key file located on the file system, used to validate downloaded release bundle files.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| Command arguments   |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| Source path         | Specifies the source path in Artifactory, from which the artifacts should be downloaded. You can use wildcards to specify multiple artifacts.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| Target path         | The second argument is optional and specifies the local file system target path.<br><br>If the target path ends with a slash, the path is assumed to be a directory. For example, if you specify the target as "repo-name/a/b/", then "b" is assumed to be a directory into which files should be downloaded. If there is no terminal slash, the target path is assumed to be a file to which the downloaded file should be renamed. For example, if you specify the target as "a/b", the downloaded file is renamed to "b".<br><br>For flexibility in specifying the target path, you can include placeholders in the form of {1}, {2} which are replaced by corresponding tokens in the source path that are enclosed in parenthesis. For more details, please refer to [Using Placeholders](https://jfrog.com/help/r/jfrog-cli/using-placeholders). |

#### Examples

##### **Example 1**

Download an artifact called **cool-froggy.zip** located at the root of the **my-local-repo** repository to the current directory.
```
jf rt dl my-local-repo/cool-froggy.zip
```

##### **Example 2**

Download all artifacts located under the **all-my-frogs** directory in the **my-local-repo** repository to the **all-my-frogs** folder under the current directory.
```
jf rt dl my-local-repo/all-my-frogs/ all-my-frogs/
```

##### **Example 3**

Download all artifacts located in the **my-local-repo **repository with a **jar** extension to the **all-my-frogs** folder under the current directory.
```
jf rt dl "my-local-repo/*.jar" all-my-frogs/
```

##### **Example 4**

Download the latest file uploaded to the all-my-frogs folder in the **my-local-repo** repository.
```
jf rt dl  "my-local-repo/all-my-frogs/" --sort-by=created --sort-order=desc --limit=1
```

### Copying Files

This command is used to copy files in Artifactory

|                   |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
|-------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command name      | rt copy                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| Abbreviation      | rt cp                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| Command options   | **Warning** <br><br>When using the * or ; characters in the copy command options or arguments, make sure to wrap the whole options or arguments string in quotes (") to make sure the * or ; characters are not interpreted as literals.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| --server-id       | \[Optional\]<br><br>Server ID configured using the config command. If not specified, the default configured Artifactory server is used.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| --spec            | \[Optional\]<br><br>Path to a file spec. For more details, please refer to [Using File Specs](https://jfrog.com/help/r/jfrog-cli/using-file-specs).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| --props           | \[Optional\]<br><br>A list of Artifactory [properties](https://jfrog.com/help/r/jfrog-artifactory-documentation/Working-With-Jfrog-Properties) specified as "key=value" pairs separated by a semi-colon. (For example, "key1=value1;key2=value2;key3=value3"). Only artifacts with these properties names and values will be copied.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| --exclude-props   | \[Optional\]<br><br>A list of Artifactory [properties](https://jfrog.com/help/r/jfrog-artifactory-documentation/Working-With-Jfrog-Properties) specified as "key=value" pairs separated by a semi-colon (for example, "key1=value1;key2=value2;key3=value3"). Only artifacts **without all** of the specified properties names and values will be copied.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| --build           | \[Optional\]<br><br>If specified, only artifacts of the specified build are matched. The property format is build-name/build-number. If you do not specify the build number, the artifacts are filtered by the latest build number.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| --bundle          | \[Optional\]<br><br>If specified, only artifacts of the specified bundle are matched. The value format is bundle-name/bundle-version.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| --flat            | \[Default: false\]<br><br>If true, artifacts are copied to the exact target path specified and their hierarchy in the source path is ignored.<br><br>If false, artifacts are copied to the target path while maintaining their source path hierarchy.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| --recursive       | \[Default: true\]<br><br>If true, artifacts are also copied from sub-paths under the specified source path.<br><br>If false, only artifacts in the specified source path directory are copied.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| --dry-run         | \[Default: false\]<br><br>If true, the command only indicates which artifacts would have been copied.<br><br>If false, the command is fully executed and copies artifacts as specified.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| --exclusions      | A list of Semicolon-separated exclude patterns. Allows using wildcards.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| --threads         | \[Default: 3\]<br><br>Number of threads used for copying the items.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| --sort-by         | \[Optional\]<br><br>A list of semicolon-separated fields to sort by. The fields must be part of the 'items' AQL domain. For more information read the [AQL documentation](https://jfrog.com/help/r/jfrog-artifactory-documentation/artifactory-query-language)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| --sort-order      | \[Default: asc\]<br><br>The order by which fields in the 'sort-by' option should be sorted. Accepts 'asc' or 'desc'.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| --limit           | \[Optional\]<br><br>The maximum number of items to fetch. Usually used with the 'sort-by' option.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| --offset          | \[Optional\]<br><br>The offset from which to fetch items (i.e. how many items should be skipped). Usually used with the 'sort-by' option.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| --fail-no-op      | \[Default: false\]<br><br>Set to true if you'd like the command to return exit code 2 in case of no files are affected.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| --archive-entries | \[Optional\]<br><br>If specified, only archive artifacts containing entries matching this pattern are matched. You can use wildcards to specify multiple artifacts.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| --insecure-tls    | \[Default: false\]<br><br>Set to true to skip TLS certificates verification.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| --retries         | \[Default: 3\]<br><br>Number for HTTP retry attempts.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| --retry-wait-time | \[Default: 0s\]<br><br>Number of seconds or milliseconds to wait between retries. The numeric value should either end with s for seconds or ms for milliseconds.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| Command arguments | The command takes two arguments                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| Source path       | Specifies the source path in Artifactory, from which the artifacts should be copied, in the following format: `[repository name]/[repository path].` You can use wildcards to specify multiple artifacts.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| Target path       | Specifies the target path in Artifactory, to which the artifacts should be copied, in the following format: `[repository name]/[repository path]`<br><br>If the pattern ends with a slash, the target path is assumed to be a folder. For example, if you specify the target as "repo-name/a/b/", then "b" is assumed to be a folder in Artifactory into which files should be copied. If there is no terminal slash, the target path is assumed to be a file to which the copied file should be renamed. For example, if you specify the target as "repo-name/a/b", the copied file is renamed to "b" in Artifactory.<br><br>For flexibility in specifying the target path, you can include placeholders in the form of {1}, {2} which are replaced by corresponding tokens in the source path that are enclosed in parenthesis. For more details, please refer to [Using Placeholders](https://jfrog.com/help/r/jfrog-cli/using-placeholders). |

#### Examples

##### **Example 1**

Copy all artifacts located under **/rabbit** in the **source-frog-repo** repository into the same path in the **target-frog-repo** repository.
```
jf rt cp source-frog-repo/rabbit/ target-frog-repo/rabbit/
```

##### **Example 2**

Copy all zip files located under **/rabbit** in the **source-frog-repo** repository into the same path in the **target-frog-repo** repository.
```
jf rt cp "source-frog-repo/rabbit/*.zip" target-frog-repo/rabbit/
```

##### **Example 3**

Copy all artifacts located under **/rabbit** in the **source-frog-repo** repository and with property "Version=1.0" into the same path in the **target-frog-repo** repository.
```
jf rt cp "source-frog-repo/rabbit/*" target-frog-repo/rabbit/ --props=Version=1.0
```

### Moving Files

This command is used to move files in Artifactory

|                   |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
|-------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command name      | rt move                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| Abbreviation      | rt mv                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| Command options   | **Warning**<br><br> When using the * or ; characters in the copy command options or arguments, make sure to wrap the whole options or arguments string in quotes (") to make sure the * or ; characters are not interpreted as literals.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| --server-id       | \[Optional\]<br><br>Server ID configured using the config command. If not specified, the default configured Artifactory server is used.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| --spec            | \[Optional\]<br><br>Path to a file spec. For more details, please refer to [Using File Specs](https://jfrog.com/help/r/jfrog-cli/using-file-specs).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| --props           | \[Optional\]<br><br>A list of Artifactory [properties](https://jfrog.com/help/r/jfrog-artifactory-documentation/Working-With-Jfrog-Properties) specified as "key=value" pairs separated by a semi-colon (for example, "key1=value1;key2=value2;key3=value3"). Only artifacts with these properties names and values will be moved.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| --exclude-props   | \[Optional\]<br><br>A list of Artifactory [properties](https://jfrog.com/help/r/jfrog-artifactory-documentation/Working-With-Jfrog-Properties) specified as "key=value" pairs separated by a semi-colon (for example, "key1=value1;key2=value2;key3=value3"). Only artifacts **without all** of the specified properties names and values will be moved.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| --build           | \[Optional\]<br><br>If specified, only artifacts of the specified build are matched. The property format is build-name/build-number. If you do not specify the build number, the artifacts are filtered by the latest build number.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| --bundle          | \[Optional\]<br><br>If specified, only artifacts of the specified bundle are matched. The value format is bundle-name/bundle-version.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| --flat            | \[Default: false\]<br><br>If true, artifacts are moved to the exact target path specified and their hierarchy in the source path is ignored.<br><br>If false, artifacts are moved to the target path while maintaining their source path hierarchy.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| --recursive       | \[Default: true\]<br><br>If true, artifacts are also moved from sub-paths under the specified source path.<br><br>If false, only artifacts in the specified source path directory are moved.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| --dry-run         | \[Default: false\]<br><br>If true, the command only indicates which artifacts would have been moved.<br><br>If false, the command is fully executed and downloads artifacts as specified.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| --exclusions      | A list of Semicolon-separated exclude patterns. Allows using wildcards.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| --threads         | \[Default: 3\]<br><br>Number of threads used for moving the items.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| --sort-by         | \[Optional\]<br><br>A list of semicolon-separated fields to sort by. The fields must be part of the 'items' AQL domain. For more information read the [AQL documentation](https://jfrog.com/help/r/jfrog-artifactory-documentation/artifactory-query-language)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| --sort-order      | \[Default: asc\]<br><br>The order by which fields in the 'sort-by' option should be sorted. Accepts 'asc' or 'desc'.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| --limit           | \[Optional\]<br><br>The maximum number of items to fetch. Usually used with the 'sort-by' option.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| --offset          | \[Optional\]<br><br>The offset from which to fetch items (i.e. how many items should be skipped). Usually used with the 'sort-by' option.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| --fail-no-op      | \[Default: false\]<br><br>Set to true if you'd like the command to return exit code 2 in case of no files are affected.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| --archive-entries | \[Optional\]<br><br>If specified, only archive artifacts containing entries matching this pattern are matched. You can use wildcards to specify multiple artifacts.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| --insecure-tls    | \[Default: false\]<br><br>Set to true to skip TLS certificates verification.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| --retries         | \[Default: 3\]<br><br>Number of HTTP retry attempts.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| --retry-wait-time | \[Default: 0s\]<br><br>Number of seconds or milliseconds to wait between retries. The numeric value should either end with s for seconds or ms for milliseconds.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| Command arguments | The command takes two arguments                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| Source path       | Specifies the source path in Artifactory, from which the artifacts should be moved, in the following format: `[repository name]/[repository path].` You can use wildcards to specify multiple artifacts.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| Target path       | Specifies the target path in Artifactory, to which the artifacts should be moved, in the following format: `[repository name]/[repository path]`<br><br>If the pattern ends with a slash, the target path is assumed to be a folder. For example, if you specify the target as "repo-name/a/b/", then "b" is assumed to be a folder in Artifactory into which files should be moved. If there is no terminal slash, the target path is assumed to be a file to which the moved file should be renamed. For example, if you specify the target as "repo-name/a/b", the moved file is renamed to "b" in Artifactory.<br><br>For flexibility in specifying the upload path, you can include placeholders in the form of {1}, {2} which are replaced by corresponding tokens in the source path that are enclosed in parenthesis. For more details, please refer to [Using Placeholders](https://jfrog.com/help/r/jfrog-cli/using-placeholders). |

#### Examples

##### **Example 1**

Move all artifacts located under **/rabbit** in the **source-frog-repo** repository into the same path in the **target-frog-repo** repository.
```
jf rt mv source-frog-repo/rabbit/ target-frog-repo/rabbit/
```

##### **Example 2**

Move all zip files located under **/rabbit** in the **source-frog-repo** repository into the same path in the **target-frog-repo** repository.
```
jf rt mv "source-frog-repo/rabbit/*.zip" target-frog-repo/rabbit/
```

##### **Example 3**

Move all artifacts located under **/rabbit** in the **source-frog-repo** repository and with property "Version=1.0" into the same path in the **target-frog-repo** repository .
```
jf rt mv "source-frog-repo/rabbit/*" target-frog-repo/rabbit/ --props=Version=1.0
```

### Deleting Files

This command is used to delete files in Artifactory

|                   |                                                                                                                                                                                                                                                                                                                                                            |
|-------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command name      | rt delete                                                                                                                                                                                                                                                                                                                                                  |
| Abbreviation      | rt del                                                                                                                                                                                                                                                                                                                                                     |
| Command options   | **Warning** <br><br>When using the * or ; characters in the delete command options or arguments, make sure to wrap the whole options or arguments string in quotes (") to make sure the * or ; characters are not interpreted as literals.                                                                                                                 |
| --server-id       | \[Optional\]<br><br>Server ID configured using the config command. If not specified, the default configured Artifactory server is used.                                                                                                                                                                                                                    |
| --spec            | \[Optional\]<br><br>Path to a file spec. For more details, please refer to [Using File Specs](https://jfrog.com/help/r/jfrog-cli/using-file-specs).                                                                                                                                                                                                        |
| --props           | \[Optional\]<br><br>A list of Artifactory [properties](https://jfrog.com/help/r/jfrog-artifactory-documentation/Working-With-Jfrog-Properties) specified as "key=value" pairs separated by a semi-colon (for example, "key1=value1;key2=value2;key3=value3"). Only artifacts with these properties names and values will be deleted.                       |
| --exclude-props   | \[Optional\]<br><br>A list of Artifactory [properties](https://jfrog.com/help/r/jfrog-artifactory-documentation/Working-With-Jfrog-Properties) specified as "key=value" pairs separated by a semi-colon (for example, "key1=value1;key2=value2;key3=value3"). Only artifacts **without all** of the specified properties names and values will be deleted. |
| --build           | \[Optional\]<br><br>If specified, only artifacts of the specified build are matched. The property format is build-name/build-number. If you do not specify the build number, the artifacts are filtered by the latest build number.                                                                                                                        |
| --bundle          | \[Optional\]<br><br>If specified, only artifacts of the specified bundle are matched. The value format is bundle-name/bundle-version.                                                                                                                                                                                                                      |
| --recursive       | \[Default: true\]<br><br>If true, artifacts are also deleted from sub-paths under the specified path.                                                                                                                                                                                                                                                      |
| --quiet           | \[Default: false\]<br><br>If true, the delete confirmation message is skipped.                                                                                                                                                                                                                                                                             |
| --dry-run         | \[Default: false\]<br><br>If true, the command only indicates which artifacts would have been deleted.<br><br>If false, the command is fully executed and deletes artifacts as specified.                                                                                                                                                                  |
| --exclusions      | A list of Semicolon-separated exclude patterns. Allows using wildcards.                                                                                                                                                                                                                                                                                    |
| --sort-by         | \[Optional\]<br><br>A list of semicolon-separated fields to sort by. The fields must be part of the 'items' AQL domain. For more information read the [AQL documentation](https://jfrog.com/help/r/jfrog-artifactory-documentation/artifactory-Query-Language)                                                                                             |
| --sort-order      | \[Default: asc\]<br><br>The order by which fields in the 'sort-by' option should be sorted. Accepts 'asc' or 'desc'.                                                                                                                                                                                                                                       |
| --limit           | \[Optional\]<br><br>The maximum number of items to fetch. Usually used with the 'sort-by' option.                                                                                                                                                                                                                                                          |
| --offset          | \[Optional\]<br><br>The offset from which to fetch items (i.e. how many items should be skipped). Usually used with the 'sort-by' option.                                                                                                                                                                                                                  |
| --fail-no-op      | \[Default: false\]<br><br>Set to true if you'd like the command to return exit code 2 in case of no files are affected.                                                                                                                                                                                                                                    |
| --archive-entries | \[Optional\]<br><br>If specified, only archive artifacts containing entries matching this pattern are matched. You can use wildcards to specify multiple artifacts.                                                                                                                                                                                        |
| --threads         | \[Default: 3\]<br><br>Number of threads used for deleting the items.                                                                                                                                                                                                                                                                                       |
| --insecure-tls    | \[Default: false\]<br><br>Set to true to skip TLS certificates verification.                                                                                                                                                                                                                                                                               |
| --retries         | \[Default: 3\]<br><br>Number of HTTP retry attempts.                                                                                                                                                                                                                                                                                                       |
| --retry-wait-time | \[Default: 0s\]<br><br>Number of seconds or milliseconds to wait between retries. The numeric value should either end with s for seconds or ms for milliseconds.--retry-wait-time                                                                                                                                                                          |
| Command arguments | The command takes one argument                                                                                                                                                                                                                                                                                                                             |
| Delete path       | Specifies the path in Artifactory of the files that should be deleted in the following format: `[repository name]/[repository path].` You can use wildcards to specify multiple artifacts.                                                                                                                                                                 |

#### Examples

##### **Example 1**

Delete all artifacts located under **/rabbit** in the **frog-repo** repository.
```
jf rt del frog-repo/rabbit/
```

##### **Example 2**

Delete all zip files located under **/rabbit** in the **frog-repo** repository.
```
jf rt del "frog-repo/rabbit/*.zip"
```

### Searching Files

This command is used to search and display files in Artifactory.

|                   |                                                                                                                                                                                                                                                                                                                                                             |
|-------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command name      | rt search                                                                                                                                                                                                                                                                                                                                                   |
| Abbreviation      | rt s                                                                                                                                                                                                                                                                                                                                                        |
| Command options   | **Warning** <br><br>When using the * or ; characters in the command options or arguments, make sure to wrap the whole options or arguments string in quotes (") to make sure the * or ; characters are not interpreted as literals.                                                                                                                         |
| --server-id       | \[Optional\]<br><br>Server ID configured using the config command. If not specified, the default configured Artifactory server is used.                                                                                                                                                                                                                     |
| --spec            | \[Optional\]<br><br>Path to a file spec. For more details, please refer to [Using File Specs](https://jfrog.com/help/r/jfrog-cli/using-file-specs).                                                                                                                                                                                                         |
| --count           | \[Optional\]<br><br>Set to true to display only the total of files or folders found.                                                                                                                                                                                                                                                                        |
| --include-dirs    | \[Optional\]<br><br>Set to true if you'd like to also apply the source path pattern for directories and not only for files                                                                                                                                                                                                                                  |
| --spec-vars       | \[Optional\]<br><br>List of variables in the form of "key1=value1;key2=value2;..." to be replaced in the File Spec. In the File Spec, the variables should be used as follows: ${key1}.                                                                                                                                                                     |
| --props           | \[Optional\]<br><br>A list of Artifactory [properties](https://jfrog.com/help/r/jfrog-artifactory-documentation/Working-With-Jfrog-Properties) specified as "key=value" pairs separated by a semi-colon (for example, "key1=value1;key2=value2;key3=value3"). Only artifacts with these properties names and values will be returned.                       |
| --exclude-props   | \[Optional\]<br><br>A list of Artifactory [properties](https://jfrog.com/help/r/jfrog-artifactory-documentation/Working-With-Jfrog-Properties) specified as "key=value" pairs separated by a semi-colon (for example, "key1=value1;key2=value2;key3=value3"). Only artifacts **without all** of the specified properties names and values will be returned. |
| --build           | \[Optional\]<br><br>If specified, only artifacts of the specified build are matched. The property format is build-name/build-number. If you do not specify the build number, the artifacts are filtered by the latest build number.                                                                                                                         |
| --bundle          | \[Optional\]<br><br>If specified, only artifacts of the specified bundle are matched. The value format is bundle-name/bundle-version.                                                                                                                                                                                                                       |
| --recursive       | \[Default: true\]<br><br>Set to false if you do not wish to search artifacts inside sub-folders in Artifactory.                                                                                                                                                                                                                                             |
| --exclusions      | A list of Semicolon-separated exclude patterns. Allows using wildcards.                                                                                                                                                                                                                                                                                     |
| --sort-by         | \[Optional\]<br><br>A list of semicolon-separated fields to sort by. The fields must be part of the 'items' AQL domain. For more information read the [AQL documentation](https://jfrog.com/help/r/jfrog-artifactory-documentation/artifactory-query-language)                                                                                              |
| --sort-order      | \[Default: asc\]<br><br>The order by which fields in the 'sort-by' option should be sorted. Accepts 'asc' or 'desc'.                                                                                                                                                                                                                                        |
| --transitive      | \[Default: false\]<br><br>Set to true to look for artifacts also in remote repositories. Available on Artifactory version 7.17.0 or higher.                                                                                                                                                                                                                 |
| --limit           | \[Optional\]<br><br>The maximum number of items to fetch. Usually used with the 'sort-by' option.                                                                                                                                                                                                                                                           |
| --offset          | \[Optional\]<br><br>The offset from which to fetch items (i.e. how many items should be skipped). Usually used with the 'sort-by' option.                                                                                                                                                                                                                   |
| --fail-no-op      | \[Default: false\]<br><br>Set to true if you'd like the command to return exit code 2 in case of no files are affected.                                                                                                                                                                                                                                     |
| --archive-entries | \[Optional\]<br><br>If specified, only archive artifacts containing entries matching this pattern are matched. You can use wildcards to specify multiple artifacts.                                                                                                                                                                                         |
| --insecure-tls    | \[Default: false\]<br><br>Set to true to skip TLS certificates verification.                                                                                                                                                                                                                                                                                |
| --retries         | \[Default: 3\]<br><br>Number of HTTP retry attempts.                                                                                                                                                                                                                                                                                                        |
| --retry-wait-time | \[Default: 0s\]<br><br>Number of seconds or milliseconds to wait between retries. The numeric value should either end with s for seconds or ms for milliseconds.retry-wait-time                                                                                                                                                                             |
| --include         | \[Optional\]<br><br> List of fields in the form of \"value1;value2;...\". <br>Only the path and the fields that are specified will be returned. The fields must be part of the 'items' AQL domain. for the full supported items list  check [AQL documentation](https://jfrog.com/help/r/jfrog-artifactory-documentation/artifactory-query-language)        |
| Command arguments |                                                                                                                                                                                                                                                                                                                                                             |
| Search path       | Specifies the search path in Artifactory, in the following format: `[repository name]/[repository path].` You can use wildcards to specify multiple artifacts.                                                                                                                                                                                              |

#### Examples

##### **Example 1**

Display a list of all artifacts located under **/rabbit** in the **frog-repo** repository.
```
jf rt s frog-repo/rabbit/
```

##### **Example 2**

Display a list of all zip files located under **/rabbit** in the **frog-repo** repository.
```
jf rt s "frog-repo/rabbit/*.zip"
```

##### **Example 3**

Display a list of the files under example-repo-local with the following fields: path, actual_md5, modified_b, updated and depth.
```
jf rt s example-repo-local --include="actual_md5;modified_by;updated;depth"
```

### Setting Properties on Files

This command is used for setting properties on existing files in Artifactory.

|                   |                                                                                                                                                                                                                                                                                                                                                             |
|-------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command name      | rt set-props                                                                                                                                                                                                                                                                                                                                                |
| Abbreviation      | rt sp                                                                                                                                                                                                                                                                                                                                                       |
| Command options   | **Warning** <br><br>When using the * or ; characters in the command options or arguments, make sure to wrap the whole options or arguments string in quotes (") to make sure the * or ; characters are not interpreted as literals.                                                                                                                         |
| --server-id       | \[Optional\]<br><br>Server ID configured using the config command. If not specified, the default configured Artifactory server is used.                                                                                                                                                                                                                     |
| --spec            | \[Optional\]<br><br>Path to a file spec. For more details, please refer to [Using File Specs](https://jfrog.com/help/r/jfrog-cli/using-file-specs).                                                                                                                                                                                                         |
| --spec-vars       | \[Optional\]<br><br>List of variables in the form of "key1=value1;key2=value2;..." to be replaced in the File Spec. In the File Spec, the variables should be used as follows: ${key1}.                                                                                                                                                                     |
| --props           | \[Optional\]<br><br>List of properties in the form of "key1=value1;key2=value2,...". Only files with these properties names and values are affected.                                                                                                                                                                                                        |
| --exclude-props   | \[Optional\]<br><br>A list of Artifactory [properties](https://jfrog.com/help/r/jfrog-artifactory-documentation/Working-With-Jfrog-Properties) specified as "key=value" pairs separated by a semi-colon (for example, "key1=value1;key2=value2;key3=value3"). Only artifacts **without all** of the specified properties names and values will be affected. |
| --recursive       | \[Default: true\]<br><br>When false, artifacts inside sub-folders in Artifactory will not be affected.                                                                                                                                                                                                                                                      |
| --build           | \[Optional\]<br><br>If specified, only artifacts of the specified build are matched. The property format is build-name/build-number. If you do not specify the build number, the artifacts are filtered by the latest build number.                                                                                                                         |
| --bundle          | \[Optional\] If specified, only artifacts of the specified bundle are matched. The value format is bundle-name/bundle-version.                                                                                                                                                                                                                              |
| --include-dirs    | \[Default: false\]<br><br>When true, the properties will also be set on folders (and not just files) in Artifactory.                                                                                                                                                                                                                                        |
| --fail-no-op      | \[Default: false\]<br><br>Set to true if you'd like the command to return exit code 2 in case of no files are affected.                                                                                                                                                                                                                                     |
| --exclusions      | A list of Semicolon-separated exclude patterns. Allows using wildcards.                                                                                                                                                                                                                                                                                     |
| --sort-by         | \[Optional\]<br><br>A list of semicolon-separated fields to sort by. The fields must be part of the 'items' AQL domain. For more information read the [AQL documentation](https://jfrog.com/help/r/jfrog-artifactory-documentation/artifactory-query-language)                                                                                              |
| --sort-order      | \[Default: asc\]<br><br>The order by which fields in the 'sort-by' option should be sorted. Accepts 'asc' or 'desc'.                                                                                                                                                                                                                                        |
| --limit           | \[Optional\]<br><br>The maximum number of items to fetch. Usually used with the 'sort-by' option.                                                                                                                                                                                                                                                           |
| --offset          | \[Optional\]<br><br>The offset from which to fetch items (i.e. how many items should be skipped). Usually used with the 'sort-by' option.                                                                                                                                                                                                                   |
| --archive-entries | \[Optional\]<br><br>If specified, only archive artifacts containing entries matching this pattern are matched. You can use wildcards to specify multiple artifacts.                                                                                                                                                                                         |
| --insecure-tls    | \[Default: false\]<br><br>Set to true to skip TLS certificates verification.                                                                                                                                                                                                                                                                                |
| --threads         | \[Default: 3\]<br><br>Number of working threads.                                                                                                                                                                                                                                                                                                            |
| --retries         | \[Default: 3\]<br><br>Number of HTTP retry attempts.                                                                                                                                                                                                                                                                                                        |
| --retry-wait-time | \[Default: 0s\]<br><br>Number of seconds or milliseconds to wait between retries. The numeric value should either end with s for seconds or ms for milliseconds.                                                                                                                                                                                            |
| Command arguments | The command takes two arguments.                                                                                                                                                                                                                                                                                                                            |
| Files pattern     | Files that match the pattern will be set with the specified properties.                                                                                                                                                                                                                                                                                     |
| Files properties  | The list of properties, in the form of key1=value1;key2=value2,..., to be set on the matching artifacts.                                                                                                                                                                                                                                                    |

#### Example

##### **Example 1**

Set the properties on all the zip files in the generic-local repository. The command will set the property "a" with "1" value and the property "b" with two values: "2" and "3".
```
jf rt sp "generic-local/*.zip" "a=1;b=2,3"
```

##### **Example 2**

The command will set the property "a" with "1" value and the property "b" with two values: "2" and "3" on all files found by the File Spec my-spec.
```
jf rt sp "a=1;b=2,3" --spec my-spec
```

### Deleting Properties from Files

This command is used for deleting properties from existing files in Artifactory.

|                   |                                                                                                                                                                                                                                                                                                                                                             |
|-------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command name      | rt delete-props                                                                                                                                                                                                                                                                                                                                             |
| Abbreviation      | rt delp                                                                                                                                                                                                                                                                                                                                                     |
| Command options   | **Warning** <br><br>When using the * or ; characters in the command options or arguments, make sure to wrap the whole options or arguments string in quotes (") to make sure the * or ; characters are not interpreted as literals.                                                                                                                         |
| --server-id       | \[Optional\]<br><br>Artifactory server ID configured using the config command. If not specified, the default configured Artifactory server is used.                                                                                                                                                                                                         |
| --props           | \[Optional\]<br><br>List of properties in the form of "key1=value1;key2=value2,...". Only files with these properties are affected.                                                                                                                                                                                                                         |
| --exclude-props   | \[Optional\]<br><br>A list of Artifactory [properties](https://jfrog.com/help/r/jfrog-artifactory-documentation/Working-With-Jfrog-Properties) specified as "key=value" pairs separated by a semi-colon (for example, "key1=value1;key2=value2;key3=value3"). Only artifacts **without all** of the specified properties names and values will be affected. |
| --recursive       | \[Default: true\]<br><br>When false, artifacts inside sub-folders in Artifactory will not be affected.                                                                                                                                                                                                                                                      |
| --build           | \[Optional\]<br><br>If specified, only artifacts of the specified build are matched. The property format is build-name/build-number. If you do not specify the build number, the artifacts are filtered by the latest build number.                                                                                                                         |
| --bundle          | \[Optional\]<br><br>If specified, only artifacts of the specified bundle are matched. The value format is bundle-name/bundle-version.                                                                                                                                                                                                                       |
| --include-dirs    | \[Default: false\]<br><br>When true, the properties will also be set on folders (and not just files) in Artifactory.                                                                                                                                                                                                                                        |
| --fail-no-op      | \[Default: false\]<br><br>Set to true if you'd like the command to return exit code 2 in case of no files are affected.                                                                                                                                                                                                                                     |
| --exclusions      | A list of Semicolon-separated exclude patterns. Allows using wildcards.                                                                                                                                                                                                                                                                                     |
| --sort-by         | \[Optional\]<br><br>A list of semicolon-separated fields to sort by. The fields must be part of the 'items' AQL domain. For more information read the [AQL documentation](https://jfrog.com/help/r/jfrog-artifactory-documentation/artifactory-query-language)                                                                                              |
| --sort-order      | \[Default: asc\]<br><br>The order by which fields in the 'sort-by' option should be sorted. Accepts 'asc' or 'desc'.                                                                                                                                                                                                                                        |
| --limit           | \[Optional\]<br><br>The maximum number of items to fetch. Usually used with the 'sort-by' option.                                                                                                                                                                                                                                                           |
| --offset          | \[Optional\]<br><br>The offset from which to fetch items (i.e. how many items should be skipped). Usually used with the 'sort-by' option.                                                                                                                                                                                                                   |
| --archive-entries | \[Optional\]<br><br>If specified, only archive artifacts containing entries matching this pattern are matched. You can use wildcards to specify multiple artifacts.                                                                                                                                                                                         |
| --insecure-tls    | \[Default: false\]<br><br>Set to true to skip TLS certificates verification.                                                                                                                                                                                                                                                                                |
| --retries         | \[Default: 3\]<br><br>Number of HTTP retry attempts.                                                                                                                                                                                                                                                                                                        |
| --retry-wait-time | \[Default: 0s\]<br><br>Number of seconds or milliseconds to wait between retries. The numeric value should either end with s for seconds or ms for milliseconds.retry-wait-time                                                                                                                                                                             |
| Command arguments | The command takes two arguments.                                                                                                                                                                                                                                                                                                                            |
| Files pattern     | The properties will be deleted from files that match the pattern.                                                                                                                                                                                                                                                                                           |
| Files properties  | The list of properties, in the form of key1,key2,..., to be deleted from the matching artifacts.                                                                                                                                                                                                                                                            |

#### Example

Delete the "status" and "phase" properties from all the zip files in the generic-local repository.
```
jf rt delp "generic-local/*.zip" "status,phase"
```

### Creating Access Tokens

This command allows creating [Access Tokens](https://jfrog.com/help/r/jfrog-platform-administration-Documentation/Access-Tokens) for users in Artifactory

|                   |                                                                                                                                                                                                                                                                                                                                                                                     |
|-------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command name      | rt access-token-create                                                                                                                                                                                                                                                                                                                                                              |
| Abbreviation      | rt atc                                                                                                                                                                                                                                                                                                                                                                              |
| Command options   |                                                                                                                                                                                                                                                                                                                                                                                     |
| --groups          | \[Default: *\]<br><br>A list of comma-separated groups for the access token to be associated with. Specify * to indicate that this is a 'user-scoped token', i.e., the token provides the same access privileges that the current subject has, and is therefore evaluated dynamically. A non-admin user can only provide a scope that is a subset of the groups to which he belongs |
| --grant-admin     | \[Default: false\]<br><br>Set to true to provides admin privileges to the access token. This is only available for administrators.                                                                                                                                                                                                                                                  |
| --expiry          | \[Default: 3600\]<br><br>The time in seconds for which the token will be valid. To specify a token that never expires, set to zero. Non-admin can only set a value that is equal to or less than the default 3600.                                                                                                                                                                  |
| --refreshable     | \[Default: false\]<br><br>Set to true if you'd like the the token to be refreshable. A refresh token will also be returned in order to be used to generate a new token once it expires.                                                                                                                                                                                             |
| --audience        | \[Optional\]<br><br>A space-separate list of the other Artifactory instances or services that should accept this token identified by their Artifactory Service IDs, as obtained by the 'jf rt curl api/system/serviceid' command.                                                                                                                                                   |
| Command arguments |                                                                                                                                                                                                                                                                                                                                                                                     |
| username          | Optional - The user name for which this token is created. If not specified, the configured user is used.                                                                                                                                                                                                                                                                            |

#### **Examples**

Create an access token for the user with the **commander-will-riker** username.
```
jf rt atc commander-will-riker
```

### Cleaning Up Unreferenced Files from a Git LFS Repository

This command is used to clean up files from a Git LFS repository. This deletes all files from a Git LFS repository, which are no longer referenced in a corresponding Git repository.

|                   |                                                                                                                                               |
|-------------------|-----------------------------------------------------------------------------------------------------------------------------------------------|
| Command name      | rt git-lfs-clean                                                                                                                              |
| Abbreviation      | rt glc                                                                                                                                        |
| Command options   |                                                                                                                                               |
| --refs            | \[Default: refs/remotes/*\] List of Git references in the form of "ref1,ref2,..." which should be preserved.                                  |
| --repo            | \[Optional\] Local Git LFS repository in Artifactory which should be cleaned. If omitted, the repository is detected from the Git repository. |
| --quiet           | \[Default: false\] Set to true to skip the delete confirmation message.                                                                       |
| --dry-run         | \[Default: false\] If true, cleanup is only simulated. No files are actually deleted.                                                         |
| Command arguments | If no arguments are passed in, the command assumes the .git repository is located at current directory.                                       |
| path to .git      | Path to the directory which includes the .git directory.                                                                                      |

#### **Examples**

##### **Example 1**

Cleans up Git LFS files from Artifactory, using the configuration in the .git directory located at the current directory.
```
 jf rt glc
```

##### **Example 2**

Cleans up Git LFS files from Artifactory, using the configuration in the .git directory located inside the path/to/git/config directory.
```
jf rt glc path/to/git/config
```

### Running cUrl

Execute a cUrl command, using the configured Artifactory details. The command expects the cUrl client to be included in the PATH.

> **Note** - This command supports only Artifactory REST APIs, which are accessible under https://&lt;JFrog base URL&gt;/artifactory/api/


|                          |                                                                                                                                                                                                                                                                                             |     |
|--------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----|
| Command name             | rt curl                                                                                                                                                                                                                                                                                     |     |
| Abbreviation             | rt cl                                                                                                                                                                                                                                                                                       |     |
| Command options          |                                                                                                                                                                                                                                                                                             |     |
| --server-id              | \[Optional\]<br><br>Server ID configured using the **jf c add** command. If not specified, the default configured server is used.                                                                                                                                                           |     |
| Command arguments        |                                                                                                                                                                                                                                                                                             |     |
| cUrl arguments and flags | The same list of arguments and flags passed to cUrl, except for the following changes:<br><br>1.  The full Artifactory URL should not be passed. Instead, the REST endpoint URI should be sent.<br>2.  The login credentials should not be passed. Instead, the --server-id should be used. |     |

Currently only servers configured with username and password / API key are supported.
  
#### **Examples**

##### **Example 1**

Execute the cUrl client, to send a GET request to the /api/build endpoint to the default Artifactory server
```
jf rt curl -XGET /api/build
```

##### **Example 2**

Execute the cUrl client, to send a GET request to the /api/build endpoint to the configured my-rt-server server ID.
```
jf rt curl -XGET /api/build --server-id my-rt-server
```

## Build Integration
### Overview

JFrog CLI integrates with any development ecosystem allowing you to collect build-info and then publish it to Artifactory. By publishing build-info to Artifactory, JFrog CLI empowers Artifactory to provide visibility into artifacts deployed, dependencies used and extensive information on the build environment to allow fully traceable builds. Read more about build-info and build integration with Artifactory [here](https://jfrog.com/help/r/jfrog-integrations-documentation/Build-Integration).

Many of JFrog CLI's commands accept two optional command options: **--build-name** and **--build-number**. When these options are added, JFrog CLI collects and records the build info locally for these commands.  
When running multiple commands using the same build and build number, JFrog CLI aggregates the collected build info into one build.  
The recorded build-info can be later published to Artifactory using the [build-publish](https://jfrog.com/help/r/jfrog-cli/publishing-build-info) command.

### Collecting Build-Info

Build-info is collected by adding the `--build-name` and `--build-number` options to different CLI commands. The CLI commands can be run several times and cumulatively collect build-info for the specified build name and number until it is published to Artifactory. For example, running the `jf rt download` command several times with the same build name and number will accumulate each downloaded file in the corresponding build-info.

#### Collecting Dependencies

Dependencies are collected by adding the `--build-name` and `--build-number` options to the `jf rt download` command.

For example, the following command downloads the `cool-froggy.zip` file found in repository `my-local-repo`, but it also specifies this file as a dependency in build `my-build-name` with build number 18:
```
jf rt dl my-local-repo/cool-froggy.zip --build-name=my-build-name --build-number=18
```
#### Collecting Build Artifacts

Build artifacts are collected by adding the `--build-name` and `--build-number` options to the `jf rt upload` command.

For example, the following command specifies that file `froggy.tgz` uploaded to repository `my-local-repo` is a build artifact of build `my-build-name` with build number 18:
```
jf rt u froggy.tgz my-local-repo --build-name=my-build-name --build-number=18
```
#### Collecting Environment Variables

This command is used to collect environment variables and attach them to a build.

Environment variables are collected using the `build-collect-env` (`bce`) command.

For example, the following command collects all currently known environment variables, and attaches them to the build-info for build `my-build-name` with build number 18:
```
jf rt bce my-build-name 18
```
The following table lists the command arguments and flags:

|                   |                                        |
|-------------------|----------------------------------------|
| Command name      | rt build-collect-env                   |
| Abbreviation      | rt bce                                 |
| Command options   |                                        |
| --project         | \[Optional\]<br><br>JFrog project key. |
| Command arguments | The command accepts two arguments.     |
| Build name        | Build name.                            |
| Build number      | Build number.                          |

##### Example
Collect environment variables for build name: frogger-build and build number: 17
```
jf rt bce frogger-build 17
```

#### Collecting Information from Git
The `build-add-git` (bag) command collects the Git revision and URL from the local .git directory and adds it to the build-info. It can also collect the list of tracked project issues (for example, issues stored in JIRA or other bug tracking systems) and add them to the build-info. The issues are collected by reading the git commit messages from the local git log. Each commit message is matched against a pre-configured regular expression, which retrieves the issue ID and issue summary. The information required for collecting the issues is retrieved from a yaml configuration file provided to the command.

The following table lists the command arguments and flags:

|                   |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
|-------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command name      | rt build-add-git                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| Abbreviation      | rt bag                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| Command options   |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| --config          | \[Optional\]<br><br>Path to a yaml configuration file, used for collecting tracked project issues and adding them to the build-info.                                                                                                                                                                                                                                                                                                                                                                                         |
| --server-id       | \[Optional\]<br><br>Server ID configured using the ['jf config' command](https://jfrog.com/help/r/jfrog-cli/jfrog-Platform-CONFIGURATION). This is the server to which the build-info will be later published, using the `jf rt build-publish` command. This option, if provided, overrides the serverID value in this command's yaml configuration. If both values are not provided, the default server, configured by the ['jf config' command](https://jfrog.com/help/r/jfrog-cli/jfrog-Platform-CONFIGURATION), is used. |
| --project         | \[Optional\]<br><br>JFrog project key.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| Command arguments | The command accepts three arguments.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| Build name        | Build name.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| Build number      | Build number.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| .git path         | Optional - Path to a directory containing the .git directory. If not specific, the .git directory is assumed to be in the current directory or in one of the parent directories.                                                                                                                                                                                                                                                                                                                                             |

##### Example
```
jf rt bag frogger-build 17 checkout-dir
```

This is the configuration file structure.
```yaml
version: 1
issues: 
  # The serverID yaml property is optional. The --server-id command option, if provided, overrides the serverID value.
  # If both the serverID property and the --server-id command options are not provided,
  # the default server, configured by the "jfrog config add" command is used.
  serverID: my-artifactory-server

  trackerName: JIRA
  regexp: (.+-\[0-9\]+)\\s-\\s(.+)
  keyGroupIndex: 1
  summaryGroupIndex: 2
  trackerUrl: https://my-jira.com/issues
  aggregate: true
  aggregationStatus: RELEASED
```

##### Configuration file properties

|                   |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
|-------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Property name     | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| Version           | The schema version is intended for internal use. Do not change!                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| serverID          | Artifactory server ID configured by the ['jf config' command](https://jfrog.com/help/r/jfrog-cli/jfrog-Platform-CONFIGURATION). The command uses this server for fetching details about previous published builds. The **--server-id** command option, if provided, overrides the **serverID** value.  <br>If both the **serverID** property and the **--server-id** command options are not provided, the default server, configured by the ['jf config' command](https://jfrog.com/help/r/jfrog-cli/jfrog-Platform-CONFIGURATION) is used. |
| trackerName       | The name (type) of the issue tracking system. For example, JIRA. This property can take any value.                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| regexp            | A regular expression used for matching the git commit messages. The expression should include two capturing groups - for the issue key (ID) and the issue summary. In the example above, the regular expression matches the commit messages as displayed in the following example:<br><br>HAP-1007 - This is a sample issue                                                                                                                                                                                                                  |
| keyGroupIndex     | The capturing group index in the regular expression used for retrieving the issue key. In the example above, setting the index to "1" retrieves **HAP-1007** from this commit message:<br><br>HAP-1007 - This is a sample issue                                                                                                                                                                                                                                                                                                              |
| summaryGroupIndex | The capturing group index in the regular expression for retrieving the issue summary. In the example above, setting the index to "2" retrieves the sample issue from this commit message:<br><br>HAP-1007 - This is a sample issue                                                                                                                                                                                                                                                                                                           |
| trackerUrl        | The issue tracking URL. This value is used for constructing a direct link to the issues in the Artifactory build UI.                                                                                                                                                                                                                                                                                                                                                                                                                         |
| aggregate         | Set to true, if you wish all builds to include issues from previous builds.                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| aggregationStatus | If aggregate is set to true, this property indicates how far in time should the issues be aggregated. In the above example, issues will be aggregated from previous builds, until a build with a RELEASE status is found. Build statuses are set when a build is promoted using the **jf rt build-promote** command.                                                                                                                                                                                                                         |

#### Adding Files as Build Dependencies

The download command, as well as other commands which download dependencies from Artifactory accept the **--build-name**  and **--build-number**  command options. Adding these options records the downloaded files as build dependencies. In some cases however,  it is necessary to add a file, which has been downloaded by another tool, to a build. Use the **build-add-dependencies** command to this.

By default, the command collects the files from the local file system. If you'd like the files to be collected from Artifactory however, add the **--from-rt** option to the command.

|                   |                                                                                                                                                                                                                                                                                                                                                                     |
|-------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command name      | rt build-add-dependencies                                                                                                                                                                                                                                                                                                                                           |
| Abbreviation      | rt bad                                                                                                                                                                                                                                                                                                                                                              |
| Command options   | **Warning**<br><br> When using the * or ; characters in the command options or arguments, make sure to wrap the whole options or arguments string in quotes (") to make sure the * or ; characters are not interpreted as literals.                                                                                                                                 |
| --from-rt         | \[Default: false\]<br><br>Set to true to search the files in Artifactory, rather than on the local file system. The --regexp option is not supported when --from-rt is set to true.                                                                                                                                                                                 |
| --server-id       | \[Optional\]<br><br>Server ID configured using the config command.                                                                                                                                                                                                                                                                                                  |
| --spec            | \[Optional\]<br><br>Path to a File Spec.                                                                                                                                                                                                                                                                                                                            |
| --spec-vars       | \[Optional\]<br><br>List of variables in the form of "key1=value1;key2=value2;..." to be replaced in the File Spec. In the File Spec, the variables should be used as follows: ${key1}.                                                                                                                                                                             |
| --recursive       | \[Default: true\]<br><br>When false, artifacts inside sub-folders in Artifactory will not be affected.                                                                                                                                                                                                                                                              |
| --regexp          | \[Optional: false\]<br><br>\[Default: false\] Set to true to use a regular expression instead of wildcards expression to collect files to be added to the build info.This option is not supported when --from-rt is set to true.                                                                                                                                    |
| --dry-run         | \[Default: false\]<br><br>Set to true to only get a summery of the dependencies that will be added to the build info.                                                                                                                                                                                                                                               |
| --module          | \[Optional\]<br><br>Optional module name in the build-info for adding the dependency.                                                                                                                                                                                                                                                                               |
| --exclusions      | A list of  Semicolon-separated  exclude patterns. Allows using wildcards or a regular expression  according to the value of the 'regexp' option.                                                                                                                                                                                                                    |
| Command arguments | The command takes three arguments.                                                                                                                                                                                                                                                                                                                                  |
| Build name        | The build name to add the dependencies to                                                                                                                                                                                                                                                                                                                           |
| Build number      | The build number to add the dependencies to                                                                                                                                                                                                                                                                                                                         |
| Pattern           | Specifies the local file system path to dependencies which should be added to the build info. You can specify multiple dependencies by using wildcards or a regular expression as designated by the --regexp command option. If you have specified that you are using regular expressions, then the first one used in the argument must be enclosed in parenthesis. |

##### Example

**Example 1**

Add all files located under the **path/to/build/dependencies/dir** directory as dependencies of a build. The build name is **my-build-name** and the build number is **7**. The build-info is only updated locally. To publish the build-info to Artifactory use the **jf rt build-publish** command.
```
jf rt bad my-build-name 7 "path/to/build/dependencies/dir/"
```
  

**Example 2**

Add all files located in the **m-local-repo** Artifactory repository, under the **dependencies** folder, as dependencies of a build. The build name is **my-build-name** and the build number is **7**.  The build-info is only updated locally. To publish the build-info to Artifactory use the **jf rt build-publish** command.
```
jf rt bad my-build-name 7 "my-local-repo/dependencies/" --from-rt
```


**Example 3**

Add all files located under the **path/to/build/dependencies/dir** directory as dependencies of a build. The build name is **my-build-name**, the build number is **7** and module is m1. The build-info is only updated locally. To publish the build-info to Artifactory use the **jf rt build-publish** command.
```
jf rt bad my-build-name 7 "path/to/build/dependencies/dir/" --module m1
```

### Publishing Build-Info

This command is used to publish build info to Artifactory. To publish the accumulated build-info for a build to Artifactory, use the **build-publish** command. For example, the following command publishes all the build-info collected for build **my-build-name** with build number 18:
```
jf bp my-build-name 18
```

This command is used to publish build info to Artifactory.

|                   |                                                                                                                                                                                          |
|-------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command name      | rt build-publish                                                                                                                                                                         |
| Abbreviation      | rt bp                                                                                                                                                                                    |
| Command options   |                                                                                                                                                                                          |
| --server-id       | \[Optional\]<br><br>Server ID configured using the config command. If not specified, the default configured Artifactory server is used.                                                  |
| --project         | \[Optional\]<br><br>JFrog project key.                                                                                                                                                   |
| --build-url       | \[Optional\]<br><br>Can be used for setting the CI server build URL in the build-info.                                                                                                   |
| --env-include     | \[Default: *\]<br><br>List of patterns in the form of "value1;value2;..." Only environment variables that match those patterns will be included in the build info.                       |
| --env-exclude     | \[Default: \*password\*;\*secret\*;\*key\*\]<br><br>List of case insensitive  patterns in the form of "value1;value2;..."   environment variables match those patterns will be excluded. |
| --dry-run         | \[Default: false\]<br><br>Set to true to disable communication with Artifactory.                                                                                                         |
| --insecure-tls    | \[Default: false\]<br><br>Set to true to skip TLS certificates verification.                                                                                                             |
| Command arguments | The command accepts two arguments.                                                                                                                                                       |
| Build name        | Build name to be published.                                                                                                                                                              |
| Build number      | Build number to be published.                                                                                                                                                            |

##### Example
```
jf rt bp my-build-name 18
```

### Aggregating Published Builds
The build-info, which is collected and published to Artifactory  by the **jf rt build-publish** command, can include multiple modules. Each module in the build-info represents a package, which is the result of a single build step, or in other words, a JFrog CLI command execution. For example, the following command adds a module named **m1** to a build named **my-build** with **1** as the build number:
```
jf rt upload "a/*.zip" generic-local --build-name my-build --build-number 1 --module m1
```

The following command, adds a second module, named **m2** to the same build:
```
jf rt upload "b/*.zip" generic-local --build-name my-build --build-number 1 --module m2
```

You now publish the generated build-info to Artifactory using the following command:
```
jf rt build-publish my-build 1
```

Now that you have your build-info published to Artifactory, you can perform actions on the entire build. For example, you can download, copy, move or delete all or some of the artifacts of a build. Here's how you do this.  
```
jf rt download "*" --build my-build/1
```

In some cases though, your build is composed of multiple build steps, which are running on multiple different machines or spread across different time periods. How do you aggregate those build steps, or in other words, aggregate those command executions, into one build-info?

The way to do this, is to create a separate build-info for every section of the build, and publish it independently to Artifactory. Once all the build-info instances are published, you can create a new build-info, which references all the previously published build-info instances. The new build-info can be viewed as a "master" build-info, which references other build-info instances.

So the next question is - how can this reference between the two build-instances be created?

The way to do this is by using the **build-append** command. Running this command on an unpublished build-info, adds a reference to a different build-info, which has already been published to Artifactory. This reference is represented by a new module in the new build-info. The ID of this module will have the following format: **&lt;referenced build name&gt;/&lt;referenced build number&gt;.

Now, when downloading the artifacts of the "master" build, you'll actually be downloading the artifacts of all of its referenced builds. The examples below demonstrates this,

|                        |                                                           |
|------------------------|-----------------------------------------------------------|
| Command name           | rt build-append                                           |
| Abbreviation           | rt ba                                                     |
| Command options        | This command has no options.                              |
| Command arguments      | The command accepts four arguments.                       |
| Build name             | The current (not yet published) build name.               |
| Build number           | The current (not yet published) build number,             |
| build name to append   | The published build name to append to the current build   |
| build number to append | The published build number to append to the current build |

##### Requirements

Artifactory version 7.25.4 and above.

##### Example
```yaml
# Create and publish build a/1
jf rt upload "a/*.zip" generic-local --build-name a --build-number 1
jf rt build-publish a 1

# Create and publish build b/1
jf rt upload "b/*.zip" generic-local --build-name b --build-number 1
jf rt build-publish b 1

# Append builds a/1 and b/1 to build aggregating-build/10
jf rt build-append aggregating-build 10 a 1
jf rt build-append aggregating-build 10 b 1

# Publish build aggregating-build/10
jf rt build-publish aggregating-build 10

# Download the artifacts of aggregating-build/10, which is the same as downloading the of a/1 and b/1
jf rt download --build aggregating-build/10
```

### Promoting a Build

This command is used to [promote build](https://jfrog.com/knowledge-base/how-does-build-promotion-work/) in Artifactory.

|                        |                                                                                                                                            |
|------------------------|--------------------------------------------------------------------------------------------------------------------------------------------|
| Command name           | rt build-promote                                                                                                                           |
| Abbreviation           | rt bpr                                                                                                                                     |
| Command options        |                                                                                                                                            |
| --server-id            | \[Optional\]<br><br>Server ID configured using the config command. If not specified, the default configured Artifactory server is used.    |
| --project              | \[Optional\]<br><br>JFrog project key.                                                                                                     |
| --status               | \[Optional\]<br><br>Build promotion status.                                                                                                |
| --comment              | \[Optional\]<br><br>Build promotion comment.                                                                                               |
| --source-repo          | \[Optional\]<br><br>Build promotion source repository.                                                                                     |
| --include-dependencies | \[Default: false\]<br><br>If set to true, the build dependencies are also promoted.                                                        |
| --copy                 | \[Default: false\]<br><br>If set true, the build artifacts and dependencies are copied to the target repository, otherwise they are moved. |
| --props                | \[Optional\]<br><br>List of properties in the form of "key1=value1;key2=value2,...". to attach to the build artifacts.                     |
| --dry-run              | \[Default: false\]<br><br>If true, promotion is only simulated. The build is not promoted.                                                 |
| --insecure-tls         | \[Default: false\]<br><br>Set to true to skip TLS certificates verification.                                                               |
| Command arguments      | The command accepts three arguments.                                                                                                       |
| Build name             | Build name to be promoted.                                                                                                                 |
| Build number           | Build number to be promoted.                                                                                                               |
| Target repository      | Build promotion target repository.                                                                                                         |

##### Example
```
jf rt bpr my-build-name 18 target-repository
```

### Cleaning up the Build

Build-info is accumulated by the CLI according to the commands you apply until you publish the build-info to Artifactory. If, for any reason, you wish to "reset" the build-info and cleanup (i.e. delete) any information accumulated so far, you can use the `build-clean` (`bc`) command.

The following table lists the command arguments and flags:

|                   |                                    |
|-------------------|------------------------------------|
| Command name      | rt build-clean                     |
| Abbreviation      | rt bc                              |
| Command options   | The command has no options.        |
| Command arguments | The command accepts two arguments. |
| Build name        | Build name.                        |
| Build number      | Build number.                      |

  

For example, the following command cleans up any build-info collected for build `my-build-name` with build number 18:
```
jf rt bc my-build-name 18
```

### Discarding Old Builds from Artifactory

This command is used to discard builds previously published to Artifactory using the [build-publish](https://jfrog.com/help/r/jfrog-cli/publishing-build-info) command.  
  
The following table lists the command arguments and flags:

|                    |                                                                                                                                         |
|--------------------|-----------------------------------------------------------------------------------------------------------------------------------------|
| Command name       | rt build-discard                                                                                                                        |
| Abbreviation       | rt bdi                                                                                                                                  |
| Command options    |                                                                                                                                         |
| --server-id        | \[Optional\]<br><br>Server ID configured using the config command. If not specified, the default configured Artifactory server is used. |
| --max-days         | \[Optional\]<br><br>The maximum number of days to keep builds in Artifactory.                                                           |
| --max-builds       | \[Optional\]<br><br>The maximum number of builds to store in Artifactory.                                                               |
| --exclude-builds   | \[Optional\]<br><br>List of build numbers in the form of "value1,value2,...", that should not be removed from Artifactory.              |
| --delete-artifacts | \[Default: false\]<br><br>If set to true, automatically removes build artifacts stored in Artifactory.                                  |
| --async            | \[Default: false\]<br><br>If set to true, build discard will run asynchronously and will not wait for response.                         |
| Command arguments  | The command accepts one argument.                                                                                                       |
| Build name         | Build name.                                                                                                                             |

##### Example

**Example 1**

Discard the oldest build numbers of build **my-build-name** from Artifactory, leaving only the 10 most recent builds.
```
jf rt bdi my-build-name --max-builds= 10
```

**Example 2**

Discard the oldest build numbers of build **my-build-name** from Artifactory, leaving only builds published during the last 7 days.
```
jf rt bdi my-build-name --max-days=7
```

**Example 3**

Discard the oldest build numbers of build **my-build-name** from Artifactory, leaving only builds published during the last 7 days. **b20** and **b21** will not be discarded.
```
jf rt bdi my-build-name --max-days 7 --exclude-builds "b20,b21"
```

Package Managers Integration
----------------------------

### Running Maven Builds

JFrog CLI includes integration with Maven, allowing you to resolve dependencies and deploy build artifacts from and to Artifactory, while collecting build-info and storing it in Artifactory.

#### Setting maven repositories

Before using the **mvn** command, the project needs to be pre-configured with the Artifactory server and repositories, to be used for building and publishing the project. The **mvn-config** command should be used once to add the configuration to the project. The command should run while inside the root directory of the project. The configuration is stored by the command in the **.jfrog** directory at the root directory of the project.

|                          |                                                                                                                                                                                                                                                                      |
|--------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name             | mvn-config                                                                                                                                                                                                                                                           |
| Abbreviation             | mvnc                                                                                                                                                                                                                                                                 |
| Command options          |                                                                                                                                                                                                                                                                      |
| --global                 | \[Optional\]<br><br>Set to true, if you'd like the configuration to be global (for all projects on the machine). Specific projects can override the global configuration.                                                                                            |
| --server-id-resolve      | \[Optional\]<br><br>Server ID for resolution. The server should configured using the 'jf rt c' command.                                                                                                                                                              |
| --server-id-deploy       | \[Optional\]<br><br>Server ID for deployment. The server should be configured using the 'jf rt c' command.                                                                                                                                                           |
| --repo-resolve-releases  | \[Optional\]<br><br>Resolution repository for release dependencies.                                                                                                                                                                                                  |
| --repo-resolve-snapshots | \[Optional\]<br><br>Resolution repository for snapshot dependencies.                                                                                                                                                                                                 |
| --repo-deploy-releases   | \[Optional\]<br><br>Deployment repository for release artifacts.                                                                                                                                                                                                     |
| --repo-deploy-snapshots  | \[Optional\]<br><br>Deployment repository for snapshot artifacts.                                                                                                                                                                                                    |
| --include-patterns       | \[Optional\]<br><br>Filter deployed artifacts by setting a wildcard pattern that specifies which artifacts to include. You may provide multiple patterns separated by a comma followed by a white-space. For example<br><br>artifact-*.jar, artifact-*.pom           |
| --exclude-patterns       | \[Optional\]<br><br>Filter deployed artifacts by setting a wildcard pattern that specifies which artifacts to exclude. You may provide multiple patterns separated by a comma followed by a white-space. For example<br><br>artifact-*-test.jar, artifact-*-test.pom |
| --scan                   | \[Default: false\]<br><br>Set if you'd like all files to be scanned by Xray on the local file system prior to the upload, and skip the upload if any of the files are found vulnerable.                                                                              |
| --format                 | \[Default: table\]<br><br>Should be used with the --scan option. Defines the scan output format. Accepts table or json as values.                                                                                                                                    |
| Command arguments        | The command accepts no arguments                                                                                                                                                                                                                                     |

#### Running maven

The **mvn** command triggers the maven client, while resolving dependencies and deploying artifacts from and to Artifactory.

> **Note**: Before running the **mvn** command on a project for the first time, the project should be configured with the **mvn-config** command.

> **Note**: If the machine running JFrog CLI has no access to the internet, make sure to read the [Downloading the Maven and Gradle Extractor JARs](https://jfrog.com/help/r/jfrog-cli/downloading-the-maven-and-gradle-extractor-jars) section.

The following table lists the command arguments and flags:

|                   |                                                                                                                                                |
|-------------------|------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name      | mvn                                                                                                                                            |
| Abbreviation      | mvn                                                                                                                                            |
| Command options   |                                                                                                                                                |
| --threads         | \[Default: 3\]<br><br>Number of threads for uploading build artifacts.                                                                         |
| --build-name      | \[Optional\]<br><br>Build name. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration).   |
| --build-number    | \[Optional\]<br><br>Build number. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration). |
| --project         | \[Optional\]<br><br>JFrog project key.                                                                                                         |
| --insecure-tls    | \[Default: false\]<br><br>Set to true to skip TLS certificates verification.                                                                   |
| Command arguments | The command accepts the same arguments and options as the mvn client.                                                                          |

**Deploying Maven Artifacts**

The deployment to Artifacts is triggered both by the deployment and install phases.
To disable artifacts deployment, add** **-Dartifactory.publish.artifacts=false** to the list of goals and options.
For example: "**clean install****-Dartifactory.publish.artifacts=false"**

##### Examples

**Example 1**

Run clean and install with maven.
```
jf mvn clean install -f path/to/pom-file
```

### Running Gradle Builds

JFrog CLI includes integration with Gradle, allowing you to resolve dependencies and deploy build artifacts from and to Artifactory, while collecting build-info and storing it in Artifactory.

#### Setting gradle repositories

Before using the **gradle** command, the project needs to be pre-configured with the Artifactory server and repositories, to be used for building and publishing the project. The **gradle-config** command should be used once to add the configuration to the project. The command should run while inside the root directory of the project. The configuration is stored by the command in the**.jfrog** directory at the root directory of the project.

|                         |                                                                                                                                                                                         |
|-------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name            | gradle-config                                                                                                                                                                           |
| Abbreviation            | gradlec                                                                                                                                                                                 |
| Command options         |                                                                                                                                                                                         |
| --global                | \[Optional\]<br><br>Set to true, if you'd like the configuration to be global (for all projects on the machine). Specific projects can override the global configuration.               |
| --server-id-resolve     | \[Optional\]<br><br>Server ID for resolution. The server should configured using the 'jf c add' command.                                                                                |
| --server-id-deploy      | \[Optional\]<br><br>Server ID for deployment. The server should be configured using the 'jf c add' command.                                                                             |
| --repo-resolve          | \[Optional\]<br><br>Repository for dependencies resolution.                                                                                                                             |
| --repo-deploy           | \[Optional\]<br><br>Repository for artifacts deployment.                                                                                                                                |
| --uses-plugin           | \[Default: false\]<br><br>Set to true if the Gradle Artifactory Plugin is already applied in the build script.                                                                          |
| --use-wrapper           | \[Default: false\]<br><br>Set to true if you'd like to use the Gradle wrapper.                                                                                                          |
| --deploy-maven-desc     | \[Default: true\]<br><br>Set to false if you do not wish to deploy Maven descriptors.                                                                                                   |
| --deploy-ivy-desc       | \[Default: true\]<br><br>Set to false if you do not wish to deploy Ivy descriptors.                                                                                                     |
| --ivy-desc-pattern      | \[Default: '\[organization\]/\[module\]/ivy-\[revision\].xml'<br><br>Set the deployed Ivy descriptor pattern.                                                                           |
| --ivy-artifacts-pattern | \[Default: '\[organization\]/\[module\]/\[revision\]/\[artifact\]-\[revision\](-\[classifier\]).\[ext\]'<br><br>Set the deployed Ivy artifacts pattern.                                 |
| --scan                  | \[Default: false\]<br><br>Set if you'd like all files to be scanned by Xray on the local file system prior to the upload, and skip the upload if any of the files are found vulnerable. |
| --format                | \[Default: table\]<br><br>Should be used with the --scan option. Defines the scan output format. Accepts table or json as values.                                                       |
| Command arguments       | The command accepts no arguments                                                                                                                                                        |

#### Running gradle

The **gradle** command triggers the gradle client, while resolving dependencies and deploying artifacts from and to Artifactory.

> **Note**: Before running the **gradle** command on a project for the first time, the project should be configured with the **gradle-config** command.

> **Note**: If the machine running JFrog CLI has no access to the internet, make sure to read the[Downloading the Maven and Gradle Extractor JARs](https://jfrog.com/help/r/jfrog-cli/downloading-the-maven-and-gradle-extractor-jars)section.

The following table lists the command arguments and flags:

|                   |                                                                                                                                                |
|-------------------|------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name      | gradle                                                                                                                                         |
| Abbreviation      | gradle                                                                                                                                         |
| Command options   |                                                                                                                                                |
| --threads         | \[Default: 3\]<br><br>Number of threads for uploading build artifacts.                                                                         |
  | --build-name      | \[Optional\]<br><br>Build name. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration).   |
| --build-number    | \[Optional\]<br><br>Build number. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration). |
| --project         | \[Optional\]<br><br>JFrog project key.                                                                                                         |
| Command arguments | The command accepts the same arguments and options as the gradle client.                                                                       |

##### Examples

**Example 1**

Build the project using the **artifactoryPublish** task, while resolving and deploying artifacts from and to Artifactory.
```
jf gradle clean artifactoryPublish -b path/to/build.gradle
```

### Running Builds with MSBuild

JFrog CLI includes integration with MSBuild and Artifactory, allowing you to resolve dependencies and deploy build artifacts from and to Artifactory, while collecting build-info and storing it in Artifactory. This is done by having JFrog CLI in your search path and adding JFrog CLI commands to the MSBuild `csproj` file.

For detailed instructions, please refer to our [MSBuild Project Example](https://github.com/eyalbe4/project-examples/tree/master/msbuild-example) on GitHub.

### Managing Docker Images

JFrog CLI provides full support for pulling and publishing docker images from and to Artifactory using the docker client running on the same machine. This allows you to collect build-info for your docker build and then publish it to Artifactory. You can also promote the pushed docker images from one repository to another in Artifactory.

To build and push your docker images to Artifactory, follow these steps:

1.  Make sure Artifactory can be used as docker registry. Please refer to [Getting Started with Docker and Artifactory](https://jfrog.com/help/r/jfrog-artifactory-documentation/Getting-Started-With-Artifactory-As-A-docker-Registry) in the JFrog Artifactory User Guide.  
2.  Make sure that the installed docker client has version **17.07.0-ce (2017-08-29)** or above. To verify this, run **docker -v**** 
3.  To ensure that the docker client and your Artifactory docker registry are correctly configured to work together, run the following code snippet.
    
```
docker pull hello-world
docker tag hello-world:latest &lt;artifactoryDockerRegistry&gt;/hello-world:latest
docker login &lt;artifactoryDockerRegistry&gt;
docker push &lt;artifactoryDockerRegistry&gt;/hello-world:latest
```
    
If everything is configured correctly, pushing any image including the hello-world image should be successfully uploaded to Artifactory.
    
> **Note**: When running the docker-pull and docker-push commands, the CLI will first attempt to log in to the docker registry. In case of a login failure, the command will not be executed.

#### Examples

Check out our [docker project examples on GitHub](https://github.com/jfrog/project-examples/tree/master/docker-oci-examples).

#### Pulling Docker Images Using the Docker Client

Running **docker-pull** command allows pulling docker images from Artifactory, while collecting the build-info and storing it locally, so that it can be later published to Artifactory, using the [build-publish](https://jfrog.com/help/r/jfrog-cli/publishing-build-info) command.

The following table lists the command arguments and flags:

  

|                   |                                                                                                                                                |
|-------------------|------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name      | docker pull                                                                                                                                    |
| Abbreviation      | dpl                                                                                                                                            |
| Command options   |                                                                                                                                                |
| --server-id       | \[Optional\]<br><br>Server ID configured using the config command. If not specified, the default configured Artifactory server is used.        |
| --build-name      | \[Optional\]<br><br>Build name. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration).   |
| --build-number    | \[Optional\]<br><br>Build number. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration). |
| --project         | \[Optional\]<br><br>JFrog project key.                                                                                                         |
| --module          | \[Optional\]<br><br>Optional module name for the build-info.                                                                                   |
| --skip-login      | \[Default: false\]<br><br>Set to true if you'd like the command to skip performing docker login.                                               |
| Command arguments | The same arguments and options supported by the docker client/                                                                                 |

##### Examples

jf docker pull my-docker-registry.io/my-docker-image:latest --build-name=my-build-name --build-number=7

You can then publish the build-info collected by the **docker-pull** command to Artifactory using the [build-publish](https://jfrog.com/help/r/jfrog-cli/publishing-build-info) command.

  

#### Pushing Docker Images Using the Docker Client

After building your image using the docker client, the **docker-push** command pushes the image layers to Artifactory, while collecting the build-info and storing it locally, so that it can be later published to Artifactory, using the **build-publish** command.

The following table lists the command arguments and flags:

  

|                    |                                                                                                                                                |
|--------------------|------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name       | docker push                                                                                                                                    |
| Abbreviation       | dp                                                                                                                                             |
| Command options    |                                                                                                                                                |
| --server-id        | \[Optional\]<br><br>Server ID configured using the config command. If not specified, the default configured Artifactory server is used.        |
| --build-name       | \[Optional\]<br><br>Build name. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration).   |
| --build-number     | \[Optional\]<br><br>Build number. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration). |
| --project          | \[Optional\]<br><br>JFrog project key.                                                                                                         |
| --module           | \[Optional\]<br><br>Optional module name for the build-info.                                                                                   |
| --skip-login       | \[Default: false\]<br><br>Set to true if you'd like the command to skip performing docker login.                                               |
| --threads          | \[Default: 3\]<br><br>Number of working threads.                                                                                               |
| --detailed-summary | \[Default: false\]<br><br>Set true to include a list of the affected files as part of the command output summary.                              |
| Command arguments  | The same arguments and options supported by the docker client/                                                                                 |

##### Examples

jf docker push my-docker-registry.io/my-docker-image:latest --build-name=my-build-name --build-number=7

You can then publish the build-info collected by the **docker-push** command to Artifactory using the [build-publish](https://jfrog.com/help/r/jfrog-cli/publishing-build-info) command.

  

#### Pulling Docker Images Using Podman

[Podman](https://podman.io/) is a daemonless container engine for developing, managing, and running OCI Containers. Running the **podman-pull** command allows pulling docker images from Artifactory using podman, while collecting the build-info and storing it locally, so that it can be later published to Artifactory, using the [build-publish](https://jfrog.com/help/r/jfrog-cli/publishing-build-info) command.

The following table lists the command arguments and flags:

|                   |                                                                                                                                                |
|-------------------|------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name      | rt podman-pull                                                                                                                                 |
| Abbreviation      | rt ppl                                                                                                                                         |
| Command options   |                                                                                                                                                |
| --server-id       | \[Optional\]<br><br>Server ID configured using the config command. If not specified, the default configured Artifactory server is used.        |
| --build-name      | \[Optional\]<br><br>Build name. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration).   |
| --build-number    | \[Optional\]<br><br>Build number. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration). |
| --project         | \[Optional\]<br><br>JFrog project key.                                                                                                         |
| --module          | \[Optional\]<br><br>Optional module name for the build-info.                                                                                   |
| --skip-login      | \[Default: false\]<br><br>Set to true if you'd like the command to skip performing docker login.                                               |
| Command argument  |                                                                                                                                                |
| Image tag         | The docker image tag to pull.                                                                                                                  |
| Source repository | Source repository in Artifactory.                                                                                                              |

##### Examples
```
jf rt podman-pull my-docker-registry.io/my-docker-image:latest docker-local --build-name=my-build-name --build-number=7
```
You can then publish the build-info collected by the **podman-pull** command to Artifactory using the [build-publish](https://jfrog.com/help/r/jfrog-cli/publishing-build-info) command.

  

#### Pushing Docker Images Using Podman

[Podman](https://podman.io/) is a daemon-less container engine for developing, managing, and running OCI Containers. After building your image, the **podman-push** command pushes the image layers to Artifactory, while collecting the build-info and storing it locally, so that it can be later published to Artifactory, using the **build-publish** command.

The following table lists the command arguments and flags:

  

|                    |                                                                                                                                                |
|--------------------|------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name       | rt podman-push                                                                                                                                 |
| Abbreviation       | rt pp                                                                                                                                          |
| Command options    |                                                                                                                                                |
| --server-id        | \[Optional\]<br><br>Server ID configured using the config command. If not specified, the default configured Artifactory server is used.        |
| --build-name       | \[Optional\]<br><br>Build name. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration).   |
| --build-number     | \[Optional\]<br><br>Build number. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration). |
| --project          | \[Optional\]<br><br>JFrog project key.                                                                                                         |
| --module           | \[Optional\]<br><br>Optional module name for the build-info.                                                                                   |
| --skip-login       | \[Default: false\]<br><br>Set to true if you'd like the command to skip performing docker login.                                               |
| --threads          | \[Default: 3\]<br><br>Number of working threads.                                                                                               |
| --detailed-summary | \[Default: false\]<br><br>Set to true to include a list of the affected files as part of the command output summary.                           |
| Command argument   |                                                                                                                                                |
| Image tag          | The docker image tag to push.                                                                                                                  |
| Target repository  | Target repository in Artifactory.                                                                                                              |

##### Examples
```
jf rt podman-push my-docker-registry.io/my-docker-image:latest docker-local --build-name=my-build-name --build-number=7
```
You can then publish the build-info collected by the **podman-push** command to Artifactory using the [build-publish](https://jfrog.com/help/r/jfrog-cli/publishing-build-info) command.

  

#### Pushing Docker Images Using Kaniko

JFrog CLI allows pushing containers to Artifactory using [Kaniko](https://github.com/GoogleContainerTools/kaniko#kaniko---build-images-in-kubernetes), while collecting build-info and storing it in Artifactory.  
For detailed instructions, please refer to our [Kaniko project example on GitHub](https://github.com/jfrog/project-examples/tree/master/docker-oci-examples/kaniko-example).

#### Pushing Docker Images Using buildx

JFrog CLI allows pushing containers to Artifactory using [buildx](https://github.com/GoogleContainerTools/kaniko#kaniko---build-images-in-kubernetes), while collecting build-info and storing it in Artifactory.  
For detailed instructions, please refer to our [buildx project example on GitHub](https://github.com/jfrog/project-examples/tree/master/docker-oci-examples/fat-manifest-example).

#### Pushing Docker Images Using the OpenShift CLI

JFrog CLI allows pushing containers to Artifactory using the [OpenShift CLI](https://docs.openshift.com/container-platform/4.2/cli_reference/openshift_cli/getting-started-cli.html), while collecting build-info and storing it in Artifactory.  
For detailed instructions, please refer to our [OpenShift build project example on GitHub](https://github.com/jfrog/project-examples/tree/master/docker-oci-examples/openshift-examples/openshift-build-example).

  

#### Adding Published Docker Images to the Build-Info

The **build-docker-create** command allows adding a docker image, which is already published to Artifactory, into the build-info. This build-info can be later published to Artifactory, using the **build-publish** command.

  

|                   |                                                                                                                                                                                                                            |
|-------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name      | rt build-docker-create                                                                                                                                                                                                     |
| Abbreviation      | rt bdc                                                                                                                                                                                                                     |
| Command options   |                                                                                                                                                                                                                            |
| --image-file      | Path to a file which includes one line in the following format: IMAGE-TAG@sha256:MANIFEST-SHA256. For example:<br><br>cat image-file-details<br>superfrog-docker.jfrog.io/hello-frog@sha256:30f04e684493fb5ccc030969df6de0 |
| --server-id       | \[Optional\]<br><br>Server ID configured using the config command. If not specified, the default configured Artifactory server is used.                                                                                    |
| --build-name      | \[Optional\]<br><br>Build name. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration).                                                                               |
| --build-number    | \[Optional\]<br><br>Build number. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration).                                                                             |
| --project         | \[Optional\]<br><br>JFrog project key.                                                                                                                                                                                     |
| --module          | \[Optional\]<br><br>Optional module name for the build-info.                                                                                                                                                               |
| --skip-login      | \[Default: false\]<br><br>Set to true if you'd like the command to skip performing docker login.                                                                                                                           |
| --threads         | \[Default: 3\]<br><br>Number of working threads.                                                                                                                                                                           |
| Command argument  |                                                                                                                                                                                                                            |
| Target repository | The name of the repository to which the image was pushed.                                                                                                                                                                  |

##### Examples
```
jf rt bdc docker-local --image-file image-file-details --build-name myBuild --build-number 1
```
You can then publish the build-info collected by the **podman-push** command to Artifactory using the [build-publish](https://jfrog.com/help/r/jfrog-cli/publishing-build-info) command.

  

#### Promoting Docker Images

Promotion is the action of moving or copying a group of artifacts from one repository to another, to support the artifacts' lifecycle. When it comes to docker images, there are two ways to promote a docker image which was pushed to Artifactory:

1.  Create build-info for the docker image, and then promote the build using the **jf rt build-promote** command.
2.  Use the **jf rt docker-promote** command as described below.

The following table lists the command arguments and flags:

|                       |                                                                                                                                         |
|-----------------------|-----------------------------------------------------------------------------------------------------------------------------------------|
| Command-name          | rt docker-promote                                                                                                                       |
| Abbreviation          | rt dpr                                                                                                                                  |
| Command options       |                                                                                                                                         |
| --server-id           | \[Optional\]<br><br>Server ID configured using the config command. If not specified, the default configured Artifactory server is used. |
| --copy                | \[Default: false\]<br><br>If set true, the Docker image is copied to the target repository, otherwise it is moved.                      |
| --source-tag          | \[Optional\]<br><br>The tag name to promote.                                                                                            |
| --target-docker-image | \[Optional\]<br><br>Docker target image name.                                                                                           |
| --target-tag          | \[Optional\]<br><br>The target tag to assign the image after promotion.                                                                 |
| Command argument      |                                                                                                                                         |
| source docker image   | The docker image name to promote.                                                                                                       |
| source repository     | Source repository in Artifactory.                                                                                                       |
| target repository     | Target repository in Artifactory.                                                                                                       |

##### Examples

Promote the **hello-world** docker image from the **docker-dev-local** repository to the **docker-staging-local** repository.
```
jf rt docker-promote hello-world docker-dev-local docker-staging-local
```
  

### Building Npm Packages Using the Npm Client

JFrog CLI provides full support for building npm packages using the npm client. This allows you to resolve npm dependencies, and publish your npm packages from and to Artifactory, while collecting build-info and storing it in Artifactory.

Follow these guidelines when building npm packages:

* You can download npm packages from any npm repository type - local, remote or virtual, but you can only publish to a local or virtual Artifactory repository, containing local repositories. To publish to a virtual repository, you first need to set a default local repository. For more details, please refer to [Deploying to a Virtual Repository](https://jfrog.com/help/r/jfrog-artifactory-documentation/virtual-repositories).
* When the **npm-publish** command runs, JFrog CLI runs the **pack** command in the background. The pack action is followed by an upload, which is not based on the npm client's publish command. Therefore, If your npm package includes the **prepublish** or **postpublish** scripts, rename them to **prepack** and **postpack** respectively.
    

##### Requirements

Npm client version 5.4.0 and above.

Artifactory version 5.5.2 and above.

#### Setting npm repositories

Before using the **npm-install**, **npm-ci** and **npm-publish** commands, the project needs to be pre-configured with the Artifactory server and repositories, to be used for building and publishing the project. The **npm-config** command should be used once to add the configuration to the project. The command should run while inside the root directory of the project. The configuration is stored by the command in the **.jfrog** directory at the root directory of the project.

|                     |                                                                                                                                                                           |
|---------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name        | npm-config                                                                                                                                                                |
| Abbreviation        | npmc                                                                                                                                                                      |
| Command options     |                                                                                                                                                                           |
| --global            | \[Optional\]<br><br>Set to true, if you'd like the configuration to be global (for all projects on the machine). Specific projects can override the global configuration. |
| --server-id-resolve | \[Optional\]<br><br>Artifactory server ID for resolution. The server should configured using the 'jfrog c add' command.                                                   |
| --server-id-deploy  | \[Optional\]<br><br>Artifactory server ID for deployment. The server should be configured using the 'jfrog c add' command.                                                |
| --repo-resolve      | \[Optional\]<br><br>Repository for dependencies resolution.                                                                                                               |
| --repo-deploy       | \[Optional\]<br><br>Repository for artifacts deployment.                                                                                                                  |
| Command arguments   | The command accepts no arguments                                                                                                                                          |

#### Installing Npm Packages

The **npm-install** and **npm-ci** commands execute npm's **install** and **ci** commands respectively, to fetches the npm dependencies from the npm repositories.

Before running the **npm-install** or **npm-ci** command on a project for the first time, the project should be configured using the **npm-config** command.

The following table lists the command arguments and flags:

|                   |                                                                                                                                                |
|-------------------|------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name      | npm                                                                                                                                            |
| Abbreviation      |                                                                                                                                                |
| Command options   |                                                                                                                                                |
| --build-name      | \[Optional\]<br><br>Build name. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration).   |
| --build-number    | \[Optional\]<br><br>Build number. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration). |
| --project         | \[Optional\]<br><br>JFrog project key.                                                                                                         |
| --module          | \[Optional\]<br><br>Optional module name for the build-info.                                                                                   |
| --threads         | \[Default: 3\]<br><br>Number of working threads for build-info collection.                                                                     |
| Command arguments | The command accepts the same arguments and options as the npm client.                                                                          |

##### Examples

##### Example 1

The following example installs the dependencies and records them locally as part of build **my-build-name/1**. The build-info can later be published to Artifactory using the [build-publish](https://jfrog.com/help/r/jfrog-cli/publishing-build-info) command. The dependencies are resolved from the Artifactory server and repository configured by **npm-config** command.
```
jf npm install --build-name=my-build-name --build-number=1
```

##### Example 2

The following example installs the dependencies. The dependencies are resolved from the Artifactory server and repository configured by **npm-config** command.
```
jf npm install
```

##### Example 3

The following example installs the dependencies using the npm-ci command. The dependencies are resolved from the Artifactory server and repository configured by **npm-config** command.
```
jf npm ci
```
  

#### Publishing the Npm Packages into Artifactory

The **npm-publish** command packs and deploys the npm package to the designated npm repository.

Before running the **npm-publish** command on a project for the first time, the project should be configured using the **npm-config** command. This configuration includes the Artifactory server and repository to which the package should deploy.

> **Warning**: If your npm package includes the prepublish or postpublish scripts, please refer to the guidelines above.

The following table lists the command arguments and flags:

|                    |                                                                                                                                                                                         |
|--------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name       | npm publish                                                                                                                                                                             |
| Abbreviation       |                                                                                                                                                                                         |
| Command options    |                                                                                                                                                                                         |
| --build-name       | \[Optional\]<br><br>Build name. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration).                                            |
| --build-number     | \[Optional\]<br><br>Build number. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration).                                          |
| --project          | \[Optional\]<br><br>JFrog project key.                                                                                                                                                  |
| --module           | \[Optional\]<br><br>Optional module name for the build-info.                                                                                                                            |
| --detailed-summary | \[Default: false\]<br><br>Set true to include a list of the affected files as part of the command output summary.                                                                       |
| --scan             | \[Default: false\]<br><br>Set if you'd like all files to be scanned by Xray on the local file system prior to the upload, and skip the upload if any of the files are found vulnerable. |
| --format           | \[Default: table\]<br><br>Should be used with the --scan option. Defines the scan output format. Accepts table or JSON as values.                                                       |
| Command argument   | The command accepts the same arguments and options that the **npm pack** command expects.                                                                                               |

##### Example

To pack and publish the npm package and also record it locally as part of build **my-build-name/1**, run the following command. The build-info can later be published to Artifactory using the [build-publish](https://jfrog.com/help/r/jfrog-cli/publishing-build-info) command. The package is published to the Artifactory server and repository configured by **npm-config** command.
```
jf npm publish --build-name=my-build-name --build-number=1
```

### Building Npm Packages Using the Yarn Client

JFrog CLI provides full support for building npm packages using the yarn client. This allows you to resolve npm dependencies, while collecting build-info and storing it in Artifactory. You can download npm packages from any npm repository type - local, remote or virtual. Publishing the packages to a local npm repository is supported through the **jf rt upload** command.

Yarn version 2.4.0 and above is supported.

#### Setting npm repositories

Before using the **jf yarn** command, the project needs to be pre-configured with the Artifactory server and repositories, to be used for building the project. The **yarn-config** command should be used once to add the configuration to the project. The command should run while inside the root directory of the project. The configuration is stored by the command in the **.jfrog** directory at the root directory of the project.

|                     |                                                                                                                                                                           |
|---------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name        | yarn-config                                                                                                                                                               |
| Abbreviation        | yarnc                                                                                                                                                                     |
| Command options     |                                                                                                                                                                           |
| --global            | \[Optional\]<br><br>Set to true, if you'd like the configuration to be global (for all projects on the machine). Specific projects can override the global configuration. |
| --server-id-resolve | \[Optional\]<br><br>Artifactory server ID for resolution. The server should configured using the 'jf c add' command.                                                      |
| --repo-resolve      | \[Optional\]<br><br>Repository for dependencies resolution.                                                                                                               |
| Command arguments   | The command accepts no arguments                                                                                                                                          |

#### Installing Npm Packages

The **jf yarn** command executes the yarn client, to fetch the npm dependencies from the npm repositories.

> **Note**: Before running the command on a project for the first time, the project should be configured using the **yarn-config** command.

The following table lists the command arguments and flags:

|                   |                                                                                                                                                |
|-------------------|------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name      | yarn                                                                                                                                           |
| Command options   |                                                                                                                                                |
| --build-name      | \[Optional\]<br><br>Build name. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration).   |
| --build-number    | \[Optional\]<br><br>Build number. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration). |
| --project         | \[Optional\]<br><br>JFrog project key.                                                                                                         |
| --module          | \[Optional\]<br><br>Optional module name for the build-info.                                                                                   |
| --threads         | \[Default: 3\]<br><br>Number of working threads for build-info collection.                                                                     |
| Command arguments | The command accepts the same arguments and options as the yarn client.                                                                         |

##### Examples

##### Example 1

The following example installs the dependencies and records them locally as part of build **my-build-name/1**. The build-info can later be published to Artifactory using the [build-publish](https://jfrog.com/help/r/jfrog-cli/publishing-build-info) command. The dependencies are resolved from the Artifactory server and repository configured by **yarn-config command.
```
jf yarn install --build-name=my-build-name --build-number=1
```
  

Example 2

The following example installs the dependencies. The dependencies are resolved from the Artifactory server and repository configured by **yarn-config command.
```
jf yarn install
```
  

### Building Go Packages

#### General

JFrog CLI provides full support for building Go packages using the Go client. This allows resolving Go dependencies from and publish your Go packages to Artifactory, while collecting build-info and storing it in Artifactory.

#### Requirements

JFrog CLI client version 1.20.0 and above.

Artifactory version 6.1.0 and above.

Go client version 1.11.0 and above.

#### Example project

To help you get started, you can use [this sample project on GitHub](https://github.com/jfrog/project-examples/tree/master/golang-example).

#### Setting Go repositories

Before you can use JFrog CLI to build your Go projects with Artifactory, you first need to set the resolutions and deployment repositories for the project.

Here's how you set the repositories.

1.  'cd' into to the root of the Go project.
2.  Run the **jf rt go-config** command.

|                     |                                                                                                                                                                                |
|---------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name        | go-config                                                                                                                                                                      |
| Abbreviation        |                                                                                                                                                                                |
| Command options     |                                                                                                                                                                                |
| --global            | \[Default false\]<br><br>Set to true, if you'd like the configuration to be global (for all projects on the machine). Specific projects can override the global configuration. |
| --server-id-resolve | \[Optional\]<br><br>Artifactory server ID for resolution. The server should configured using the 'jf c add' command.                                                           |
| --server-id-deploy  | \[Optional\]<br><br>Artifactory server ID for deployment. The server should be configured using the 'jf c add' command.                                                        |
| --repo-resolve      | \[Optional\]<br><br>Repository for dependencies resolution.                                                                                                                    |
| --repo-deploy       | \[Optional\]<br><br>Repository for artifacts deployment.                                                                                                                       |

##### Examples

##### Example 1

Set repositories for this go project.
```
jf go-config
```

##### Example 2

Set repositories for all go projects on this machine.
```
jf go-config --global
```
  

#### Running Go commands

The **go** command triggers the go client.

> **Note**: Before running the **go** command on a project for the first time, the project should be configured using the **go-config**  command.

The following table lists the command arguments and flags:

  

|                   |                                                                                                                                                |
|-------------------|------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name      | go                                                                                                                                             |
| Abbreviation      | go                                                                                                                                             |
| Command options   |                                                                                                                                                |
| --build-name      | \[Optional\]<br><br>Build name. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration).   |
| --build-number    | \[Optional\]<br><br>Build number. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration). |
| --project         | \[Optional\]<br><br>JFrog project key.                                                                                                         |
| --no-fallback     | \[Default: false\]<br><br>Set to avoid downloading packages from the VCS, if they are missing in Artifactory.                                  |
| --module          | \[Optional\]<br><br>Optional module name for the build-info.                                                                                   |
| Command arguments |                                                                                                                                                |
| Go command        | The command accepts the same arguments and options as the go client.                                                                           |

##### Examples

##### Example 1

The following example runs Go build command. The dependencies resolved from Artifactory via the go-virtual repository.

> **Note**: Before using this example, please make sure to set repositories for the Go project using the go-config command.

```
jf rt go build
```

##### Example 2

The following example runs Go build command, while recording the build-info locally under build name **my-build** and build number **1**. The build-info can later be published to Artifactory using the [build-publish](https://jfrog.com/help/r/jfrog-cli/publishing-build-info)command.

> **Note**: Before using this example, please make sure to set repositories for the Go project using the go-config command.
```
jf rt go build --build-name=my-build --build-number=1
```
  

#### Publishing Go Packages to Artifactory

The **go-publish** command packs and deploys the Go package to the designated Go repository in Artifactory.

> **Note**: Before running the **go-publish** command on a project for the first time, the project should be configured using the **go-config** command.

The following table lists the command arguments and flags:

  

|                    |                                                                                                                                                |
|--------------------|------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name       | go-publish                                                                                                                                     |
| Abbreviation       | gp                                                                                                                                             |
| Command options    |                                                                                                                                                |
| --build-name       | \[Optional\]<br><br>Build name. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration).   |
| --build-number     | \[Optional\]<br><br>Build number. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration). |
| --project          | \[Optional\]<br><br>JFrog project key.                                                                                                         |
| --module           | \[Optional\]<br><br>Optional module name for the build-info.                                                                                   |
| --detailed-summary | \[Default: false\]<br><br>Set true to include a list of the affected files as part of the command output summary.                              |
| Command argument   |                                                                                                                                                |
| Version            | The version of the Go project that is being published                                                                                          |

##### Examples

##### Example 1

To pack and publish the Go package, run the following command. Before running this command on a project for the first time, the project should be configured using the **go-config** command.
```
jf gp v1.2.3 
```

##### Example 2

To pack and publish the Go package and also record the build-info as part of build **my-build-name/1** , run the following command. The build-info can later be published to Artifactory using the [build-publish](https://jfrog.com/help/r/jfrog-cli/publishing-build-info) command. Before running this command on a project for the first time, the project should be configured using the **go-config** command.
```
jf gp v1.2.3 --build-name my-build-name --build-number 1
```

### Building Python Packages

JFrog CLI provides full support for building Python packages using the **pip** and **pipenv** and poetry package installers. This allows resolving python dependencies from Artifactory, while recording the downloaded packages. The downloaded packages are stored as dependencies in the build-info stored in Artifactory.

Once the packages are installed, the Python project can be then built and packaged using the pip, pipenv or poetry clients. Once built, the produced artifacts can be uploaded to Artifactory using JFrog CLI's upload command and registered as artifacts in the build-info.

#### Example projects

To help you get started, you can use [the sample projects on GitHub](https://github.com/jfrog/project-examples/tree/master/python-example).

#### Setting Python repository

Before you can use JFrog CLI to build your Python projects with Artifactory, you first need to set the repository for the project.

Here's how you set the repositories.

1.  'cd' into the root of the Python project.
2.  Run the**jf pip-config**,  **jf pipenv-config** or **jf poetry-configc** commands, depending on whether you're using the pip or pipenv clients.

|                     |                                                                                                                                                                                |
|---------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name        | pip-config / pipenv-config / poetry-config                                                                                                                                     |
| Abbreviation        | pipc / pipec / poc                                                                                                                                                             |
| Command options     |                                                                                                                                                                                |
| --global            | \[Default false\]<br><br>Set to true, if you'd like the configuration to be global (for all projects on the machine). Specific projects can override the global configuration. |
| --server-id-resolve | \[Optional\]<br><br>Artifactory server ID for resolution. The server should configured using the 'jf c add' command.                                                           |
| --repo-resolve      | \[Optional\]<br><br>Repository for dependencies resolution.                                                                                                                    |

##### Examples

##### Example 1

Set repositories for this Python project when using the pip client.
```
jf pipc
```

##### Example 2

Set repositories for all Python projects using the pip client on this machine.
```
jf pipc --global
```

##### Example 3

Set repositories for this Python project when using the pipenv client.
```
jf pipec
```

##### Example 4

Set repositories for all Python projects using the poetry client on this machine.
```
jf poc --global
```

##### Example 5

Set repositories for this Python project when using the poetry client.
```
jf poc
```

##### Example 6

Set repositories for all Python projects using the pipenv client on this machine.
```
jf pipec --global
```

#### Installing Python packages

The **pip install**,  **pipenv install** and **poetry install** commands use the **pip**, **pipenv** and **poetry** clients respectively, to install the project dependencies from Artifactory. The commands can also record these packages as build dependencies as part of the build-info published to Artifactory.

> **Note**: Before running the **pip install**, **pipenv install** and **poetry install** commands on a project for the first time, the project should be configured using the **pip-config** ,**pipenv-config** or **poetry-config** commands respectively.

**Recording all dependencies**
JFrog CLI records the installed packages as build-info dependencies. The recorded dependencies are packages installed during the **jf rt pip-install** command execution. When running the command inside a Python environment, which already has some of the packages installed, the installed packages will not be included as part of the build-info, because they were not originally installed by JFrog CLI. A warning message will be added to the log in this case.
**How to include all packages in the build-info?**

The details of all the installed packages are always cached by the **jf pip install** and **jf pipenv install** command in the **.jfrog/projects/deps.cache.json** file, located under the root of the project. JFrog CLI uses this cache for including previously installed packages in the build-info.  
If the Python environment had some packages installed prior to the first execution of the install command, those previously installed packages will be missing from the cache and therefore will not be included in the build-info.

Running the install command with both the **no-cache-dir** and **force-reinstall** pip options, should re-download and install these packages, and they will therefore be included in the build-info and added to the cache. It is also recommended to run the command from inside a [virtual environment](https://packaging.python.org/guides/installing-using-pip-and-virtual-environments/).


|                  |                                                                                                                                                |
|------------------|------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name     | pip / pipenv / poetry                                                                                                                          |
| Abbreviation     |                                                                                                                                                |
| Command options  |                                                                                                                                                |
| --build-name     | \[Optional\]<br><br>Build name. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration).   |
| --build-number   | \[Optional\]<br><br>Build number. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration). |
| --project        | \[Optional\]<br><br>JFrog project key.                                                                                                         |
| --module         | \[Optional\]<br><br>Optional module name for the build-info.                                                                                   |
| Command argument |                                                                                                                                                |
| Pip arguments    | Arguments and options for the pip-install command.                                                                                             |

  

##### Examples

Example 1

The following command triggers pip install, while recording the build dependencies as part of build name **my-build** and build number **1** .
```
jf pip install . --build-name my-build --build-number 1
```

Example 2

The following command triggers pipenv install, while recording the build dependencies as part of build name **my-build** and build number **1** .
```
jf pipenv install . --build-name my-build --build-number 1
```

Example 3

The following are command triggers poetry install, while recording the build dependencies as part of build name **my-build** and build number **1** .
```
jf poetry install . --build-name my-build --build-number 1
```
### Building NuGet Packages

JFrog CLI provides full support for restoring NuGet packages using the NuGet client or the .NET Core CLI. This allows you to resolve NuGet dependencies  from and publish your NuGet packages to  Artifactory,  while collecting build-info and storing it in Artifactory.  

NuGet dependencies resolution is supported by the `jf nuget` command, which uses the NuGet client or the `jf dotnet` command, which uses the .NET Core CLI.  

To publish your NuGet packages to Artifactory, use the [jf rt upload](https://jfrog.com/help/r/jfrog-cli/uploading-files) command.

#### Setting NuGet repositories

Before using the**nuget** or **dotnet** commands, the project needs to be pre-configured with the Artifactory server and repository, to be used for building the project.

Before using the nuget or dotnet commands, the **nuget-config** or **dotnet-config** commands should be used respectively. These commands configure the project with the details of the Artifactory server and repository, to be used for the build.  The **nuget-config** or **dotnet-config** commands should be executed while inside the root directory of the project. The configuration is stored by the command in the **.jfrog** directory at the root directory of the project. You then have the option of storing the .jfrog directory with the project sources, or  creating this configuration after the sources are checked out.

The following table lists the commands' options:

|                     |                                                                                                                                                                           |
|---------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name        | nuget-config / dotnet-config                                                                                                                                              |
| Abbreviation        | nugetc / dotnetc                                                                                                                                                          |
| Command options     |                                                                                                                                                                           |
| --global            | \[Optional\]<br><br>Set to true, if you'd like the configuration to be global (for all projects on the machine). Specific projects can override the global configuration. |
| --server-id-resolve | \[Optional\]<br><br>Artifactory server ID for resolution. The server should configured using the 'jf c add' command.                                                      |
| --repo-resolve      | \[Optional\]<br><br>Repository for dependencies resolution.                                                                                                               |
| --nuget-v2          | \[Default: false\]  <br>Set to true if you'd like to use the NuGet V2 protocol when restoring packages from Artifactory (instead of NuGet V3).                            |
| Command arguments   | The command accepts  no arguments                                                                                                                                         |

#### Running Nuget and Dotnet commands

The **nuget** command runs the **NuGet client** and the **dotnet** command runs the **.NET Core CLI.

> Before running the nuget command on a project for the first time, the project should be configured using the nuget-config command.

> Before running the dotnet command on a project for the first time, the project should be configured using the dotnet-config command.

The following table lists the commands arguments and options:

|                  |                                                                                                                                                |
|------------------|------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name     | nuget / dotnet                                                                                                                                 |
| Abbreviation     |                                                                                                                                                |
| Command options  |                                                                                                                                                |
| --build-name     | \[Optional\]<br><br>Build name. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration).   |
| --build-number   | \[Optional\]<br><br>Build number. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration). |
| --project        | \[Optional\]<br><br>JFrog project key.                                                                                                         |
| --module         | \[Optional\]<br><br>Optional module name for the build-info.                                                                                   |
| Command argument | The command accepts the same arguments and options as the NuGet client / .NET Core CLI.                                                        |

##### Examples

##### Example 1

Run nuget restore for the solution at the current directory, while resolving the NuGet dependencies from the pre-configured Artifactory repository. Use the NuGet client for this command
```
jf nuget restore
```

##### Example 2

Run dotnet restore for the solution at the current directory, while resolving the NuGet dependencies from the pre-configured Artifactory repository. Use the .NET Core CLI for this command
```
jf dotnet restore
```

##### Example 3

Run dotnet restore for the solution at the current directory, while resolving the NuGet dependencies from the pre-configured Artifactory repository.

In addition, record the build-info as part of build **my-build-name/1**. The build-info can later be published to Artifactory using the [build-publish](https://jfrog.com/help/r/jfrog-cli/publishing-build-info) command:
```
jf dotnet restore --build-name=my-build-name --build-number=1
```

### Packaging and Publishing Terraform Modules

JFrog CLI supports packaging Terraform modules and publishing them to a Terraform repository in Artifactory using the **jf terraform publish** command.

We recommend using [this example project on GitHub](https://github.com/jfrog/project-examples/tree/master/terraform-example) for an easy start up.

Before using the **jf terraform publish** command for the first time, you first need to configure the Terraform repository for your Terraform project. To do this, follow these steps:

1.  'cd' into the root directory for your Terraform project.
2.  Run the interactive **jf terraform-config** command and set deployment repository name.

The **jf terraform-config** command will store the repository name inside the **.jfrog** directory located in the current directory. You can also add the **--global** command option, if you prefer the repository configuration applies to all projects on the machine. In that case, the configuration will be saved in JFrog CLI's home directory.

The following table lists the command options:

|                    |                                                                                                                                                                           |
|--------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name       | terraform-config                                                                                                                                                          |
| Abbreviation       | tfc                                                                                                                                                                       |
| Command options    |                                                                                                                                                                           |
| --global           | \[Optional\]<br><br>Set to true, if you'd like the configuration to be global (for all projects on the machine). Specific projects can override the global configuration. |
| --server-id-deploy | \[Optional\]<br><br>Artifactory server ID for deployment. The server should configured using the 'jf c add' command.                                                      |
| --repo-deploy      | \[Optional\]<br><br>Repository for artifacts deployment.                                                                                                                  |
| Command arguments  | The command accepts no arguments                                                                                                                                          |

##### Examples

##### Example 1

Configuring the Terraform repository for a project, while inside the root directory of the project
```
jf tfc
```

##### Example 2

Configuring the Terraform repository for all projects on the machine
```
jf tfc --global
```
  

The **terraform publish** command creates a terraform package for the module in the current directory, and publishes it to the configured Terraform repository in Artifactory.

The following table lists the commands arguments and options:

|                  |                                                                                                                                                                            |
|------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name     | terraform publish                                                                                                                                                          |
| Abbreviation     | tf p                                                                                                                                                                       |
| Command options  |                                                                                                                                                                            |
| --namespace      | \[Mandatory\]<br><br>Terraform module namespace                                                                                                                            |
| --provider       | \[Mandatory\]<br><br>Terraform module provider                                                                                                                             |
| --tag            | \[Mandatory\]<br><br>Terraform module tag                                                                                                                                  |
| --exclusions     | \[Optional\]<br><br>A list of Semicolon-separated exclude patterns wildcards. Paths inside the module matching one of the patterns are excluded from the deployed package. |
| --build-name     | \[Optional\]<br><br>Build name. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration).                               |
| --build-number   | \[Optional\]<br><br>Build number. For more details, please refer to [Build Integration](https://jfrog.com/help/r/jfrog-cli/build-integration).                             |
| --project        |                                                                                                                                                                            |
| Command argument | The command accepts no arguments                                                                                                                                           |

##### Examples

##### Example 1

The command creates a package for the Terraform module in the current directory, and publishes it to the Terraform repository (configured by the **jf tfc command**) with the provides namespace, provider and tag.
```
jf tf p --namespace example --provider aws --tag v0.0.1
```

##### Example 2

The command creates a package for the Terraform module in the current directory, and publishes it to the Terraform repository (configured by the **jf tfc command**) with the provides namespace, provider and tag. The published package will not include the module paths which include either **test** or **ignore** .
```
jf tf p --namespace example --provider aws --tag v0.0.1 --exclusions "\*test\*;\*ignore\*"
```

##### Example 3

The command creates a package for the Terraform module in the current directory, and publishes it to the Terraform repository (configured by the **jf tfc** command) with the provides namespace, provider and tag. The published module will be recorded as an artifact of a build named **my-build** with build number **1**. The **jf rt bp** command publishes the build to Artifactory.
```
jf tf p --namespace example --provider aws --tag v0.0.1 --build-name my-build --build-number 1
jf rt bp my-build 1
```

## Managing Configuration Entities

JFrog CLI offers a set of commands for managing Artifactory configuration entities.

### Creating Users

This command allows creating a bulk of users. The details of the users are provided in a CSV format file. Here's the file format.
```
"username","password","email"
"username1","password1","john@c.com"
"username2","password1","alice@c.com"
```

> **Note**: The first line in the CSV is cells' headers. It is mandatory and is used by the command to map the cell value to the users' details.

The CSV can include additional columns, with different headers, which will be ignored by the command.

|                   |                                                                                                                                            |
|-------------------|--------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name      | rt users-create                                                                                                                            |
| Abbreviation      | rt uc                                                                                                                                      |
| Command options   |                                                                                                                                            |
| --server-id       | \[Optional\]<br><br>Artifactory server ID configured using the config command.                                                             |
| --csv             | \[Mandatory\]<br><br>Path to a CSV file with the users' details. The first row of the file should include the name,password,email headers. |
| --replace         | \[Optional\]<br><br>Set to true if you'd like existing users or groups to be replaced.                                                     |
| --users-groups    | \[Optional\]<br><br>A list of comma-separated groups for the new users to be associated to.                                                |
| Command arguments | The command accepts no arguments                                                                                                           |

##### Example

Create new users according to details defined in the path/to/users.csv file.
```
jf rt users-create --csv path/to/users.csv
```
  

### Deleting Users

This command allows deleting a bulk of users. The command a list of usernames to delete. The list can be either provided as a comma-seperated argument, or as a CSV file, which includes one column with the usernames. Here's the CSV format.
```
"username"
"username1"
"username2"
"username2"
```

The first line in the CSV is cells' headers. It is mandatory and is used by the command to map the cell value to the users' details.

The CSV can include additional columns, with different headers, which will be ignored by the command.

|                   |                                                                                                                                                                               |
|-------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name      | rt users-delete                                                                                                                                                               |
| Abbreviation      | rt udel                                                                                                                                                                       |
| Command options   |                                                                                                                                                                               |
| --server-id       | \[Optional\]<br><br>Artifactory server ID configured using the config command.                                                                                                |
| --csv             | \[Optional\]<br><br>Path to a csv file with the usernames to delete. The first row of the file is the reserved for the cells' headers. It must include the "username" header. |
| Command arguments |                                                                                                                                                                               |
| users list        | Comma-separated list of usernames to delete. If the --csv command option is used, then this argument becomes optional.                                                        |

##### Example 1

Delete the users according to the usernames defined in the path/to/users.csv file.
```
jf rt users-delete --csv path/to/users.csv
```

##### Example 2

Delete the users according to the u1, u2 and u3 usernames.
```
jf rt users-delete "u1,u2,u3"
```
  
### Creating Groups

This command creates a new users group.

|                   |                                                                                |
|-------------------|--------------------------------------------------------------------------------|
| Command-name      | rt group-create                                                                |
| Abbreviation      | rt gc                                                                          |
| Command options   |                                                                                |
| --server-id       | \[Optional\]<br><br>Artifactory server ID configured using the config command. |
| Command arguments |                                                                                |
| group name        | The name of the group to create.                                               |

##### Example

Create a new group name **reviewers** .
```
jf rt group-create reviewers
```

### Adding Users to Groups

This command adds a list fo existing users to a group.

|                   |                                                                                |
|-------------------|--------------------------------------------------------------------------------|
| Command-name      | rt group-add-users                                                             |
| Abbreviation      | rt gau                                                                         |
| Command options   |                                                                                |
| --server-id       | \[Optional\]<br><br>Artifactory server ID configured using the config command. |
| Command arguments |                                                                                |
| group name        | The name of the group to add users to.                                         |
| users list        | Comma-seperated list of usernames to add to the specified group.               |

##### Example

Add to group reviewers the users with the following usernames: u1, u2 and u3.
```
jf rt group-add-users "reviewers" "u1,u2,u3"
```
  
### Deleting Groups

This command deletes a group.

|                   |                                                                                |
|-------------------|--------------------------------------------------------------------------------|
| Command-name      | rt group-delete                                                                |
| Abbreviation      | rt gdel                                                                        |
| Command options   |                                                                                |
| --server-id       | \[Optional\]<br><br>Artifactory server ID configured using the config command. |
| Command arguments |                                                                                |
| group name        | The name of the group to delete.                                               |

##### Example

Delete the **reviewers** group.
```
jf rt group-delete "reviewers"
```

### Managing Repositories

JFrog CLI offers a set of commands for managing Artifactory repositories. You can create, update and delete repositories. To make it easier to manage repositories, the commands which create and update the repositories accept a pre-defined configuration template file. This template file can also include variables. which can be later replaced with values, when creating or updating the repositories. The configuration template file is created using the **jf rt repo-template** command.

#### Creating or Configuration Template

This is an interactive command, which creates a configuration template file. This file should be used as an argument for the **jf rt repo-create** or the **jf rt repo-update** commands.

When using this command to create the template, you can also provide replaceable variable, instead of fixes values. Then, when the template is used to create or update repositories, values can be provided to replace the variables in the template.

|                   |                                                                                                               |
|-------------------|---------------------------------------------------------------------------------------------------------------|
| Command-name      | rt repo-template                                                                                              |
| Abbreviation      | rt rpt                                                                                                        |
| Command options   | The command has no options.                                                                                   |
| Command arguments |                                                                                                               |
| template path     | Specifies the local file system path for the template file created by the command. The file should not exist. |

##### Example

Create a configuration template, with a variable for the repository name. Then, create a repository using this template, and provide repository name to replace the variable.
```
$ jf rt repo-template template.json

Select the template type (press Tab for options): create
Insert the repository key > ${repo-name}
Select the repository class (press Tab for options): local
Select the repository's package type (press Tab for options): generic
You can type ":x" at any time to save and exit.
Select the next configuration key (press Tab for options): :x
[Info] Repository configuration template successfully created at template.json.
$
$ jf rt repo-create template.json --vars "repo-name=my-repo"
[Info] Creating local repository...
[Info] Done creating repository.
```

#### Creating / Updating Repositories

These two commands create a new repository and updates an existing a repository. Both commands accept as an argument a configuration template, which can be created by the **jf rt repo-template** command. The template also supports variables, which can be replaced with values, provided when it is used.

|                   |                                                                                                                                                                                       |
|-------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name      | rt repo-create / rt repo-update                                                                                                                                                       |
| Abbreviation      | rt rc / rt ru                                                                                                                                                                         |
| Command options   |                                                                                                                                                                                       |
| --server-id       | \[Optional\]<br><br>Artifactory server ID configured using the config command.                                                                                                        |
| --vars            | \[Optional\]<br><br>List of variables in the form of "key1=value1;key2=value2;..." to be replaced in the template. In the template, the variables should be used as follows: ${key1}. |
| Command arguments |                                                                                                                                                                                       |
| template path     | Specifies the local file system path for the template file to be used for the repository creation. The template can be created using the "jf rt rpt" command.                         |

##### Example 1

Create a repository, using the **template.json** file previously generated by the **repo-template** command.
```
jf rt repo-create template.json
```

##### Example 2

Update a repository, using the **template.json** file previously generated by the **repo-template** command.
```
jf rt repo-update template.json
```

##### Example 3

Update a repository, using the **template.json** file previously generated by the **repo-template** command. Replace the repo-name variable inside the template with a name for the updated repository.
```
jf rt repo-update template.json --vars "repo-name=my-repo"
```

#### Deleting Repositories

This command permanently deletes a repository, including all of its content.

|                   |                                                                                                            |
|-------------------|------------------------------------------------------------------------------------------------------------|
| Command name      | rt repo-delete                                                                                             |
| Abbreviation      | rt rdel                                                                                                    |
| Command options   |                                                                                                            |
| --server-id       | \[Optional\]<br><br>Artifactory server ID configured using the config command.                             |
| --quiet           | \[Default: $CI\]<br><br>Set to true to skip the delete confirmation message.                               |
| Command arguments |                                                                                                            |
| repository key    | Specifies the repositories that should be removed. You can use wildcards to specify multiple repositories. |

##### Example

Delete a repository from Artifactory.
```
jf rt repo-delete generic-local
```

### Managing Replications

JFrog CLI offers commands creating and deleting replication jobs in Artifactory. To make it easier to create replication jobs, the commands which creates the replication job accepts a pre-defined configuration template file. This template file can also include variables. which can be later replaced with values, when creating the replication job. The configuration template file is created using the **jf rt replication-template** command.

### Creating a Configuration Template

This command creates a configuration template file, which should be used as an argument for the **jf rt replication-create** command.

When using this command to create the template, you can also provide replaceable variable, instead of fixes values. Then, when the template is used to create replication jobs, values can be provided to replace the variables in the template.

|                   |                                                                                                               |
|-------------------|---------------------------------------------------------------------------------------------------------------|
| Command-name      | rt replication-template                                                                                       |
| Abbreviation      | rt rplt                                                                                                       |
| Command options   | The command has no options.                                                                                   |
| Command arguments |                                                                                                               |
| template path     | Specifies the local file system path for the template file created by the command. The file should not exist. |

##### Example

Create a configuration template, with two variables for the source and target repositories. Then, create a replication job using this template, and provide source and target repository names to replace the variables.
```
$ jf rt rplt template.json
Select replication job type (press Tab for options): push
Enter source repo key > ${source}
Enter target repo key > ${target}
Enter target server id (press Tab for options): my-server-id
Enter cron expression for frequency (for example, 0 0 12 * * ? will replicate daily) > 0 0 12 * * ?
You can type ":x" at any time to save and exit.
Select the next property > :x
[Info] Replication creation config template successfully created at template.json.
$
$ jf rt rplc template.json --vars "source=generic-local;target=generic-local"
[Info] Done creating replication job.
```
  

#### Creating Replication Jobs

This command creates a new replication job for a repository. The command accepts as an argument a configuration template, which can be created by the **jf rt replication-template** command. The template also supports variables, which can be replaced with values, provided when it is used.

|                   |                                                                                                                                                                                       |
|-------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name      | replication-create                                                                                                                                                                    |
| Abbreviation      | rt rplc                                                                                                                                                                               |
| Command options   |                                                                                                                                                                                       |
| --server-id       | \[Optional\]<br><br>Artifactory server ID configured using the config command.                                                                                                        |
| --vars            | \[Optional\]<br><br>List of variables in the form of "key1=value1;key2=value2;..." to be replaced in the template. In the template, the variables should be used as follows: ${key1}. |
| Command arguments |                                                                                                                                                                                       |
| template path     | Specifies the local file system path for the template file to be used for the replication job creation. The template can be created using the "jf rt rplt" command.                   |

##### Example 1

Create a replication job, using the **template.json** file previously generated by the **replication-template** command.
```
jf rt rplc template.json
```

##### Example 2

Update a replication job, using the **template.json** file previously generated by the **replication-template** command. Replace the source and target variables inside the template with the names of the replication source and target repositories.
```
jf rt rplc template.json --vars "source=my-source-repo;target=my-target-repo"
```

#### Deleting Replication jobs 

This command permanently deletes a replication jobs from a repository.

|                   |                                                                                |
|-------------------|--------------------------------------------------------------------------------|
| Command name      | rt replication-delete                                                          |
| Abbreviation      | rt rpldel                                                                      |
| Command options   |                                                                                |
| --server-id       | \[Optional\]<br><br>Artifactory server ID configured using the config command. |
| --quiet           | \[Default: $CI\]<br><br>Set to true to skip the delete confirmation message.   |
| Command arguments |                                                                                |
| repository key    | The repository from which the replications will be deleted.                    |

##### Example

Delete a repository from Artifactory.
```
jf rt rpldel my-repo-name
```

### Managing Permission Targets

JFrog CLI offers commands creating, updating and deleting permission targets in Artifactory. To make it easier to create and update permission targets, the commands which create and update the permission targets accept a pre-defined configuration template file. This template file can also include variables. which can be later replaced with values, when creating or updating the permission target. The configuration template file is created using the **jf rt permission-target-template** command.

#### Creating a Configuration Template

This command creates a configuration template file, which should be used as an argument for the **jf rt permission-target-create** and **jf rt permission-target-update** commands.

|                   |                                                                                                               |
|-------------------|---------------------------------------------------------------------------------------------------------------|
| Command-name      | rt permission-target-template                                                                                 |
| Abbreviation      | rt ptt                                                                                                        |
| Command options   | The command has no options.                                                                                   |
| Command arguments |                                                                                                               |
| template path     | Specifies the local file system path for the template file created by the command. The file should not exist. |

#### Creating / Updating Permission Targets

This command creates a new permission target. The command accepts as an argument a configuration template, which can be created by the **jf rt permission-target-template** command. The template also supports variables, which can be replaced with values, provided when it is used.

|                   |                                                                                                                                                                                       |
|-------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name      | permission-target-create / permission-target-update                                                                                                                                   |
| Abbreviation      | rt ptc / rt ptu                                                                                                                                                                       |
| Command options   |                                                                                                                                                                                       |
| --server-id       | \[Optional\]<br><br>Artifactory server ID configured using the config command.                                                                                                        |
| --vars            | \[Optional\]<br><br>List of variables in the form of "key1=value1;key2=value2;..." to be replaced in the template. In the template, the variables should be used as follows: ${key1}. |
| Command arguments |                                                                                                                                                                                       |
| template path     | Specifies the local file system path for the template file to be used for the permission target creation or update. The template can be created using the "jf rt ptt" command.        |

#### Deleting Permission Targets

This command permanently deletes a permission target.

|                        |                                                                                |
|------------------------|--------------------------------------------------------------------------------|
| Command name           | rt permission-target-delete                                                    |
| Abbreviation           | rt ptdel                                                                       |
| Command options        |                                                                                |
| --server-id            | \[Optional\]<br><br>Artifactory server ID configured using the config command. |
| --quiet                | \[Default: $CI\]<br><br>Set to true to skip the delete confirmation message.   |
| Command arguments      |                                                                                |
| permission target name | The permission target that should be removed.                                  |

## Using File Specs

To achieve complex file manipulations you may require several CLI commands. For example, you may need to upload several different sets of files to different repositories. To simplify the implementation of these complex manipulations, you can apply JFrog CLI **download**, **upload**, **move**, **copy** and **delete** commands with JFrog Artifactory using **--spec** option to replace the inline command arguments and options. Similarly, you can **create and update release bundles** by providing the `--spec` command option.  Each command uses an array of file specifications in JSON format with a corresponding schema as described in the sections below. Note that if any of these commands are issued using both inline options and the file specs, then the inline options override their counterparts specified in the file specs.

### File Spec Schemas

#### Copy and Move Commands Spec Schema

The file spec schema for the copy and move commands is as follows:
```
{
  "files": [
  {
    "pattern" or "aql": "[Mandatory]",
    "target": "[Mandatory]",
    "props": "[Optional]",
    "excludeProps": "[Optional]",
    "recursive": "[Optional, Default: 'true']",
    "flat": "[Optional, Default: 'false']",
    "exclusions": "[Optional, Applicable only when 'pattern' is specified]",
    "archiveEntries": "[Optional]",
    "build": "[Optional]",
    "bundle": "[Optional]",
    "validateSymlinks": "[Optional]",
    "sortBy": "[Optional]",
    "sortOrder": "[Optional, Default: 'asc']",
    "limit": "[Optional],
    "offset": [Optional] }
  ]
}
```

#### Download Command Spec Schema

The file spec schema for the download command is as follows:
```
{
  "files": [
  {
    "pattern" or "aql": "[Mandatory]",
    "target": "[Optional]",
    "props": "[Optional]",
    "excludeProps": "[Optional]",
    "recursive": "[Optional, Default: 'true']",
    "flat": "[Optional, Default: 'false']",
    "exclusions": "[Optional, Applicable only when 'pattern' is specified]",
    "archiveEntries": "[Optional]",
    "build": "[Optional]",
    "bundle": "[Optional]",
    "sortBy": "[Optional]",
    "sortOrder": "[Optional, Default: 'asc']",
    "limit": [Optional],
    "offset": [Optional] }
  ]
}
```

#### Create and Update Release Bundle Commands Spec Schema

The file spec schema for the create and update release bundle commands is as follows:
```
{
"files": [
  {
    "pattern" or "aql": "[Mandatory]",
    "target": "[Optional]",
    "props": "[Optional]",
    "targetProps": "[Optional]",
    "excludeProps": "[Optional]",
    "recursive": "[Optional, Default: 'true']",
    "flat": "[Optional, Default: 'false']",
    "exclusions": "[Optional, Applicable only when 'pattern' is specified]",
    "archiveEntries": "[Optional]",
    "build": "[Optional]",
    "bundle": "[Optional]",
    "sortBy": "[Optional]",
    "sortOrder": "[Optional, Default: 'asc']",
    "limit": [Optional],
    "offset": [Optional] }
  ]
}
```

#### Upload Command Spec Schema

The file spec schema for the upload command is as follows:
```
{
  "files": [
  {
    "pattern": "[Mandatory]",
    "target": "[Mandatory]",
    "targetProps": "[Optional]",
    "recursive": "[Optional, Default: 'true']",
    "flat": "[Optional, Default: 'true']",
    "regexp": "[Optional, Default: 'false']",
    "ant": "[Optional, Default: 'false']",
    "archive": "[Optional, Must be: 'zip']",
    "exclusions": "[Optional]" }
  ]
}
```

#### Search, Set-Props and Delete Commands Spec Schema

The file spec schema for the search and delete commands are as follows:
```
{
  "files": [
  {
    "pattern" or "aql": "[Mandatory]",
    "props": "[Optional]",
    "excludeProps": "[Optional]",
    "recursive": "[Optional, Default: 'true']",
    "exclusions": "[Optional, Applicable only when 'pattern' is specified]",
    "archiveEntries": "[Optional]",
    "build": "[Optional]",
    "bundle": "[Optional]",
    "sortBy": "[Optional]",
    "sortOrder": "[Optional, Default: 'asc']",
    "limit": [Optional],
    "offset": [Optional] }
  ]
}
```

##### Examples

The following examples can help you get started using File Specs.

##### Example 1:

Download all files located under the **all-my-frogs** directory in the **my-local-repo** repository to the **froggy** directory.
```
{
  "files": [
  {
    "pattern": "my-local-repo/all-my-frogs/",
    "target": "froggy/" }
  ]
}
```

##### Example 2:

Download all files located under the **all-my-frogs** directory in the **my-local-repo** repository to the **froggy** directory. Download only files which are artifacts of build number 5 of build **my-build** .

```
{
  "files": [
    {
      "pattern": "my-local-repo/all-my-frogs/",
      "target": "froggy/",
      "build": "my-build/5"
    }
  ]
}
```
  
##### Example 3:

Download all files retrieved by the AQL query to the **froggy** directory.
```
{
  "files": [
    {
      "aql": {
        "items.find": {
          "repo": "my-local-repo",
          "$or": [
            {
              "$and": [
                {
                  "path": {
                    "$match": "."
                  },
                  "name": {
                    "$match": "a1.in"
                  }
                }
              ]
            },
            {
              "$and": [
                {
                  "path": {
                    "$match": "*"
                  },
                  "name": {
                    "$match": "a1.in"
                  }
                }
              ]
            }
          ]
        }
      },
      "target": "froggy/"
    }
  ]
}
```

##### Example 4: Upload

1.  All zip files located under the **resources** directory to the **zip** folder, under the **all-my-frogs** repository.
    AND
2.  All TGZ files located under the **resources** directory to the **tgz folder, under the **all-my-frogs** repository.
3.  Tag all zip files with type = zip and status = ready.
4.  Tag all tgz files with type = tgz and status = ready.
    
```
{
  "files": [
    {
      "pattern": "resources/*.zip",
      "target": "all-my-frogs/zip/",
      "props": "type=zip;status=ready"
    },
    {
      "pattern": "resources/*.tgz",
      "target": "all-my-frogs/tgz/",
      "props": "type=tgz;status=ready"
    }
  ]
}
```

##### Example 5:

Upload all zip files located under the **resources** directory to the **zip** folder, under the **all-my-frogs** repository.
```
{
  "files": [
    {
      "pattern": "resources/*.zip",
      "target": "all-my-frogs/zip/"
    }
  ]
}
```

##### Example 6:

Package all files located (including subdirectories) under the **resources** directory into a zip archive named **archive.zip** , and upload it into the root of the **all-my-frogs** repository.
```
{
  "files": [
    {
      "pattern": "resources/",
      "archive": "zip",
      "target": "all-my-frogs/"
    }
  ]
}
```

###### **Example 7:** 

Download all files located under the **all-my-frogs** directory in the **my-local-repo** repository **except** for files with .txt extension and all files inside the **all-my-frogs** directory with the props. prefix.`

Notice that the exclude patterns do not include the repository.
```
{
    "files": [
     {
        "pattern": "my-local-repo/all-my-frogs/",
        "exclusions": ["*.txt","all-my-frog/props.*"]
     }
    ]
}
```

###### **Example 8:**

Download The latest file uploaded to the **all-my-frogs** directory in the **my-local-repo** repository.
```
{
    "files": [
     {
        "pattern": "my-local-repo/all-my-frogs/",
        "target": "all-my-frogs/files/",
        "sortBy": ["created"],
        "sortOrder": "desc",
        "limit": 1
     }
    ]
}
```

###### **Example 9:**

Search for the three largest files located under the **all-my-frogs** directory in the **my-local-repo** repository. If there are files with the same size, sort them "internally" by creation date.
```
{
    "files": [
     {
        "pattern": "my-local-repo/all-my-frogs/",
        "sortBy": ["size","created"],
        "sortOrder": "desc",
        "limit": 3
     }
    ]
}
```

###### **Example 10:**

Download The second-latest file uploaded to the **all-my-frogs** directory in the **my-local-repo** repository.
```
{
    "files": [
     {
        "pattern": "my-local-repo/all-my-frogs/",
        "target": "all-my-frogs/files/",
        "sortBy": ["created"],
        "sortOrder": "desc",
        "limit": 1,
        "offset": 1
     }
    ]
}
```

###### Example 11:

This example shows how to [delete artifacts in artifactory under specified path based on how old they are](https://stackoverflow.com/questions/58328701/delete-artifacts-in-artifactory-under-specified-path-based-on-how-old-they-are).

The following File Spec finds all the folders which match the following criteria:

1.  They are under the my-repo repository.
2.  They are inside a folder with a name that matches abc-*-xyz and is located at the root of the repository.
3.  Their name matches ver*
4.  They were created more than 7 days ago.
```
{
  "files": [
    {
      "aql": {
        "items.find": {
          "repo": "myrepo",
          "path": {"$match":"abc-*-xyz"},
          "name": {"$match":"ver*"},
          "type": "folder",
          "$or": [
            {
              "$and": [
                {
                  "created": { "$before":"7d" }
                }
              ]
            }
          ]
        }
      }
    }
  ]
}
```

###### Example 12

This example uses [Using Placeholders](https://jfrog.com/help/r/jfrog-cli/using-placeholders). For each .tgz file in the source directory, create a corresponding directory with the same name in the target repository and upload it there. For example, a file named froggy.tgz should be uploaded to my-local-rep/froggy. (froggy will be created a folder in Artifactory).
```
{
    "files": [
      {
        "pattern": "(*).tgz",
        "target": "my-local-repo/{1}/",
      }
    ]
}
```
  

###### Example 13

This examples uses [Using Placeholders](https://jfrog.com/help/r/jfrog-cli/using-placeholders). Upload all files whose name begins with "frog" to folder frogfiles in the target repository, but append its name with the text "-up". For example, a file called froggy.tgz should be renamed froggy.tgz-up.
```
{
    "files": [
      {
        "pattern": "(frog*)",
        "target": "my-local-repo/frogfiles/{1}-up",
        "recursive": "false"
      }
    ]
}
```
  

###### Example 14

The following two examples lead to the exact same outcome.  
The first one uses [Using Placeholders](https://jfrog.com/help/r/jfrog-cli/using-placeholders), while the second one does not. Both examples download all files from the generic-local repository to be under the ֿֿmy/local/path/ local file-system path, while maintaining the original Artifactory folder hierarchy. Notice the different flat values in the two examples.
```
{
    "files": [
      {
        "pattern": "generic-local/{*}",
        "target": "my/local/path/{1}",
        "flat": "true"
      }
    ]
}

{
    "files": [
      {
        "pattern": "generic-local/",
        "target": "my/local/path/",
        "flat": "false"
      }
    ]
}
```

### Schema Validation

[JSON schemas](https://json-schema.org/) allow you to annotate and validate JSON files. The JFrog File Spec schema is available in the [JSON Schema Store](https://www.schemastore.org/json/) catalog and in the following link: [https://github.com/jfrog/jfrog-cli/blob/v2/schema/filespec-schema.json](https://github.com/jfrog/jfrog-cli/blob/v2/schema/filespec-schema.json).

###### Using Jetbrains IDEs (Intellij IDEA, Webstorm, Goland, etc...)?

The File Spec schema is automatically applied to the following file patterns:

**/filespecs/*.json

\*filespec\*.json  

*.filespec

###### Using Visual Studio Code?

To apply the File Spec schema validation, install the [JFrog VS-Code extension](https://marketplace.visualstudio.com/items?itemName=JFrog.jfrog-vscode-extension).

Alternatively, copy the following to your settings.json file:

**settings.json**
```
"json.schemas": [
  {
    "fileMatch": ["**/filespecs/*.json", "\*filespec\*.json", "*.filespec"],
    "url": "https://raw.githubusercontent.com/jfrog/jfrog-cli/v2/schema/filespec-schema.json"
  }
]
```

## Downloading the Maven and Gradle Extractor JARs

For integrating with Maven and Gradle, JFrog CLI uses the build-info-extractor jars files. These jar files are downloaded by JFrog CLI from jcenter the first time they are needed.

If you're using JFrog CLI on a machine which has no access to the internet, you can configure JFrog CLI to download these jar files from an Artifactory instance. Here's how to configure Artifactory and JFrog CLI to download the jars files.

1.  Create a remote Maven repository in Artifactory and name it **extractors**. When creating the repository, configure it to proxy [https://releases.jfrog.io/artifactory/oss-release-local](https://releases.jfrog.io/artifactory/oss-release-local)
2.  Make sure that this Artifactory server is known to JFrog CLI, using the [jfrog c show](https://jfrog.com/help/r/jfrog-cli/showing-the-configured-servers) command. If not, configure it using the [jfrog c add](https://jfrog.com/help/r/jfrog-cli/Adding-and-Editing-Configured-Servers) command.
3.  Set the  **JFROG_CLI_EXTRACTORS_REMOTE** environment variable with the server ID of the Artifactory server you configured, followed by a slash, and then the name of the repository you created. For example **_my-rt-server/extractors_**

## Transferring Files Between Artifactory Servers

### Overview
The transfer-files command allows transferring (copying) all the files stored in one Artifactory instance to a different Artifactory instance. The command allows transferring the files stored in a single or multiple repositories. The command expects the relevant repository to already exist on the target instance and have the same name and type as the repositories on the source.

### Limitations
1. Artifacts in remote repositories caches are not transferred.
2. The files transfer process allows transferring files that were created or modified on the source instance after the process started. However, files that were deleted on the source instance after the process started, are not deleted on the target instance by the process.
3. The files transfer process allows transferring files that were created or modified on the source instance after the process started. The custom properties of those files are also updated on the target instance. However, if only the custom properties of those file were modified on the source, but not the files' content, the properties are not modified on the target instance by the process.
4. The source and target repositories should have the same name and type.
5. Since the file are pushed from the source to the target instance, the source instance must have network connection to the target.

### Before You Begin
1. Ensure that you can log in to the UI of both the source and target instances with users that have admin permissions and that you have the connection details (including credentials) to both instances.
2. Ensure that all the repositories on source Artifactory instance which files you'd like to transfer, also exist on the target instance, and have the same name and type on both instances.
3. Ensure that JFrog CLI is installed on a machine that has network access to both the source and target instances.

### Running the Transfer Process
#### Step 1 - Set Up the Source Instance for Files Transfer

To set up the source instance for files transfer, you must install the **data-transfer** user plugin in the primary node of the source instance. This section guides you through the installation steps.

1. Install JFrog CLI on the primary node machine of the source instance as described [here](https://jfrog.com/help/r/jfrog-cli/installing-the-data-transfer-user-plugin-on-the-source-machine-manually).
2. Configure the connection details of the source Artifactory instance with your admin credentials by running the following command from the terminal.
    ```
    jf c add source-server
    ```
3. Ensure that the **JFROG_HOME** environment variable is set and holds the value of JFrog installation directory. It usually points to the **/opt/jfrog** directory. In case the variable isn't set, set its value to point to the correct directory as described in the JFrog Product Directory Structure article.
4. If the source instance has internet access, you can install the **data-transfer** user plugin on the source machine automatically by running the following command from the terminal `jf rt transfer-plugin-install source-server`. If however the source instance has no internet access, install the plugin manually as described [here](https://jfrog.com/help/r/jfrog-cli/installing-the-data-transfer-user-plugin-on-the-source-machine-manually).
  
#### Step 2 - Push the Files from the Source to the Target Instance

Install JFrog CLI on any machine that has access to both the source and the target JFrog instances. To do this, follow the steps described [here](https://jfrog.com/help/r/jfrog-cli/installing-jfrog-cli-on-a-machine-with-network-access-to-the-source-and-target-machines).

Run the following command to start pushing the files from all the repositories in source instance to the target instance.

```
jf rt transfer-files source-server target-server
```

This command may take a few days to push all the files, depending on your system size and your network speed. While the command is running, It displays the transfer progress visually inside the terminal.

![transfer-files-progress](https://github.com/jfrog/jfrog-cli/raw/v2/documentation/images/transfer-files-progress.png)

If you're running the command in the background, you use the following command to view the transfer progress.

```
jf rt transfer-files --status
```

![transfer-files-status](https://github.com/jfrog/jfrog-cli/raw/v2/documentation/images/transfer-files-status.png)

In case you do not wish to transfer the files from all repositories, or wish to run the transfer in phases, you can use the `--include-repos` and `--exclude-repos` command options. Run the following command to see the usage of these options.

```
jf rt transfer-files -h
```

If the traffic between the source and target instance needs to be routed through an HTTPS proxy, refer to [this](https://jfrog.com/help/r/jfrog-cli/routing-the-traffic-from-the-source-to-the-target-through-an-https-proxy) section.

You can stop the transfer process by hitting on **CTRL+C** if the process is running in the foreground, or by running the following command, if you're running the process in the background.

```
jf rt transfer-files --stop
```

The process will continue from the point it stopped when you re-run the command.

While the file transfer is running, monitor the load on your source instance, and if needed, reduce the transfer speed or increase it for better performance. For more information, see the [Controlling the File Transfer Speed](https://jfrog.com/help/r/jfrog-cli/controlling-the-file-transfer-speed) section.

A path to an errors summary file will be printed at the end of the run, referring to a generated CSV file. Each line on the summary CSV represents an error of a file that failed to be transferred. On subsequent executions of the `jf rt transfer-files` command, JFrog CLI will attempt to transfer these files again.

Once the `jf rt transfer-files` command finishes transferring the files, you can run it again to transfer files which were created or modified during the transfer. You can run the command as many times as needed. Subsequent executions of the command will also attempt to transfer files failed to be transferred during previous executions of the command.

**Note:**
> Read more about how the transfer files works in [this](https://jfrog.com/help/r/jfrog-cli/how-does-files-transfer-work) section.
---

## Installing the data-transfer User Plugin on the Source Machine Manually
To install the **data-transfer** user plugin on the source machine manually, follow these steps.
1. Download the following two files from a machine that has internet access. Download **data-transfer.jar** from https://releases.jfrog.io/artifactory/jfrog-releases/data-transfer/[RELEASE]/lib/data-transfer.jar and **dataTransfer.groovy** from https://releases.jfrog.io/artifactory/jfrog-releases/data-transfer/[RELEASE]/dataTransfer.groovy
2. Create a new directory on the primary node machine of the source instance and place the two files you downloaded inside this directory.
3. Install the **data-transfer** user plugin by running the following command from the terminal. Replace the *[plugin files dir]* token with the full path to the directory which includes the plugin files you downloaded.
    ```
    jf rt transfer-plugin-install source-server --dir "[plugin files dir]"
    ```

## Installing JFrog CLI on the Source Instance Machine

Install JFrog CLI on your source instance by using one of the [#JFrog CLI Installers]. For example:

```curl -fL https://install-cli.jfrog.io | sh```

**Note**

If the source instance is running as a docker container, and you're not able to install JFrog CLI while inside the container, follow these steps.

1. Connect to the host machine through the terminal.
2. Download the JFrog CLI executable into the correct directory by running this command:

    ```curl -fL https://getcli.jfrog.io/v2-jf | sh```
3. Copy the JFrog CLI executable you've just downloaded into the container, by running the following docker command. Make sure to replace *[the container name]* with the name of the container.

    ```docker cp jf [the container name]:/usr/bin/jf```
4. Connect to the container and run the following command to ensure JFrog CLI is installed:

    ```jf -v```

## How Does Files Transfer Work?
### Files Transfer Phases
The `jf rt transfer-files` command pushes the files from the source instance to the target instance as follows:

- The files are pushed for each repository, one by one in sequence.
- For each repository, the process includes the following three phases:
  - **Phase 1** pushes all the files in the repository to the target.
  - **Phase 2** pushes files which have been created or modified after phase 1 started running (diffs).
  - **Phase 3** attempts to push files which failed to be transferred in earlier phases (Phase 1 or Phase 2) or in previous executions of the command.
- If Phase 1 finished running for a specific repository, and you run the `jf rt transfer-files` command again, only **Phase 2** and **Phase 3** will be triggered. You can run the `jf rt transfer-files` as many times as needed, till you are ready to move your traffic to the target instance permanently. In any subsequent run of the command, **Phase 2** will transfer the newly created and modified files and **Phase 3** will retry transferring files which failed to be transferred in previous phases and also in previous runs of the command.

**Using Replication**
> To help reduce the time it takes for Phase 2 to run, you may configure Event Based Push Replication for some or all of the local repositories on the source instance. With Replication configured, when files are created or updated on the source repository, they are immediately replicated to the corresponding repository on the target instance.
The replication can be configured at any time. Before, during or after the files transfer process.
---

### Files Transfer State
You can run the `jf rt transfer-files` command multiple times. This is needed to allow transferring files which have been created or updated after previous command executions. To achieve this, JFrog CLI stores the current state of the files transfer process in a directory named **transfer** located under the JFrog CLI home directory. You can usually find this directory at this location `~/.jfrog/transfer`.

JFrog CLI uses the state stored in this directory to avoid repeating transfer actions performed in previous executions of the command. For example, once **Phase 1** is completed for a specific repository, subsequent executions of the command will skip **Phase 1** and run **Phase 2** and **Phase 3** only.

In case you'd like to ignore the stored state, and restart the files transfer from scratch, you can add the `--ignore-state` option to the `jf rt transfer-files` command.

## Installing JFrog CLI on a Machine with Network Access to the Source and Target Machines
It is recommended to run the `transfer-files` command from a machine that has network access to the source Artifactory URL. This allows spreading the transfer load on all the Artifactory cluster nodes. This machine should also have network access to the target Artifactory URL.

Follows these steps to installing JFrog CLI on that machine.

1. Install JFrog CLI by using one of the [#JFrog CLI Installers]. For example:

    ```
    curl -fL https://install-cli.jfrog.io | sh
    ```

2. If your source instance is accessible only through an HTTP/HTTPS proxy, set the proxy environment variable as described [#here].
3. Configure the connection details of the source Artifactory instance with your admin credentials. Run the following command and follow the instructions.

    ```
    jf c add source-server
    ```

4. Configure the connection details of the target Artifactory instance as follows.

    ```
    jf c add target-server
    ```

## Controlling the File Transfer Speed
The `jf rt transfer-files` command pushes the binaries from the source instance to the target instance. This transfer can take days, depending on the size of the total data transferred, the network bandwidth between the source and the target instance, and additional factors.

Since the process is expected to run while the source instance is still being used, monitor the instance to ensure that the transfer does not add too much load to it. Also, you might decide to increase the load for faster a transfer rate, while you monitor the transfer. This section describes how to control the file transfer speed.

By default, the `jf rt transfer-files` command uses 8 working threads to push files from the source instance to the target instance. Reducing this value will cause slower transfer speed and lower load on the source instance, and increasing it will do the opposite. We therefore recommend increasing it gradually. This value can be changed while the `jf rt transfer-files` command is running. There's no need to stop the process to change the total of working threads. The new value set will be cached by JFrog CLI and also used for subsequent runs from the same machine. To set the value, simply run the following interactive command from a new terminal window on the same machine which runs the `jf rt transfer-files` command.

```
jf rt transfer-settings
```

**Build-info repositories**
> When transferring files in build-info repositories, JFrog CLI limits the total of working threads to 8. This is done in order to limit the load on the target instance while transferring build-info.
---

## Routing the Traffic from the Source to the Target Through an HTTPS Proxy
The `jf rt transfer-files` command pushes the files directly from the source to the target instance over the network. In case the traffic from the source instance needs to be routed through an HTTPS proxy, follow these steps.

1. Define the proxy details in the source instance UI as described in the [Managing Proxies documentation](https://jfrog.com/help/r/jfrog-platform-administration-documentation/managing-proxies).

2. When running the `jf rt transfer-files` command, add the `--proxy-key` option to the command, with Proxy Key you configured in the UI as the option value. For example, if the Proxy Key you configured is my-proxy-key, run the command as follows:

    ```
    jf rt transfer-files my-source my-target --proxy-key my-proxy-key
    ```
