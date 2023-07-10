JFrog CLI : CLI for JFrog Curation
======================================


Overview
--------

This page describes how to use JFrog CLI with JFrog Curation.

Read more about JFrog CLI [here](https://jfrog.com/help/r/jfrog-cli).

---
**Note**
> JFrog Curation is only available since [Artifactory 7.63.2](https://jfrog.com/help/r/jfrog-release-information/artifactory-7.63.2-cloud) And [Xray 3.78.9](https://jfrog.com/help/r/jfrog-release-information/xray-3.78.9).

---

### Syntax

When used with JFrog Distribution, JFrog CLI uses the following syntax:

	$ jf ca command-name command-options 

Where:


|                 |                                                                                                 |
|-----------------|-------------------------------------------------------------------------------------------------|
| command-name    | The command to execute. Note that you can use either the full command name or its abbreviation. |
| command-options | A set of options corresponding to the command                                                   |



### Commands

The following sections describe the commands available in the JFrog CLI for use with JFrog Curation.

Curation-Audit
---------------------
**Note**
>The command _curation-audit_ currently supports only [npm](https://www.npmjs.com/) projects.

The _jf curation-audit_ command enables developers to scan project dependencies to find packages that were blocked by the JFrog curation service. This command provides developers with more detailed information, such as whether the blocked package is the project’s direct dependency or is a transitive dependency. This information helps developers to resolve blocked packages more efficiently as they will be able to make a more informative decision based on what Policy violation occurred and what exactly needs to be resolved.

For each blocked package the CLI provides the violated Curation Policies, The command builds a deep dependencies graph for the project, and requests the Curation status by a HEAD request for each node in the tree. It uses the package manager that is used in the project to build the dependencies graph.
Before running the command, first, you need to connect the JFrog CLI to your JFrog Platform instance with the _jf c add_ command. Then ensure your project is configured in the JFrog CLI with the repository you would like to resolve dependencies from. To do this, set the repository with the _jf npmc_ command inside the project directory.




|                       |                                                                                                                                   |
|-----------------------|-----------------------------------------------------------------------------------------------------------------------------------|
| **Command name**      | curation-audit                                                                                                                    |
| **Abbreviation**      | ca                                                                                                                                |
| **Command options**   |                                                                                                                                   |
| --format              | \[Default: table\]<br><br>Defines the output format of the command. Acceptable values are: table and json.                        |
| --working-dirs        | \[Optional\]<br><br>A comma separated list of relative working directories, to determine the audit targets locations.             |
| --threads             | \[Default: 10\]<br><br>The number of parallel threads used to determine the curation status for each package in the project tree. |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |

#### **Output Example**

![image](images/jf-ca-output.png)


**Example 1**

Curation-Audit the project at the current directory. Show all known packages blocked by curation policies.

	jf curation-audit

**Example 2**

Curation-Audit the projects at the paths mentioned in the "working-dirs" option. Show all known packages blocked by curation policies for both projects in separate tables.

	jf curation-audit --working-dirs="/path/to/project/npm_project1,/path/to/project/npm_project2"

**Example 1**

Curation-Audit the project at the current directory using 5 threads to check packages curation status in parallel. Show all known packages blocked by curation policies.

	jf curation-audit --threads=5
