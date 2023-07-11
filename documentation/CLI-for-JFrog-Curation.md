JFrog CLI : CLI for JFrog Curation
======================================


Overview
--------
JFrog Curation enables you to block malicious or risky open-source packages entering your software supply chain, preventing consumption in any software pipeline.
What can you do with Curation?
Track the open-source packages downloaded by your organization to gain centralized visibility and control.
Prevent harmful packages from getting into your software development pipelines.
Protect against known and unknown threats, allowing only trusted software packages into your SDLC.
Create policies to block packages with known vulnerabilities, malicious code, operational risk, or license compliance issues.
For more information on JFrog Curation and how to set it up, see the JFrog Curation general documentation. (need to add link after it goes live).

JFrog Curation requires [Xray version 3.78.9](https://jfrog.com/help/r/jfrog-release-information/xray-3.78.9) and above, and [Artifactory version 7.63.2](https://jfrog.com/help/r/jfrog-release-information/artifactory-7.63.2-cloud) and above. It also requires Enterprise X and Enterprise + subscriptions.

---

### Commands

Audit your Project with JFrog Curation


**Note**
> The **curation-audit** command currently only supports npm projects.

The **jf curation-audit** command enables developers to scan project dependencies to find packages that were blocked by the JFrog curation service. This command provides developers with more detailed information, such as whether the blocked package is the project’s direct dependency or is a transitive dependency. This information helps developers to resolve blocked packages more efficiently as they will be able to make a more informative decision based on what Policy violation occurred and what exactly needs to be resolved.

For each blocked package the CLI provides the violated Curation Policies, The command builds a deep dependencies graph for the project, and requests the Curation status by a HEAD request for each node in the tree. It uses the package manager that is used in the project to build the dependencies graph.

Before running the command, do the following:
1) Connect the JFrog CLI to your JFrog Platform instance by running the jf c add command.
2) Ensure your project is configured in the JFrog CLI with the repository you would like to resolve dependencies from. Set the repository with the jf npmc command inside the project directory.
   <br><br>


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

Audit the project in the current directory. Displays all known packages that were blocked by Curation Policies.

	jf curation-audit

**Example 2**

Audit the projects according to the specific paths defined in the "working-dirs" option. Displays all known packages that were blocked by Curation Policies for all projects. The data is displayed in separate tables.

	jf curation-audit --working-dirs="/path/to/project/npm_project1,/path/to/project/npm_project2"

**Example 3**

Audit the project in the current directory using 5 threads to check the packages Curation status in parallel. Displays all known packages blocked by Curation Policies.

	jf curation-audit --threads=5
