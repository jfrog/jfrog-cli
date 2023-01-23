JFrog CLI : CLI for JFrog Distribution
======================================


Overview
--------

This page describes how to use JFrog CLI with [JFrog Distribution](https://jfrog-staging-external.fluidtopics.net/r/help/Software-Distribution-Distribution).

Read more about JFrog CLI [here](https://jfrog-staging-external.fluidtopics.net/r/help/JFrog-CLI/JFrog-CLI).

### Syntax

When used with JFrog Distribution, JFrog CLI uses the following syntax:

	$ jf ds command-name global-options command-options arguments

### Managing Access Keys



### Commands

The following sections describe the commands available in the JFrog CLI for use with JFrog Distribution. 

### Creating or updating an unsigned Release Bundle

This commands creates and updates an unsigned Release Bundle on JFrog Distribution.

---
**Note**
> This commands require version 2.0 or higher of[JFrog Distribution](https://jfrog-staging-external.fluidtopics.net/r/help/Software-Distribution-Distribution).
---

|     |     |
| --- | --- |
| Command-name | release-bundle-create / release-bundle-update |
| Abbreviation | rbc / rbu |
| Command options |     |
| --server-id | \[Optional\]<br><br>Artifactory server ID configured using the config command. |
| --spec | \[Optional\]<br><br>Path to a file spec. For more details, please refer to[Using File Specs](https://www.jfrog.com/confluence/display/CLI/CLI+for+JFrog+Artifactory#CLIforJFrogArtifactory-UsingFileSpecs). |
| --spec-vars | \[Optional\]<br><br>List of variables in the form of "key1=value1;key2=value2;..." to be replaced in the File Spec. In the File Spec, the variables should be used as follows: ${key1}. |
| --target-props | \[Optional\]<br><br>The list of properties, in the form of key1=value1;key2=value2,..., to be added to the artifacts after distribution of the release bundle. |
| --target | \[Optional\]<br><br>The target path for distributed artifacts on the edge node. If not specified, the artifacts will have the same path and name on the edge node, as on the source Artifactory server. For flexibility in specifying the distribution path, you can include [placeholders](https://www.jfrog.com/confluence/display/CLI/CLI+for+JFrog+Artifactory#CLIforJFrogArtifactory-UsingPlaceholders) in the form of {1}, {2} which are replaced by corresponding tokens in the pattern path that are enclosed in parenthesis. |
| --dry-run | \[Default: false\]<br><br>Set to true to disable communication with JFrog Distribution. |
| --sign | \[Default: false\]<br><br>If set to true, automatically signs the release bundle version. |
| --passphrase | \[Optional\]<br><br>The passphrase for the signing key. |
| --desc | \[Optional\]<br><br>Description of the release bundle. |
| --release-notes-path | \[Optional\]<br><br>Path to a file describes the release notes for the release bundle version. |
| --release-notes-syntax | \[Default: plain_text\]<br><br>The syntax for the release notes. Can be one of _markdown_, _asciidoc_, or _plain_text_. |
| --exclusions | \[Optional\]<br><br>A list of Semicolon-separated exclude path patterns, to be excluded from the Release Bundle. Allows using wildcards. |
| --repo | \[Optional\]<br><br>A repository name at source Artifactory to store release bundle artifacts in. If not provided, Artifactory will use the default one. |
| --insecure-tls | \[Default: false\]<br><br>Set to true to skip TLS certificates verification. |
| --detailed-summary | \[Default: false\]<br><br>Set to true to return the SHA256 value of the release bundle manifest. |
| Command arguments |     |
| release bundle name | The name of the release bundle. |
| release bundle version | The release bundle version. |
| pattern | Specifies the source path in Artifactory, from which the artifacts should be bundled, in the following format: &lt;repository name&gt;/&lt;repository path&gt;. You can use wildcards to specify multiple artifacts. This argument should not be sent along with the _--spec_ option. |

##### Examples

##### Example 1

Create a release bundle with name myApp and version 1.0.0. The release bundle will include the files defined in the File Spec specified by the --spec option.

	jf ds rbc --spec=/path/to/rb-spec.json myApp 1.0.0

##### Example 2

Create a release bundle with name myApp and version 1.0.0. The release bundle will include the files defined in the File Spec specified by the --spec option. GPG sign the release bundle after it is created.

	jf ds rbc --spec=/path/to/rb-spec.json --sign myApp 1.0.0

##### Example 3

Update the release bundle with name myApp and version 1.0.0. The release bundle will include the files defined in the File Spec specified by the --spec option.

	jf ds rbu --spec=/path/to/rb-spec.json myApp 1.0.0

##### Example 4

Update the release bundle with name myApp and version 1.0.0. The release bundle will include all the zip files inside the zip folder, located at the root of the _my-local-repo_ repository.

	jf ds rbu myApp 1.0.0 "my-local-repo/zips/*.zip"

##### Example 5

Update the release bundle with name myApp and version 1.0.0. The release bundle will include all the zip files inside the zip folder, located at the root of the _my-local-repo_ repository. The files will be distributed on the Edge Node to the _target-zips_ folder, under the root of the _my-target-repo_ repository.

	jf ds rbu myApp 1.0.0 "my-local-repo/zips/*.zip" --target my-target-repo/target-zips/

##### Example 6

This example uses [placeholders](https://www.jfrog.com/confluence/display/CLI/CLI+for+JFrog+Artifactory#CLIforJFrogArtifactory-UsingPlaceholders). It creates the release bundle with name myApp and version 1.0.0. The release bundle will include all the zip files inside the zip folder, located at the root of the _my-local-repo_ repository. The files will be distributed on the Edge Node to the _target-zips_ folder, under the root of the _my-target-repo_ repository. In addition, the distributed files will be renamed on the Edge Node, by adding _-target_ to the name of each file.

	jf ds rbc myApp 1.0.0 "my-local-repo/zips/(*).zip" --target "my-target-repo/target-zips/{1}-target.zip"

### Signing an Existing Release Bundle

This command GPG signs an existing Release Bundle on JFrog Distribution.

---
**Note**
> These commands require version 2.0 or higher of[JFrog Distribution](https://www.jfrog.com/confluence/display/JFROG/JFrog+Distribution).
---

|     |     |
| --- | --- |
| Command-name | release-bundle-sign |
| Abbreviation | rbs |
| Command options |     |
| --server-id | \[Optional\]<br><br>Artifactory server ID configured using the config command. |
| --passphrase | \[Optional\]<br><br>The passphrase for the signing key. |
| --repo | \[Optional\]<br><br>A repository name at source Artifactory to store release bundle artifacts in. If not provided, Artifactory will use the default one. |
| --insecure-tls | \[Default: false\]<br><br>Set to true to skip TLS certificates verification. |
| --detailed-summary | \[Default: false\]<br><br>Set to true to return the SHA256 value of the release bundle manifest. |
| Command arguments |     |
| release bundle name | The name of the release bundle. |
| release bundle version | The release bundle version. |

##### Examples

##### Example 1

GPG sign the release bundle with name myApp and version 1.0.0.

	jf ds rbs --passphrase="&lt;passphrase&gt;" myApp 1.0.0 

### Distributing a Release Bundle

This command distributes a release bundle to the Edge Nodes.

---
**Note**
> These commands require version 2.0 or higher of[JFrog Distribution](https://www.jfrog.com/confluence/display/JFROG/JFrog+Distribution).
---

|     |     |
| --- | --- |
| Command-name | release-bundle-distribute |
| Abbreviation | rbd |
| Command options |     |
| --server-id | \[Optional\]<br><br>Artifactory server ID configured using the config command. |
| --sync | \[Default: false\]<br><br>Set to true to enable sync distribution (the command execution will end when the distribution process ends). |
| --max-wait-minutes | \[Default: 60\]<br><br>Max minutes to wait for sync distribution. |
| --create-repo | \[Default: false\]<br><br>Set to true to create the repository on the edge if it does not exist. |
| --dry-run | \[Default: false\]<br><br>Set to true to disable communication with JFrog Distribution. |
| --dist-rules | \[Optional\]<br><br>Path to a file, which includes the Distribution Rules in a JSON format.<br><br>**Distribution Rules JSON structure**<br><br>{<br>    "distribution_rules": \[<br>      {<br>        "site_name": "DC-1",<br>        "city_name": "New-York",<br>        "country_codes": \["1"\]<br>      },<br>      {<br>        "site_name": "DC-2",<br>        "city_name": "Tel-Aviv",<br>        "country_codes": \["972"\]<br>      }<br>    \]<br>}<br><br>The Distribution Rules format also supports wildcards. For example:<br><br>{<br>    "distribution_rules": \[<br>      {<br>        "site_name": "*",<br>        "city_name": "*",<br>        "country_codes": \["*"\]<br>      }<br>    \]<br>} |
| --site | \[Default: *\]<br><br>Wildcard filter for site name. |
| --city | \[Default: *\]<br><br>Wildcard filter for site city name. |
| --country-codes | \[Default: *\]<br><br>Semicolon-separated list of wildcard filters for site country codes. |
| --insecure-tls | \[Default: false\]<br><br>Set to true to skip TLS certificates verification. |
| Command arguments |     |
| release bundle name | The name of the release bundle. |
| release bundle version | The release bundle version. |

##### Examples

##### Example 1

Distribute the release bundle with name myApp and version 1.0.0. Use the distribution rules defined in the specified file.

	jf ds rbd --dist-rules=/path/to/dist-rules.json myApp 1.0.0

  

### Deleting a Release Bundle

This command deletes a Release Bundle from the Edge Nodes and optionally from Distribution as well.

---
**Note**
> These commands require version 2.0 or higher of[JFrog Distribution](https://www.jfrog.com/confluence/display/JFROG/JFrog+Distribution).
---

|     |     |
| --- | --- |
| Command-name | release-bundle-delete |
| Abbreviation | rbdel |
| Command options |     |
| --server-id | \[Optional\]<br><br>Artifactory server ID configured using the config command. |
| --sync | \[Default: false\]<br><br>Set to true to enable sync deletion (the command execution will end when the deletion process ends). |
| --max-wait-minutes | \[Default: 60\]<br><br>Max minutes to wait for sync deletion. |
| --dry-run | \[Default: false\]<br><br>Set to true to disable communication with JFrog Distribution. |
| --dist-rules | \[Optional\]<br><br>Path to a file, which includes the distribution rules in a JSON format. |
| --site | \[Default: *\]<br><br>Wildcard filter for site name. |
| --city | \[Default: *\]<br><br>Wildcard filter for site city name. |
| --country-codes | \[Default: *\]<br><br>Semicolon-separated list of wildcard filters for site country codes. |
| --delete-from-dist | \[Default: false\]<br><br>Set to true to delete release bundle version in JFrog Distribution itself after deletion is complete in the specified Edge nodes. |
| --quiet | \[Default: false\]<br><br>Set to true to skip the delete confirmation message. |
| --insecure-tls | \[Default: false\]<br><br>Set to true to skip TLS certificates verification. |
| Command arguments |     |
| release bundle name | The name of the release bundle. |
| release bundle version | The release bundle version. |

##### Examples

##### Example 1

Delete the release bundle with name myApp and version 1.0.0 from the Edge Nodes only, according to the definition in the distribution rules file.

	jf ds rbdel --dist-rules=/path/to/dist-rules.json myApp 1.0.0

##### Example 2

Delete the release bundle with name myApp and version 1.0.0 from the Edge Nodes, according to the definition in the distribution rules file. The release bundle will also be deleted from the Distribution service itself.

	jf ds rbdel --delete-from-dist --dist-rules=/path/to/dist-rules.json myApp 1.0.0


