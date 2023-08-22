# Transfer Artifactory Configuration and Files fron Self-Hosted to JFrog Cloud

## General
JFrog provides you the ability to migrate from a self-hosted JFrog Platform installation to JFrog Cloud so that you can seamlessly transition into JFrog Cloud. You can use the JFrog CLI to transfer the Artifactory configuration settings and binaries to JFrog Cloud.

JFrog Cloud provides the same cutting edge functionalities of a self-hosted JFrog Platform Deployment (JPD), without the overhead of managing the databases and systems. If you are an existing JFrog self-hosted customer, you might want to move to JFrog Cloud to ease operations. JFrog provides a solution that allows you to replicate your self-hosted JPD to a JFrog Cloud JPD painlessly.

The Artifactory Transfer solution, currently transfers the config and data of JFrog Artifactory only. Other products such as JFrog Xray, Distribution and JFrog Pipelines are currently not supported by this solution.

In this page, we refer the source self-hosted instance as the source instance, and the target JFrog Cloud instance as the target instance.

## Artifactory Version Support

The Artifactory Transfer solution is supported for any version of Artifactory 7.x and Artifactory version 6.23.21 and above.

## Transfer phases

The transfer process include two phases, that you must perform in the following order:

1. **Configuration Transfer:** Transfers the configuration entities like users, permissions, and repositories from the source instance to the target instance.

2. **File Transfer:** Transfers the files (binaries) stored in the source instance repositories to the target instance repositories.

---
**Note**
> * Files that are cached by remote repositories aren't transferred.
> * The content of Artifactory's Trash Can isn't transferred.
---

You can do both steps while the source instance is in use. No downtime on the source instance is required while the transfer is in progress.

## Limitations

The following limitations need to be kept in mind before you start the migration process

1. The Archive Search Enabled feature is not supported on JFrog Cloud.

2. Artifactory System Properties are not transferred and JFrog Cloud defaults are applied after the transfer.

3. User plugins are not supported on JFrog Cloud.

4. Artifact Cold Storage is not supported in JFrog Cloud.

5. Artifacts in remote repositories caches are not transferred.

6. Federated repositories are transferred without their federation members. After the transfer, you'll need to reconfigure the federation as described in the Federated Repositories documentation.Federated Repositories

7. Docker repositories with names that include dots aren't allowed in JFrog Cloud.

8. Artifact properties with value longer than 2.4K characters are not supported in JFrog Cloud. Such properties are generally seen in Conan artifacts. The artifacts will transferred without the properties in this case. A report with these artifacts will become available to you at the end of the transfer.

9. The files transfer process allows transferring files that were created or modified on the source instance after the process started. However, files that were deleted on the source instance after the process started, are not deleted on the target instance by the process.

10. The files transfer process allows transferring files that were created or modified on the source instance after the process started. The custom properties of those files are also updated on the target instance . However, if only the custom properties of those file were modified on the source, but not the files' content, the properties are not modified on the target instance by the process.


## Before you begin

1. If your source instance hosts files that are larger than 25 GB, they will be blocked during the transfer. To learn how to check whether large files are hosted by your source instance, and what to do in that case, read [this section](#Transferring files larger than 25 GB).

2. Ensure that you can login to the UI of both the source and target instances with users that have admin permissions.

3. Ensure that the target instance license does not support less features than the source instance license.

4. Run the files transfer pre-checks as described [here](#running-pre-checks-before-initiating-the-files-transfer-process).

5. Ensure that all the remote repositories on the source Artifactory instance have network access to their destination URL once they are created in the target instance. Even if one remote or federated repository does not have access, the configuration transfer operation will be canceled. You do have the option of excluding specific repositories from being transferred.

6. Ensure that all the replications configured on the source Artifactory instance have network access to their destination URL once they are created in the target instance.

7. Ensure that you have a user that can log in to MyJFrog. For more information on MyJFrog, see MyJFrog Cloud Portal.

8. Ensure that you can login to the primary node of your source instance through a terminal.

## Running the transfer process

Perform the following steps to transfer configuration and artifacts from the source to the target instance. You must run the steps in the exact sequence and do not run any of the commands in parallel.

### Step 1: Enabling configuration transfer on the target instance

By default, the target does not have the APIs required for the configuration transfer. Enabling the target instance for configuration transfer is done through MyJFrog. Once the configuration transfer is complete, you must disable the configuration transfer in MyJFrog as described in Step 4 below.

 :::Warning

* Enabling configuration transfer will trigger a shutdown of JFrog Xray, Distribution, Insights and Pipelines in the cloud and these services will therefore become unavailable. Once you disable the configuration transfer later on in the process, these services will be started up again.

* Enabling configuration transfer will scale down JFrog Artifactory, which will reduce its available resources. Once you disable the configuration transfer later on in the process, Artifactory will be scaled up again.

Follow the below steps for enabling the configuration transfer.

1. Log in to MyJFrog.

2. Click on **Settings**.

3. Under the **Transfer Artifactory Configuration from Self-Hosted to Cloud** section, click on the **acknowledgment**checkbox. You cannot enable configuration transfer until you select the checkbox.

4. If you have an Enterprise+ subscription with more than one Artifactory instance, select the target instance from the drop-down menu.

5. Toggle **Enable Configuration Transfer** on to enable the transfer. The process may take a few minutes to complete.

6. The configuration transfer is now enabled and you can continue with the transfer process.

### Step 2: Set up the source instance for pushing files to the target instance

To set up the source instance, you must install the data-transfer user plugin in the primary node of the source instance. This section guides you through the installation steps.

1. Install JFrog CLI on the primary node machine of the source instance as described [here](#installing-jfrog-cli-on-the-source-instance-machine).

2. Configure the connection details of the source Artifactory instance with your admin credentials by running the following command from the terminal.

```sh
jf c add source-server
```

3. Ensure that the **JFROG\_HOME** environment variable is set and holds the value of JFrog installation directory. It usually points to the **/opt/jfrog** directory. In case the variable isn't set, set its value to point to the correct directory as described in the JFrog Product Directory Structure article.System Directories

If the source instance has internet access, follow this single step:

**Download and install the data-transfer user plugin by running the following command from the terminal.**

```sh
jf rt transfer-plugin-install source-server
```

If the source instance has no internet access, follow these steps instead.

1. Download the following two files from a machine that has internet access:

Download **data-transfer.jar** from [https://releases.jfrog.io/artifactory/jfrog-releases/data-transfer/\[RELEASE\]/lib/data-transfer.jar](https://releases.jfrog.io/artifactory/jfrog-releases/data-transfer/%5BRELEASE%5D/lib/data-transfer.jar).

Download **dataTransfer.groovy** from [https://releases.jfrog.io/artifactory/jfrog-releases/data-transfer/\[RELEASE\]/dataTransfer.groovy](https://releases.jfrog.io/artifactory/jfrog-releases/data-transfer/%5BRELEASE%5D/dataTransfer.groovy).

2. Create a new directory on the primary node machine of the source instance and place the two files you downloaded inside this directory.

3. Install the data-transfer user plugin by running the following command from the terminal. Replace the `<plugin files dir>` token with the full path to the directory which includes the plugin files you downloaded.

```sh
jf rt transfer-plugin-install source-server --dir "<plugin files dir>"
```


### Step 3: Transfer configuration from the source instance to the target instance

 :::Warning

The following process will wipe out the entire configuration of the target instance, and replace it with the configuration of the source instance. This includes repositories and users.

1. Install JFrog CLI on the source instance machine as described [here](#installing-jfrog-cli-on-the-source-instance-machine).

2. Configure the connection details of the source Artifactory instance with your admin credentials by running the following command from the terminal.

```sh
jf c add source-server
```

3. Configure the connection details of the target Artifactory server with your admin credentials by running the following command from the terminal.

```sh
jf c add target-server
```

4. Run the following command to verify that the target URLs of all the remote repositories are accessible from the target.

```sh
jf rt transfer-config source-server target-server --prechecks
```

If the command output shows that a target URL isn't accessible for any of the repositories, you'll need to make the URL accessible before proceeding to transferring the config. You can then rerun the command to ensure that the URLs are accessible.

:::Note

The process of transferring the config will fail if any of the target URLs is not accessible from the target. You can however exclude repositories with target URLs that aren't accessible from being transferred.

5. Transfer the configuration from the source to the target by running the following command.

```sh
jf rt transfer-config source-server target-server
```

This command might take up to two minutes to run.

:::Note

* By default, the command will not transfer the configuration if it finds that the target instance isn't empty. This can happen for example if you ran the transfer-config command before. If you'd like to force the command run anyway, and overwrite the existing configuration on the target, run the command with the `--force` option.

* In case you do not wish to transfer all repositories, you can use the `--include-repos` and `--exclude-repos` command options. Run the following command to see the usage of these options.

```sh
jf rt transfer-config -h
```

6. View the command output in the terminal to verify that there are no errors.

The command output is divided in to the following four phases:

```sh
========== Phase 1/4 - Preparations ==========
========== Phase 2/4 - Export configuration from the source Artifactory ==========
========== Phase 3/4 - Download and modify configuration ==========
========== Phase 4/4 - Import configuration to the target Artifactory ==========
```

7. View the log to verify there are no errors.

The target instance should now be accessible with the admin credentials of the source instance. Log into the target instance UI. The target instance must have the same repositories as the source.

### Step 4: Disabling configuration transfer on the target instance

Once the configuration transfer is successful, you need to disable the configuration transfer on the target instance. This is important both for security reasons and the target server is set to be low on resources while configuration transfer is enabled.

1. Login to MyJFrog

2. Under the Actions menu, choose **Enable Configuration Transfer**.

3. Toggle **Enable Configuration Transfer** to **off** to disable configuration transfer.

Disabling the configuration transfer might take some time.

### Step 5: Push the files from the source to the target instance

1. Install JFrog CLI on any machine that has access to both the source and the target JFrog instances. To do this, follow the steps described [here](#installing-jfrog-cli-on-a-machine-with-network-access-to-the-source-and-target-machines).

2. Run the following command to start pushing the files from all the repositories in source instance to the target instance.

```sh
jf rt transfer-files source-server target-server
```

This command may take a few days to push all the files, depending on your system size and your network speed. While the command is running, It displays the transfer progress visually inside the terminal.

If you're running the command in the background, you use the following command to view the transfer progress.

```sh
jf rt transfer-files --status
```

:::Note

* In case you do not wish to transfer the files from all repositories, or wish to run the transfer in phases, you can use the `--include-repos` and `--exclude-repos` command options. Run the following command to see the usage of these options.

```sh
jf rt transfer-files -h
```

* If the traffic between the source and target instance needs to be routed through an HTTPS proxy, refer to [this](#routing-the-traffic-from-the-source-to-the-target-through-an-https-proxy) section.

* You can stop the transfer process by hitting on CTRL+C if the process is running in the foreground, or by running the following command, if you're running the process in the background.

```sh
jf rt transfer-files --stop
```

The process will continue from the point it stopped when you re-run the command.

While the file transfer is running, monitor the load on your source instance, and if needed, reduce the transfer speed or increase it for better performance. For more information, see the [Controlling the file transfer speed](#controlling-the-file-transfer-speed).

1. A path to an errors summary file will be printed at the end of the run, referring to a generated CSV file. Each line on the summary CSV represents an error log of a file that failed to to be transferred. On subsequent executions of the `jf rt transfer-files`command, JFrog CLI will attempt to transfer these files again.

2. Once the`jf rt transfer-files`command finishes transferring the files, you can run it again to transfer files which were created or modified while the transfer. You can run the command as many times as needed. Subsequent executions of the command will also attempt to transfer files that failed to be transferred during previous executions of the command.

:::Note

Read more about how the transfer files works in [this](#how-does-files-transfer-work) section.

### Step 6: Sync the configuration between the source and the target

You have the option to sync the configuration between the source and target after the files transfer process is complete. You may want to so this if new config entities, such as projects, repositories or users were created or modified on the source, while the files transfer process has been running. To do this, simply repeat steps 1-3 above.

## Transferring projects and repositories from multiple source instances

The **jf rt transfer-config** command transfers all the config entities (users, groups, projects, repositories and more) from the source to the target instance. While doing so, the existing configuration on the target is deleted and replaced with the new configuration from the source. If you'd like to transfer the projects and repositories from multiple source instances to a single target instance, while preserving the existing configuration on the target, follow the below steps.

:::Note

These steps trigger the transfer of the projects and repositories only. Other configuration entities like users are currently not supported.

1. Ensure that you have admin access tokens for both the source and target instances. You'll have to use an admin access token and not an Admin username and password.

2. Install JFrog CLI on any machine that has access to both the source and the target instances using the steps described [here](#installing-jfrog-cli-on-a-machine-with-network-access-to-the-source-and-target-machines). Make sure to use the admin access tokens and not an admin username and password when configuring the connection details of the source and the target.

3. Run the following command to merge all the projects and repositories from the source to the target instance.

```sh
jf rt transfer-config-merge source-server target-server
```

:::Note

In case you do not wish to transfer the files from all projects or the repositories, or wish to run the transfer in phases, you can use the `--include-projects, --exclude-projects, --include-repos` and `--exclude-repos` command options. Run the following command to see the usage of these options.

```sh
jf rt transfer-config-merge -h
```

## How does files transfer work?

### Files transfer phases

The `jf rt transfer-files` command pushes the files from the source instance to the target instance as follows:

* The files are pushed for each repository, one by one in sequence.

* For each repository, the process includes the following three phases:

* **Phase 1** pushes all the files in the repository to the target.

* **Phase 2** pushes files which have been created or modified after phase 1 started running (diffs).

* **Phase 3** attempts to push files which failed to be transferred in earlier phases (**Phase 1** or **Phase 2**) or in previous executions of the command.

* If **Phase 1** finished running for a specific repository, and you run the `jf rt transfer-files` command again, only **Phase 2** and **Phase 3** will be triggered. You can run the `jf rt transfer-files` as many times as needed, till you are ready to move your traffic to the target instance permanently. In any subsequent run of the command, **Phase 2** will transfer the newly created and modified files and **Phase 3** will retry transferring files which failed to be transferred in previous phases and also **in previous runs of the command**.

### Using Replication

To help reduce the time it takes for Phase 2 to run, you may configure Event Based Push Replication for some or all of the local repositories on the source instance. With Replication configured, when files are created or updated on the source repository, they are immediately replicated to the corresponding repository on the target instance.Repository Replication

The replication can be configured at any time. Before, during or after the files transfer process.

### Files transfer state

You can run the `jf rt transfer-files` command multiple times. This is needed to allow transferring files which have been created or updated after previous command executions. To achieve this, JFrog CLI stores the current state of the files transfer process in a directory named `transfer` under the JFrog CLI home directory. You can usually find this directory at this location `~/.jfrog/transfer`.

JFrog CLI uses the state stored in this directory to avoid repeating transfer actions performed in previous executions of the command. For example, once **Phase 1** is completed for a specific repository, subsequent executions of the command will skip **Phase 1** and run **Phase 2** and **Phase 3** only.

In case you'd like to ignore the stored state, and restart the files transfer from scratch, you can add the `--ignore-state` option to the `jf rt transfer-files` command.

## Installing JFrog CLI on a machine with network access to the source and target machines

Unlike the transfer-config command, which should be run from the primary note machines of Artifactory, it is recommended to run the transfer-files command from a machine that has network access to the source Artifactory URL. This allows spreading the transfer load on all the Artifactory cluster nodes. This machine should also have network access to the target Artifactory URL.

Follows these steps to installing JFrog CLI on that machine.

1. Install JFrog CLI by using one of the [JFrog CLI Installers](https://jfrog.com/getcli/). For example:

```sh
curl -fL https://install-cli.jfrog.io | sh
```

2. If your source instance is accessible only through an HTTP/HTTPS proxy, set the proxy environment variable as described [here](https://jfrog-staging-external.fluidtopics.net/r/help/jfrog-cli/proxy-support).

3. Configure the connection details of the source Artifactory instance with your admin credentials.

Run the following command and follow the instructions.

```sh
jf c add source-server
```

4. Configure the connection details of the target Artifactory instance.

```sh
jf c add target-server
```

## Installing JFrog CLI on the source instance machine

Install JFrog CLI on your source instance by using one of the [JFrog CLI Installers](https://jfrog.com/getcli/). For example:

```sh
curl -fL https://install-cli.jfrog.io | sh
```

:::Note

If the source instance is running as a docker container, and you're not able to install JFrog CLI while inside the container, follow these steps.

1. Connect to the host machine through the terminal.

2. Download the JFrog CLI executable into the correct directory by running this command.

```sh
curl -fL https://getcli.jfrog.io/v2-jf | sh
```

3. Copy the JFrog CLI executable you've justdownloadedto the container, by running the following docker command. Make sure to replace`<the container name>`with the name of the container.

```sh
docker cp jf <the container name>:/usr/bin/jf
```

4. Connect to the container and run the following command to ensure JFrog CLI is installed

```sh
jf -v
```

## Controlling the file transfer speed

The `jf rt transfer-files` command pushes the binaries from the source instance to the target instance. This transfer can take days, depending on the size of the total data transferred, the network bandwidth between the source and the target instance, and additional factors.

Since the process is expected to run while the source instance is still being used, monitor the instance to ensure that the transfer does not add too much load to it. Also, you might decide to increase the load for faster transfer while you monitor the transfer. This section describes how to control the file transfer speed.

By default, the `jf rt transfer-files` command uses 8 working threads to push files from the source instance to the target instance. Reducing this value will cause slower transfer speed and lower load on the source instance, and increasing it will do the opposite. We therefore recommend increasing it gradually. This value can be changed while the `jf rt transfer-files` command is running. There's no need to stop the process to change the total of working threads. The new value set will be cached by JFrog CLI and also used for subsequent runs from the same machine. To set the value, simply run the following interactive command from a new terminal window on the same machine which runs the `jf rt transfer-files` command.

```sh
jf rt transfer-settings
```

## Build-info repositories

When transferring files in build-info repositories, JFrog CLI limits the total of working threads to 8. This is done in order to limit the load on the target instance while transferring build-info.

## Manually copying the filestore to reduce the transfer time

When your self-hosted Artifactory hosts hundreds of terabytes of binaries, you may consult with your JFrog account manager about the option of reducing the files transfer time by manually copying the entire filestore to the JFrog Cloud storage. This reduces the transfer time, because the binaries content do not need to be transferred over the network.

The `jf rt transfer-files` command transfers the metadata of the binaries to the database (file paths, file names, properties and statistics). The command also transfers the binaries that have been created and modified after you copy the filestore.

To run the files transfer after you copy the filestore, add the `--filestore` command option to the `jf rt transfer-files` command.

## Running pre-checks before initiating the files transfer process

Before initiating the files transfer process, we highly recommend running pre-checks, to identify issues that can affect the transfer. You trigger the pre-checks by running a JFrog CLI command on your terminal. The pre-checks will verify the following:

1. There's network connectivity between the source and target instances.

2. The source instance does not include artifacts with properties with values longer than 2.4K characters. This is important, because values longer than 2.4K characters are not supported in JFrog Cloud, and those properties are skipped during the transfer process.

To run the pre-checks, follow these steps:

1. Install JFrog CLI on any machine that has access to both the source and the target JFrog instances. To do this, follow the steps described [here](#installing-jfrog-cli-on-a-machine-with-network-access-to-the-source-and-target-machines).

2. Run the following command:

```sh
jf rt transfer-files source-server target-server --prechecks
```

:::Note

If the traffic between the source and target instance needs to be routed through an HTTPS proxy, add the --proxy-key command option as described in [this](#routing-the-traffic-from-the-source-to-the-target-through-an-https-proxy) section.


## Transferring files larger than 25 GB

By default, files that are larger than 25 GB will be blocked by the JFrog Cloud infrastructure during the files transfer. To check whether your source Artifactory instance hosts files larger than that size, do the following.

1. Run the following curl command from your terminal, after replacing the `<source instance URL>`, `<username>` and `<password>` tokens with your source instance details. The command execution may take a few minutes, depending on the number of files hosted by Artifactory.

```sh
curl -X POST <source instance URL>/artifactory/api/search/aql -H "Content-Type: text/plain" -d 'items.find({"name" : {"$match":"\*"}}).include("size").sort({"$desc" : \["size"\]}).limit(1)' -u <username>:<password>
```

2. You should get a result that looks like the following.

```json
{
"results" : \[ {
  "size" : 132359021
} \],
"range" : {
  "start\_pos" : 0,
  "end\_pos" : 1,
  "total" : 1,
  "limit" : 1
}
}
```

The value of_**size**_represents the largest file size hosted by your source Artifactory instance.

3. If the size value you received is larger than 25000000000, please avoid initiating the files transfer before contacting JFrog Support, to check whether this size limit can be increased for you. You can contact Support by sending an email to [support@jfrog.com](mailto:support@jfrog.com)

## Routing the traffic from the source to the target through an HTTPS proxy

The `jf rt transfer-files` command pushes the files directly from the source to the target instance over the network. In case the traffic from the source instance needs to be routed through an HTTPS proxy, follow these steps.

1. Define the proxy details in the source instance UI as described in the Managing ProxiesManaging Proxies documentation.

2. When running the `jf rt transfer-files` command, add the `--proxy-key` option to the command, with Proxy Key you configured in the UI as the option value. For example, if the Proxy Key you configured is **my-proxy-key**, run the command as follows:

jf rt transfer-files my-source my-target --proxy-key my-proxy-key

## Frequently asked questions

**Why is the total files count on my source and target instances different after the files transfer finishes?**

It is expected to see sometimes significant differences between the files count on the source and target instances, after the transfer ends. These differences can be caused by many reasons, and in most cases are not an indication for an issue. For example, Artifactory may include file cleanup policies that are triggered by the files deployment. This can cause some files to be cleaned up from the target repository after they are transferred.

**How can I validate that all files were transferred from the source to the target instance?**

There's actually no need to validate that all files were transferred at the end of the transfer process. JFrog CLI performs this validation for you while the process is running. It does that as follows.

1. JFrog CLI traverses the repositories on the source instance and pushes all files to the target instance.

2. If a file fails to reach the target instance or isn't deployed there successfully, the source instance logs this error with the file details.

3. At the end of the transfer process, JFrog CLI provides you with a summary of all files which failed to be pushed.

4. The failures are also logged inside the `transfer` directory under the JFrog CLI home directory. This directory is usually located at `~/.jfrog/transfer`. Subsequent runs of the jf rt `transfer-files` command use this information for attempting another transfer of the files.

**Does JFrog CLI validate the integrity of files, after they are transferred to the target instance?**

Yes. The source Artifactory instance stores a checksum for every file it hosts. When files are transferred to the target instance, they are transferred with the checksums as HTTP headers. The target instance calculates the checksum for each file it receives, and then compares it to the received checksum. If the checksums don't match, the target reports this to the source, which will attempt to transfer the file again at a later stage of the process.

**Can I stop the jf rt transfer-files command and then start it again? Would that cause any issue?**

You can stop the command at any time by hitting CTRL+C and then run it again. JFrog CLI stores the state of the transfer process in the "transfer" directory under the JFrog CLI home directory. This directory is usually located at ~/.jfrog/transfer. Subsequent executions of the command use the data stored in that directory to try and avoid transferring files that have already been transferred in previous command executions.
