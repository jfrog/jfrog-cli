JFrog CLI : CLI for JFrog Mission Control
=========================================

Overview
--------

This page describes how to use JFrog CLI with JFrog Mission Control.

Read more about JFrog CLI [here](https://jfrog.com/help/r/jfrog-cli).

Syntax
------

	$ jfrog mc command-name arguments global-options command-options

Where:

|     |     |
| --- | --- |
| command-name | The command to execute. Note that you can use either the full command name or its abbreviation. |
| global-options | A set of global options that may be used for all commands:<br><br>`--url:` (Optional) Mission Control URL.<br><br>`--access-token:` (Optional) Mission Control admin access token.<br><br>**Use the config command**<br><br>To avoid having to set these for every command, you may set them once using the [config](#CLIforJFrogMissionControl-Configuration) command and then omit them for every following command. |
| command-options | A set of options corresponding to the command |
| arguments | A set of arguments corresponding to the command |

  

* * *

Commands
--------

The following sections describe the commands available in the JFrog CLI for use with JFrog Mission Control.

### Adding a JPD 

|     |     |
| --- | --- |
| Command name | jpd-add |
| Abbreviation | ja  |
| Description | Adds a JPD to Mission Control |
| Command arguments |     |
| Config | Path to a JSON configuration file containing the JPD details. |
| Command options | The command accepts no options, other than the global options. |

#### **Config JSON schema**
```
{
  "name" : "jpd-0",
  "url" : "http://jpd:8080/test",
  "token" : "some-token",
  "location" : {
    "city_name" : "San Francisco",
    "country_code" : "US",
    "latitude" : 37.7749,
    "longitude" : 122.4194
  },
  "tags" : \[ "tag0", "tag1" \]
}
```
  

**Example**

	jf mc ja path/to/jpd/config.json

### Deleting a JPD

|     |     |
| --- | --- |
| Command name | jpd-delete |
| Abbreviation | jd  |
| Description | Delete a JPD from Mission Control. |
| Command arguments |     |
| JPD ID | The ID of the JPD to be removed from Mission Control. |
|     |     |
| Command options | The command accepts no options, other than the global options. |

**Example**

	jf mc jd my-jpd-id

### Acquiring a License

|     |     |
| --- | --- |
| Command name | license-acquire |
| Abbreviation | la  |
| Description | Acquire a license from the specified bucket and mark it as taken by the provided name. |
| Command arguments |     |
| Bucket ID | Bucket name or identifier to acquire license from. |
| Name | A custom name used to mark the license as taken. Can be a JPD ID or a temporary name. If the license does not end up being used by a JPD, this is the name that should be used to release the license. |
|     |     |
| Command options | The command accepts no options, other than the global options. |

**Examples**

##### Example 1

Assign a license from the _my-bucket-id_ and mark it as taken by _my-unique-name_.

	jf mc la my-bucket-id my-unique-name

### Deploying a License

|     |     |
| --- | --- |
| Command name | license-deploy |
| Abbreviation | ld  |
| Description | pecified bucket to an existing JPD. You may also deploy a number of licenses to an Artifactory HA. |
| Command arguments |     |
| Bucket ID | Bucket name or identifier to deploy licenses from. |
| JPD ID | An existing JPD's ID. |
|     |     |
| Command options |     |
| --license-count | \[Default: 1\]<br><br>The number of licenses to deploy. Minimum value is 1. |

**Example**

Deploy a single license from _my-bucket-id_ on _my-jpd-id_.

	jf mc ld my-bucket-id my-jpd-id

### Releasing a License

|     |     |
| --- | --- |
| Command name | license-release |
| Abbreviation | lr  |
| Description | Release all licenses of a JPD and return them to the specified bucket. |
| Command arguments |     |
| Bucket ID | Bucket name or identifier to release all of its licenses. |
| JPD ID | If the license is used by a JPD, pass the JPD's ID. If the license was only acquired but is not used, pass the name it was acquired with. |
|     |     |
| Command options | The command accepts no options, other than the global options. |

**Example**

Releases all licenses of _my-jpd-id_ to to _my-bucket-id_.

	jf mc lr my-bucket-id my-jpd-id
