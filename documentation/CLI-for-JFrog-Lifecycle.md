# JFrog CLI: CLI for JFrog Release Lifecycle Management

## Overview

This page describes how to use JFrog CLI with [JFrog Release Lifecycle Management](https://jfrog.com/help/r/jfrog-artifactory-documentation/jfrog-release-lifecycle-management-solution).

Read more about JFrog CLI [here](https://jfrog.com/help/r/jfrog-cli).

---
**Note**
> JFrog Release Lifecycle Management is only available since [Artifactory 7.63.2](https://jfrog.com/help/r/jfrog-release-information/artifactory-7.63.2-cloud).
---

### Commands

The following sections describe the commands available in JFrog CLI when performing Release Lifecycle Management operations on Release Bundles v2.

### Creating a Release Bundle v2 from builds or from existing Release Bundles

This command creates a Release Bundle v2 from a published build-info or from an existing Release Bundle.  
1. To create a Release Bundle from published build-infos, provide the `--builds` option, which accepts a path to a file using the following JSON format:
    ```json
    {
      "builds": [
        {
          "name": "build name",
          "number": "build number",
          "project": "project"
        }
      ]
    }
    ```
    `number` is optional (if left empty, the latest build will be used)
    
    `project` is optional (if left empty, the default project will be used)

2. To create a Release Bundle v2 from existing Release Bundles, provide the `--release-bundles` option, which accepts a path to a file using the following JSON format:
    ```json
    {
      "releaseBundles": [
        {
          "name": "release bundle name",
          "version": "release bundle version",
          "project": "project"
        }
      ]
    }
    ```
    `project` is optional (if left empty, the default project will be used)

|                        |                                                                                                                                                                                                                                                                                       |
|------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name           | release-bundle-create                                                                                                                                                                                                                                                                 |
| Abbreviation           | rbc                                                                                                                                                                                                                                                                                   |
| Command options        |                                                                                                                                                                                                                                                                                       |
| --builds               | \[Optional\]<br><br>Path to a JSON file containing information about the source builds from which to create a Release Bundle.                                                                                                                                                            |
| --project              | \[Optional\]<br><br>JFrog Project key associated with the Release Bundle version.                                                                                                                                                                                                           |
| --release-bundles      | \[Optional\]<br><br>Path to a JSON file containing information about the source Release Bundles from which to create a Release Bundle.                                                                                                                                                   |
| --server-id            | \[Optional\]<br><br>Platform server ID configured using the `jf c add` command.                                                                                                                                                                                                           |
| --signing-key          | \[Mandatory\]<br><br>The GPG/RSA key-pair name given in Artifactory.                                                                                                                                                                                                                  |
| --sync                 | \[Default: false\]<br><br>Set to true to run synchronously.                                                                                                                                                                                                                           |
| Command arguments      |                                                                                                                                                                                                                                                                                       |
| release bundle name    | Name of the newly created Release Bundle.                                                                                                                                                                                                                                             |
| release bundle version | Version of the newly created Release Bundle.                                                                                                                                                                                                                                          |

#### Examples

##### Example 1

Create a Release Bundle v2 with the name "myApp" and version "1.0.0", with signing key pair "myKeyPair".
The Release Bundle will include the artifacts of the builds that were provided in the builds spec. 
```
jf rbc --builds=/path/to/builds-spec.json --signing-key=myKeyPair myApp 1.0.0
```
##### Example 2

Create a Release Bundle v2 with the name "myApp" and version "1.0.0", with signing key pair "myKeyPair".
The Release Bundle will include the artifacts of the Release Bundles that were provided in the Release Bundles spec.
```
jf rbc --spec=/path/to/release-bundles-spec.json --signing-key=myKeyPair myApp 1.0.0
```
##### Example 3

Create a Release Bundle v2 synchronously with the name "myApp" and version "1.0.0", in project "project0", with signing key pair "myKeyPair".
The Release Bundle will include the artifacts of the Release Bundles that were provided in the Release Bundles spec.
```
jf rbc --spec=/path/to/release-bundles-spec.json --signing-key=myKeyPair --sync=true --project=project0 myApp 1.0.0
```
### Promoting a Release Bundle v2

This command promotes a Release Bundle v2 to a target environment.

|                        |                                                                                                                                                                                                                        |
|------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name           | release-bundle-promote                                                                                                                                                                                                 |
| Abbreviation           | rbp                                                                                                                                                                                                                    |
| Command options        |                                                                                                                                                                                                                        |
| --overwrite            | \[Default: false\]<br><br>Set to true to replace artifacts with the same name but a different checksum, if such already exist at the promotion targets. By default, the promotion is stopped when a conflict occurs.|
| --project              | \[Optional\]<br><br>Project key associated with the Release Bundle version.                                                                                                                                            |
| --server-id            | \[Optional\]<br><br>Platform server ID configured using the config command.                                                                                                                                            |
| --signing-key          | \[Mandatory\]<br><br>The GPG/RSA key-pair name given in Artifactory.                                                                                                                                                   |
| --sync                 | \[Default: false\]<br><br>Set to true to run synchronously.                                                                                                                                                            |
| Command arguments      |                                                                                                                                                                                                                        |
| release bundle name    | Name of the Release Bundle to promote.                                                                                                                                                                                 |
| release bundle version | Version of the Release Bundle to promote.                                                                                                                                                                              |
| environment            | Name of the target environment for the promotion.                                                                                                                                                                      |

#### Examples
##### Example 1

Promote a Release Bundle v2 named "myApp" version "1.0.0" to environment "PROD".
Use signing key pair "myKeyPair".
```
jf rbp --signing-key=myKeyPair myApp 1.0.0 PROD
```
##### Example 2

Promote a Release Bundle v2 synchronously to environment "PROD".
The Release Bundle is named "myApp", version "1.0.0", of project "project0".
Use signing key pair "myKeyPair" and overwrite in case of conflicts.
```
jf rbp --signing-key=myKeyPair --project=project0 --overwrite=true --sync=true myApp 1.0.0 PROD
```
