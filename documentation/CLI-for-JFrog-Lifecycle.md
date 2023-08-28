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

### Distributing a release bundle

This command distributes a Release Bundle v2 to an edge node.

|                        |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
|------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Command-name           | release-bundle-distribute                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| Abbreviation           | rbd                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| Command options        |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| --create-repo          | \[Default: false\]<br><br>Set to true to create the repository on the edge if it does not exist.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| --dry-run              | \[Default: false\]<br><br>Set to true to disable communication with JFrog Distribution.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| --dist-rules           | \[Optional\]<br><br>Path to a file, which includes the Distribution Rules in a JSON format.<br><br>**Distribution Rules JSON structure**<br><br>{<br>    "distribution_rules": \[<br>      {<br>        "site_name": "DC-1",<br>        "city_name": "New-York",<br>        "country_codes": \["1"\]<br>      },<br>      {<br>        "site_name": "DC-2",<br>        "city_name": "Tel-Aviv",<br>        "country_codes": \["972"\]<br>      }<br>    \]<br>}<br><br>The Distribution Rules format also supports wildcards. For example:<br><br>{<br>    "distribution_rules": \[<br>      {<br>        "site_name": "*",<br>        "city_name": "*",<br>        "country_codes": \["*"\]<br>      }<br>    \]<br>} |
| --site                 | \[Default: *\]<br><br>Wildcard filter for site name.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| --city                 | \[Default: *\]<br><br>Wildcard filter for site city name.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| --country-codes        | \[Default: *\]<br><br>Semicolon-separated list of wildcard filters for site country codes.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| --insecure-tls         | \[Default: false\]<br><br>Set to true to skip TLS certificates verification.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| --mapping-pattern      | \[Optional\]<br><br>Specify along with 'mapping-target' to distribute artifacts to a different path on the edge node. You can use wildcards to specify multiple artifacts.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| --mapping-target       | \[Optional\]<br><br>The target path for distributed artifacts on the edge node. If not specified, the artifacts will have the same path and name on the edge node, as on the source Artifactory server. For flexibility in specifying the distribution path, you can include [placeholders](https://www.jfrog.com/confluence/display/CLI/CLI+for+JFrog+Artifactory#CLIforJFrogArtifactory-UsingPlaceholders) in the form of {1}, {2} which are replaced by corresponding tokens in the pattern path that are enclosed in parenthesis.                                                                                                                                                                                  |
| Command arguments      |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| release bundle name    | Name of the Release Bundle to distribute.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| release bundle version | Version of the Release Bundle to distribute.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |


#### Examples

#### Example 1
Distribute the Release Bundle v2 named myApp with version 1.0.0. Use the distribution rules defined in the specified file.

	jf rbd --dist-rules=/path/to/dist-rules.json myApp 1.0.0

#### Example 2

Distribute the Release Bundle v2 named myApp with version 1.0.0 using the default distribution rules.
Map files under the 'source' directory to be placed under the 'target' directory. 

	jf rbd --dist-rules=/path/to/dist-rules.json --mapping-pattern="(*)/source/(*)" --mapping-target="{1}/target/{2}" myApp 1.0.0