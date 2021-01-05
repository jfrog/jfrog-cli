# JFrog CLI Plugin Developer Guide
## Overview
**JFrog CLI Plugins** allow enhancing the functionality of JFrog CLI to meet the specific user and organization needs. The source code of a plugin is maintained as an open source Go project on GitHub. All public plugins are registered in [JFrog CLI's Plugins Registry](https://github.com/jfrog/jfrog-cli-plugins-reg). We encourage you, as developers, to create plugins and share them publically with the rest of the community. When a plugin is included in the registry, it becomes publicly available and can be installed using JFrog CLI. Version 1.41.1 or above is required. Plugins can be installed using the following JFrog CLI command:
```
$ jfrog plugin install the-plugin-name
```
This article guides you through the process of creating and publishing your own JFrog CLI Plugin.

## Creating your first plugin
### Prepare your development machine
* Make sure Go 1.14 or above is installed on your local machine and is included in your system PATH.
* Make sure git is installed on your local machine and is included in your system PATH. 

### Build and run your first plugin
1. Go to [https://github.com/jfrog/jfrog-cli-plugin-template.git](https://github.com/jfrog/jfrog-cli-plugin-template.git).
2. Press the **Use this template** button to create a new repository. You may name it as you like.
3. Clone your new repository to your local machine. For example:
```
$ git clone https://github.com/jfrog/jfrog-cli-plugin-template.git
```
4. Run the following commands, to build and run the template plugin.
```
$ cd jfrog-cli-plugin-template
$ go build -o hello-frog
$ ./hello-frog --help
$ ./hello-frog hello --help
$ ./hello-frog hello Yey!
```
5. Open the plugin code with your favorite IDE and start having fun.

## What can plugins do?
Well, plugins can do almost anything. The sky is the limit.
1. You have access to most of the JFrog CLI code base. This is because your plugin code depends on the [https://github.com/jfrog/jfrog-cli-core](https://github.com/jfrog/jfrog-cli-core) module. It is a depedency declared in your project's *go.mod* file. Feel free to explore the *jfrog-cli-core* code base, and use it as part of your plugin.
2. You can also add other Go packages to your *go.mod* and use them in your code.

## Guidelines for developing and publishing a plugin
To publish your plugin, you need to include it in [JFrog CLI's Plugins Registry](https://github.com/jfrog/jfrog-cli-plugins-reg). Please make sure your plugin meets the following guidelines before publishing it.

* Read the [Developer Terms](https://github.com/jfrog/jfrog-cli-plugins-reg/blob/master/DEVELOPERS_TERMS.md) document. You'll be asked to accept it before your plugin becomes available.
* **Code structure.** Make sure the plugin code is structured similarly to the jfrog cli plugin template. Specifically, it should include a *commands* package, and a separate file for each command.
* **Tests.** The plugin code should include a series of thorough tests. Use the [jfrog-cli-plugin-template](https://github.com/jfrog/jfrog-cli-plugin-template.git) as a reference on how the tests should be included as part of the source code. The tests should be executed using the following Go command while inside the root directory of the plugin project. **Note:** The Registry verifies the plugin and tries to run your plugin tests using the following command. ```go vet -v ./... && go test -v ./...```
* **Code formatting.** To make sure the code formatted properly, run the following go command on your plugin sources, while inside the root of your project directory. ```go fmt ./...```
* **Plugin name.** The plugin name should include only lower-case characters, numbers and dashes. The name length should not exceed 30 characters. It is recommended to use a short name for the users' convenience, but also make sure that its name hints on its functionality.
* **Create a Readme.** Make sure that your plugin code includes a README.md file and place it in the root of the repository. The README needs to be structured according to the [jfrog-cli-plugin-template]((https://github.com/jfrog/jfrog-cli-plugin-template.git)) README. It needs to include all the information and relevant details for the relevant plugin users..
* **Consider create a tag for your plugin sources.** Although this is not mandatory, we recommend creating a tag for your GitHub repository before publishing the plugin. You can then provide this tag to the Registry when publishing the plugin, to make sure the correct code is built.
* **Plugin version**. Make sure that your built plugin has the correct version. The version is declared as part of the plugin sources. To check your plugin version, run the plugin executable with the -v option. For example: ```./my-plugin -v```. The plugin version should have a **v** prefix. For example ```v1.0.0``` and it should follow the semantic versioning guidelines.

## Publishing a new plugin
### General
To publish a new plugin, you need to register the plugin in the JFrog CLI Plugins Registry. The registry is hosted in [https://github.com/jfrog/jfrog-cli-plugins-reg](https://github.com/jfrog/jfrog-cli-plugins-reg). The registry includes a descriptor file in YAML format for each registered plugin, inside the *plugins* directory. To include your plugin in the registry, create a pull request to add the plugin descriptor file for your plugin according to this file name format: *your-plugin-name.yml*.

### Important notes
* Please make sure that the extension of your plugin descriptor file is **yml** and not **yaml**.
* Please make sure your pull request includes only one file, which is the plugin descriptor.

### YAML format example
```
# Mandatory:
pluginName: hello-frog
version: v1.0.0
repository: https://github.com/my-org/my-amazing-plugin
maintainers:
    - github-username1
    - github-username2
# Optional:
relativePath: build-info-analyzer
# You may set either branch or tag, but noth both
branch: my-release-branch
tag: my-release-tag
```

## Publishing a new version of a published plugin
To publish a new version of your plugin, all you need to do is create a pull request, which updates the version inside your plugin descriptor file. If needed, your change can also include either the branch or tag. 
