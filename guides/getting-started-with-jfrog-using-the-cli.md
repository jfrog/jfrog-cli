<h1 align="center">Getting started with JFrog using the CLI</h1>

<div align="center">
    <img src="../images/jfrog-cli-intro.png"></img>
</div>

## Overview
This short guide helps you get started using your new JFrog environment, after you set it up using one of the following commands.<br><br>
**Linux / MacOS**<br>
```
curl -fL https://getcli.jfrog.io/setup | sh
```
**Windows**<br>
```
powershell "Start-Process -Wait -Verb RunAs powershell '-NoProfile iwr https://releases.jfrog.io/artifactory/jfrog-cli/v2-jf/[RELEASE]/jfrog-cli-windows-amd64/jf.exe -OutFile $env:SYSTEMROOT\system32\jf.exe'" ; jf setup
```
## Getting started
### Init your project
* ``cd`` into the root directory of your source code project
* Run -
```
jf project init
```
### Scan your code & software packages
#### From the terminal
* Scan your project dependencies for security vulnerabilities, while inside the root directory of your project -
```
jf audit
```
* Scan any software package on your local machine for security vulnerabilities - 
```
jf scan path/to/dir/or/package
```
* Scan any local docker image on you machine for security vulnerabilities by running -
```
jf docker scan <image name>:<image tag>
```
#### From the IDE
If you're using VS Code, IntelliJ IDEA, WebStorm, PyCharm, Android Studio or GoLand - 
* Open the IDE
* Install the JFrog extension or plugin
* View the JFrog panel
### Build & deploy
Depending on the build tool you use, run one of the following commands. Feel free to modify the build tool arguments and options - 
**npm - Install**
```
jf npm install
```
**npm - Publish**
```
jf npm publish
```
**Maven - Install and deploy**
```
jf mvn install deploy
```
**Gradle - Install and deploy**
```
jf gradle artifactoryP
```
**pip - Install**
```
jf pip install
```
**pip - Publish**
```
jf rt u <path/to/package/file> default-pypi-local
```
**Go - Build**
```
jf go build
```
**Go - Publish**
```
jf gp v1.0.0
```
### Publish build-info
To publish build-info for your build to your JFrog environment, run the following command - 
```
jf rt bp
```
### More
* Read more about [JFrog CLI](https://www.jfrog.com/confluence/display/CLI/JFrog+CLI) & [JFrog's IDE Integrations](https://www.jfrog.com/confluence/display/JFROG/IDE+Integration)
* **Your feedback is important!** We'd love to get your feedback and answer any questions you may have.
Communicate with us through [GitHub](https://github.com/jfrog/jfrog-cli/issues).
