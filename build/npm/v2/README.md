<p align="center">
  <a href="https://jfrog.com/">
    <img alt="JFrog" src="https://github.com/jfrog/jfrog-cli/blob/v2/build/npm/v2/assets/jfrog.jpg?raw=true" width="200">
  </a>
</p>

# JFrog CLI

[Website](http://www.jfrog.com)  •  [Docs](https://docs.jfrog.com/integrations/docs/jfrog-cli)  •  [Issues](https://github.com/jfrog/jfrog-cli/issues)  •  [Blog](https://jfrog.com/blog/)  •  [We're Hiring](https://join.jfrog.com/)  •  [Artifactory Free Trial](https://jfrog.com/artifactory/free-trial/)

> [!IMPORTANT]
> **Please migrate to [`jfrog-cli-v2-jf`](https://www.npmjs.com/package/jfrog-cli-v2-jf).**
>
> This package (`jfrog-cli-v2`) installs the CLI under the **`jfrog`** executable name. The actively maintained package, [`jfrog-cli-v2-jf`](https://www.npmjs.com/package/jfrog-cli-v2-jf), installs the very same CLI but under the shorter **`jf`** executable name, which is now the standard across JFrog's documentation, examples, and CI/CD integrations.
>
> **What this means for you:**
> - The command you type changes from `jfrog ...` to `jf ...` — update your scripts, pipelines, aliases, and shell completions accordingly.
> - Both executables share the same configuration, so your existing servers and settings continue to work.
> - You can install both side by side during the transition, but to avoid confusion we recommend switching fully to `jf`.
>
> To migrate:
> ```bash
> npm uninstall -g jfrog-cli-v2
> npm install -g jfrog-cli-v2-jf
> ```

## Overview

JFrog CLI is a compact and smart client that provides a simple interface that automates access to *Artifactory*, *Xray*,
*Distribution*, *Pipelines* and *Mission Control* through their respective REST APIs.
By using the JFrog CLI, you can greatly simplify your automation scripts making them more readable and easier to
maintain.
Several features of the JFrog CLI makes your scripts more efficient and reliable.

## Requirements

npm version 5.0.0 or above.

## Install with npm:

  ```bash
  npm install -g jfrog-cli-v2
  
  # When running as root user:
  
  npm install -g -unsafe-perm jfrog-cli-v2

  ```
  
