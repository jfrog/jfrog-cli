[![JFrog CLI](images/jfrog-cli-intro.png)](https://github.com/jfrog/jfrog-cli)

<div align="center">

# JFrog CLI

[![Scanned by Frogbot](https://raw.github.com/jfrog/frogbot/master/images/frogbot-badge.svg)](https://github.com/jfrog/frogbot#readme)
[![Go Report Card](https://goreportcard.com/badge/github.com/jfrog/jfrog-cli)](https://goreportcard.com/report/github.com/jfrog/jfrog-cli)
[![license](https://img.shields.io/badge/License-Apache_2.0-blue.svg?style=flat)](https://raw.githubusercontent.com/jfrog/jfrog-cli/v2/LICENSE) [![](https://img.shields.io/badge/Docs-%F0%9F%93%96-blue)](https://www.jfrog.com/confluence/display/CLI/JFrog+CLI)
[![Go version](https://img.shields.io/github/go-mod/go-version/jfrog/jfrog-cli)](https://tip.golang.org/doc/go1.20)

</div>

<details>
    <summary>Tests status</summary>
    <table>
        <tr>
            <th></th>
            <th width="100">V2</th>
            <th width="100">DEV</th>
        </tr>
        <div align="center">
            <tr>
                <td><img src="./images/artifactory.png" alt="artifactory"> Artifactory</td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/artifactoryTests.yml?query=branch%3Av2">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/artifactoryTests.yml/badge.svg?branch=v2" alt="">
                    </a>
                </td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/artifactoryTests.yml?query=branch%3Adev">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/artifactoryTests.yml/badge.svg?branch=dev" alt="">
                    </a>
                </td>
            </tr>
            <tr>
                <td><img src="./images/xray.png" alt="xray"> Xray</td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/xrayTests.yml?query=branch%3Av2">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/xrayTests.yml/badge.svg?branch=v2" alt="">
                    </a>
                </td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/xrayTests.yml?query=branch%3Adev">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/xrayTests.yml/badge.svg?branch=dev" alt="">
                    </a>
                </td>
            </tr>
            <tr>
                <td><img src="./images/distribution.png" alt="distribution"> Distribution</td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/distributionTests.yml?query=branch%3Av2">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/distributionTests.yml/badge.svg?branch=v2" alt="">
                    </a>
                </td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/distributionTests.yml?query=branch%3Adev">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/distributionTests.yml/badge.svg?branch=dev" alt="">
                    </a>
                </td>
            </tr>
            <tr>
                <td><img src="./images/access.png" alt="access"> Access</td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/accessTests.yml?query=branch%3Av2">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/accessTests.yml/badge.svg?branch=v2" alt="">
                    </a>
                </td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/accessTests.yml?query=branch%3Adev">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/accessTests.yml/badge.svg?branch=dev" alt="">
                    </a>
                </td>
            </tr>
            <tr>
                <td><img src="./images/maven.png" alt="maven"> Maven</td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/mavenTests.yml?query=branch%3Av2">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/mavenTests.yml/badge.svg?branch=v2" alt="">
                    </a>
                </td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/mavenTests.yml?query=branch%3Adev">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/mavenTests.yml/badge.svg?branch=dev" alt="">
                    </a>
                </td>
            </tr>
            <tr>
                <td><img src="./images/gradle.png" alt="gradle"> Gradle</td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/gradleTests.yml?query=branch%3Av2">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/gradleTests.yml/badge.svg?branch=v2" alt="">
                    </a>
                </td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/gradleTests.yml?query=branch%3Adev">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/gradleTests.yml/badge.svg?branch=dev" alt="">
                    </a>
                </td>
            </tr>
            <tr>
                <td><img src="./images/npm.png" alt="npm"> npm</td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/npmTests.yml?query=branch%3Av2">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/npmTests.yml/badge.svg?branch=v2" alt="">
                    </a>
                </td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/npmTests.yml?query=branch%3Adev">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/npmTests.yml/badge.svg?branch=dev" alt="">
                    </a>
                </td>
            </tr>
            <tr>
                <td><img src="./images/docker.png" alt="docker"> Docker</td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/dockerTests.yml?query=branch%3Av2">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/dockerTests.yml/badge.svg?branch=v2" alt="">
                    </a>
                </td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/dockerTests.yml?query=branch%3Adev">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/dockerTests.yml/badge.svg?branch=dev" alt="">
                    </a>
                </td>
            </tr>
            <tr>
               <td><img src="./images/podman.png" alt="podman"> Podman</td>
               <td>
                  <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/podmanTests.yml?query=branch%3Av2">
                     <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/podmanTests.yml/badge.svg?branch=v2" alt="">
                  </a>
               </td>
               <td>
                  <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/podmanTests.yml?query=branch%3Adev">
                     <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/podmanTests.yml/badge.svg?branch=dev" alt="">
                  </a>
               </td>
            </tr>
            <tr>
                <td><img src="./images/nuget.png" alt="nuget"> NuGet</td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/nugetTests.yml?query=branch%3Av2">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/nugetTests.yml/badge.svg?branch=v2" alt="">
                    </a>
                </td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/nugetTests.yml?query=branch%3Adev">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/nugetTests.yml/badge.svg?branch=dev" alt="">
                    </a>
                </td>
            </tr>
            <tr>
                <td><img src="./images/python.png" alt="python"> Python</td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/pythonTests.yml?query=branch%3Av2">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/pythonTests.yml/badge.svg?branch=v2" alt="">
                    </a>
                </td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/pythonTests.yml?query=branch%3Adev">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/pythonTests.yml/badge.svg?branch=dev" alt="">
                    </a>
                </td>
            </tr>
            <tr>
                <td><img src="./images/go.png" alt="go"> Go</td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/goTests.yml?query=branch%3Av2">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/goTests.yml/badge.svg?branch=v2" alt="">
                    </a>
                </td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/goTests.yml?query=branch%3Adev">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/goTests.yml/badge.svg?branch=dev" alt="">
                    </a>
                </td>
            </tr>
            <tr>
                <td> 📃 Scripts</td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/scriptTests.yml?query=branch%3Av2">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/goTests.yml/badge.svg?branch=v2" alt="">
                    </a>
                </td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/scriptTests.yml?query=branch%3Adev">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/goTests.yml/badge.svg?branch=dev" alt="">
                    </a>
                </td>
            </tr>
            <tr>
                <td>📊 Code Analysis</td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/analysis.yml?query=branch%3Av2">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/analysis.yml/badge.svg?branch=v2" alt="">
                    </a>
                </td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/analysis.yml?query=branch%3Adev">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/analysis.yml/badge.svg?branch=dev" alt="">
                    </a>
                </td>
            </tr>
            <tr>
                <td>🔌 Plugins</td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/pluginsTests.yml?query=branch%3Av2">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/pluginsTests.yml/badge.svg?branch=v2" alt="">
                    </a>
                </td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/pluginsTests.yml?query=branch%3Adev">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/pluginsTests.yml/badge.svg?branch=dev" alt="">
                    </a>
                </td>
            </tr>
            <tr>
                <td>☁️ Transfer To Cloud</td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/transferTests.yml?query=branch%3Av2">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/transferTests.yml/badge.svg?branch=v2" alt="">
                    </a>
                </td>
                <td>
                    <a href="https://github.com/jfrog/jfrog-cli/actions/workflows/transferTests.yml?query=branch%3Adev">
                        <img src="https://github.com/jfrog/jfrog-cli/actions/workflows/transferTests.yml/badge.svg?branch=dev" alt="">
                    </a>
                </td>
            </tr>
        </div>
    </table>
</details>

# Table of Contents

- [Overview](#overview)
- [Download and Installation](#download-and-installation)
- [Code Contributions](#code-contributions)
- [Using JFrog CLI](#using-jfrog-cli)
- [JFrog CLI Plugins](#jfrog-cli-plugins)
- [Release Notes](#release-notes)

# Overview

JFrog CLI is a compact and smart client that provides a simple interface that automates access to _Artifactory_ and
_Mission Control_ through their respective REST APIs.
By using the JFrog CLI, you can greatly simplify your automation scripts making them more readable and easier to
maintain.
Several features of the JFrog CLI makes your scripts more efficient and reliable:

- Multithreaded upload and download of artifacts make builds run faster
- Checksum optimization reduces redundant file transfers
- Wildcards and regular expressions give you an easy way to collect all the artifacts you wish to upload or download.
- "Dry run" gives you a preview of file transfer operations before you actually run them

# Download and Installation

You can either install JFrog CLI using one of the supported installers or download its executable directly. Visit
the [Install JFrog CLI Page](https://jfrog.com/getcli/) for details.

# Code Contributions

We welcome pull requests from the community. To help us improve this project, please read our [contribution](CONTRIBUTING.md) guide.

# Using JFrog CLI

JFrog CLI can be used for a variety of functions with Artifactory, Xray and Mission Control,
and has a dedicated set of commands for each product.
To learn how to use JFrog CLI, please visit
the [JFrog CLI User Guide](https://jfrog.com/help/r/jfrog-cli).

# JFrog CLI Plugins

JFrog CLI plugins support enhancing the functionality of JFrog CLI to meet the specific user and organization needs. The
source code of a plugin is maintained as an open source Go project on GitHub. All public plugins are registered in JFrog
CLI's Plugins Registry, which is hosted in the [jfrog-cli-plugins-reg](https://github.com/jfrog/jfrog-cli-plugins-reg)
GitHub repository. We encourage you, as developers, to create plugins and share them publicly with the rest of the
community. Read more about this in the [JFrog CLI Plugin Developer Guide](guides/jfrog-cli-plugins-developer-guide.md).

# Release Notes

The release notes are available [here](https://github.com/jfrog/jfrog-cli/releases).
