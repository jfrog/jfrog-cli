# Apple Bundle Structure README

This README file serves as a guide to maintaining the integrity of the Apple bundle structure required for macOS applications. It is crucial to keep this file and adhere to the outlined structure to ensure the application functions correctly on macOS.

## Structure Overview

The Apple bundle for a macOS application typically has the following directory structure:### Key Components
```
              YOUR_APP.app
               ├── Contents
                   ├── MacOS
                   │   └── YOUR_APP (executable file)
                   └── Info.plist

```
- **YOUR_APP.app**: This is the root directory of your application bundle. Replace `YOUR_APP` with the name of your application.

- **Contents**: A mandatory directory that contains all the files needed by the application.

- **MacOS**: This directory should contain the executable file for your application. The name of the executable should match the `YOUR_APP` part of your application bundle's name.

- **Info.plist**: A required file that contains configuration and permissions for your application. It informs the macOS about how your app should be treated and what capabilities it has.

### Important Notes

- **Do Not Delete**: This README file and the structure it describes are essential for the application's deployment and functionality on macOS. Removing or altering the structure may result in application failures.

- **Executable File**: Ensure your application's executable file is placed inside the `MacOS` directory. The executable's name must match the `YOUR_APP` portion of your application bundle's name for macOS to recognize and launch it correctly.

- **Info.plist Configuration**: Properly configure the `Info.plist` file according to your application's needs. This file includes critical information such as the app version, display name, permissions, and more.

By adhering to this structure and guidelines, you ensure that your macOS application is packaged correctly for distribution and use.