# Jfrog Client

## General
    This section includes a few usage examples of the Jfrog client APIs from your application code.
## Artifactory Client

### Setting up Artifactory details
 ```
    rtDetails := auth.NewArtifactoryDetails()
    rtDetails.SetUrl("http://localhost:8081/artifactory")
    rtDetails.SetSshKeysPath("path/to/.ssh/")
    rtDetails.SetApiKey("apikey")
    rtDetails.SetUser("user")
    rtDetails.SetPassword("password")
 ```

### Setting up Artifactory service manager
```
    serviceConfig, err := artifactory.NewConfigBuilder().
        SetArtDetails(rtDetails).
        SetCertifactesPath(certPath).
        SetMinChecksumDeploy(minChecksumDeploySize).
        SetSplitCount(splitCount).
        SetMinSplitSize(minSplitSize).
        SetThreads(threads).
        SetDryRun(false).
        SetLogger(logger).
        Build()
    // Check for errors

    rtManager, err := artifactory.New(serviceConfig)
    // Check for errors
```

### Services Execution:

#### Upload
```
    commonParams := &utils.ArtifactoryCommonParams{}
    commonParams.Pattern = "*/*.zip"
    commonParams.Target = "repo/path/"
    
    params := &services.UploadParamsImpl{ArtifactoryCommonParams: commonParams}
    rtManager.UploadFiles(params)
```

#### Download
```
    commonParams := &utils.ArtifactoryCommonParams{}
    commonParams.Pattern = "repo/*/*.zip"
    commonParams.Target = "target/path/"
    
    params := &services.DownloadParamsImpl{ArtifactoryCommonParams: commonParams}
    rtManager.DownloadFiles(params)
```

#### Search
```
    commonParams := &utils.ArtifactoryCommonParams{}
    commonParams.Pattern = "repo/*/*.zip"
    
    params := &services.SearchParamsImpl{ArtifactoryCommonParams: commonParams}
    rtManager.DownloadFiles(params)
```

#### Delete
```
    commonParams := &utils.ArtifactoryCommonParams{}
    commonParams.Pattern = "repo/*/*.zip"
    
    params := &services.DeleteParamsImpl{ArtifactoryCommonParams: commonParams}
    rtManager.DownloadFiles(params)
```

#### Get Unreferenced Git Lfs Files
```
    param := &services.GitLfsCleanParamsImpl{}
    param.Refs = "refs/remotes/origin/master"
    params.Repo = "my-project-lfs"
    
    rtManager.DownloadFiles(params)
```

#### Move
```
    commonParams := &utils.ArtifactoryCommonParams{}
    commonParams.Pattern = "repo/*/*.zip"
    commonParams.Target = "repo2/path/"
    
    params := &services.MoveCopyParamsImpl{ArtifactoryCommonParams: commonParams}
    rtManager.Move(params)
```

#### Copy
```
    commonParams := &utils.ArtifactoryCommonParams{}
    commonParams.Pattern = "repo/*/*.zip"
    commonParams.Target = "repo2/path/"
    
    params := &services.MoveCopyParamsImpl{ArtifactoryCommonParams: commonParams}
    rtManager.Copy(params)
```

#### Distribute
```
    params := &services.BuildDistributionParamsImpl{}
    params.SourceRepos = "source-repo"
    params.TargetRepo = "target-repo"
    params.GpgPassphrase = "GpgPassphrase"
    params.Publish = false
    params.OverrideExistingFiles = false
    params.Async = true
    params.BuildName = "buildName"
    params.BuildNumber = "10"
    params.Pattern = "repo/*/*.zip"
    
    rtManager.DistributeBuild(params)
```

#### Promote
```
    params := &services.PromotionParamsImpl{}
    params.BuildName = "buildName"
    params.BuildNumber = "10"
    params.TargetRepo = "target-repo"
    params.Status = "status"
    params.Comment = "comment"
    params.Copy = true
    params.IncludeDependencies = false
    params.SourceRepo = "source-repo"
    
    rtManager.DownloadFiles(params)
```

#### Set Properties
```
    commonParams := &utils.ArtifactoryCommonParams{}
    commonParams.Pattern = "repo/*/*.zip"
    item, err := rtManager.Search(&clientutils.SearchParamsImpl{ArtifactoryCommonParams: commonParams})
    // Check for errors
    var items []clientutils.ResultItem
    items = append(items, item)
    prop := "key=value"
    
    params := &services.SetPropsParamsImpl{Items:items, Props:prop}
    rtManager.SetProps(params)
```

#### Execute AQL
```
    rtManager.Aql(aql string)
```

#### Xray Scan
```
    xrayScanParams := &services.XrayScanParamsImpl{}
    xrayScanParams.BuildName = buildName
    xrayScanParams.BuildNumber = buildNumber
    
    rtManager.XrayScanBuild(params)
```

## Bintray Client

### Setting up Bintray details
 ```
    btDetails := auth.NewBintrayDetails()
    btDetails.SetUser("user")
    btDetails.SetKey("key")
    btDetails.SetDefPackageLicense("Apache 2.0")
 ```

### Setting up Bintray service manager
```
    serviceConfig := bintray.NewConfigBuilder().
        SetBintrayDetails(btDetails).
        SetDryRun(false).
        SetThreads(threads).
        SetMinSplitSize(minSplitSize).
        SetSplitCount(splitCount).
        SetLogger(logger).
        Build()

    btManager, err := bintray.New(serviceConfig)
    // Check for errors
```
### Services Execution:

#### Upload
```
    params := services.NewUploadParams()
    params.Pattern = "*/*.zip"
    params.Path = versions.CreatePath("subject/repo/pkg/version")
    params.TargetPath = "path/to/files"
    params.Deb = "distribution/component/architecture"
    params.Recursive = true
    params.Flat = true
    params.Publish = false
    params.Override = false
    params.Explode = false
    params.UseRegExp = false
    
    btManager.UploadFiles(params)
```

#### Download File
```
    params := services.NewDownloadFileParams()
    params.Flat = false
    params.IncludeUnpublished = false
    params.PathDetails = "path/to/file"  
    params.TargetPath = "target/path/"

    btManager.DownloadFile(params)
```

#### Download Version
```
    params := services.NewDownloadVersionParams()
    params.Path, err = versions.CreatePath("subject/repo/pkg/version")
    // Check for errors
    params.IncludeUnpublished = false
    params.TargetPath = "target/path/"

    btManager.DownloadVersion(params)
```

#### Show/Delete Package
```
    pkgPath, err := packages.CreatePath("subject/repo/pkg")
    // Check for errors

    btManager.ShowPackage(pkgPath)
    btManager.DeletePackage(pkgPath)
```

#### Create/Update Package
```
    params := packages.NewPackageParams()
    params.Path, err = packages.CreatePath("subject/repo/pkg")
    // Check for errors
    params.Desc = "description"
    params.Labels = "labels"
    params.Licenses = "licences"
    params.CustomLicenses = "custum-licenses"
    params.VcsUrl = "https://github.com/JFrogDev/jfrog-cli-go"
    params.WebsiteUrl = "https://jfrog.com"
    params.IssueTrackerUrl = "https://github.com/bintray/bintray-client-java/issues"
    params.GithubRepo = "bintray/bintray-client-java"
    params.GithubReleaseNotesFile = "RELEASE_1.2.3.txt" "github-rel-notes"
    params.PublicDownloadNumbers = "true"
    params.PublicStats = "true"

    btManager.CreatePackage(params)
    btManager.UpdatePackage(params)
```

#### Show/Delete Version
```
    versionPath, err := versions.CreatePath("subject/repo/pkg/version")
    // Check for errors

    btManager.ShowVersion(versionPath)
    btManager.DeleteVersion(versionPath)
```

#### Create/Update Version
```
    params := versions.NewVersionParams()
    params.Path, err = versions.CreatePath("subject/repo/pkg/version")
    // Check for errors
    params.Desc = "description"
    params.VcsTag = "1.1.5"
    params.Released = "true"
    params.GithubReleaseNotesFile = "RELEASE_1.2.3.txt"
    params.GithubUseTagReleaseNotes = "false"

    btManager.CreateVersion(params)
    btManager.UpdateVersion(params)
```

#### Show/Delete Version
```
    path, err := versions.CreatePath("subject/repo/pkg/version")
    // Check for errors

    btManager.ShowVersion(path)
    btManager.DeleteVersion(path)
```

#### Create/Update Entitlements
```
    params := entitlements.NewEntitlementsParams()
    params.VersionPath, err = versions.CreatePath("subject/repo/pkg/version")
    // Check for errors
    params.Path = "a/b/c"
    params.Access = "rw"
    params.Keys = "keys"

    btManager.CreateEntitlement(params)
    
    params.Id = "entitlementID"
    btManager.UpdateEntitlement(params)
```

#### Show/Delete Entitlements
```
    versionPath, err := versions.CreatePath("subject/repo/pkg/version")
    // Check for errors

    btManager.ShowAllEntitlements(versionPath)
    btManager.ShowEntitlement("entitelmentID", versionPath)
    btManager.DeleteEntitlement("entitelmentID", versionPath)
```

#### Create/Update Access Keys
```
    params := accesskeys.NewAccessKeysParams()
    params.Password = "password"
    params.Org = "org"
    params.Expiry = time.Now() + time.Hour * 10
    params.ExistenceCheckUrl = "http://callbacks.myci.org/username=:username,password=:password"
    params.ExistenceCheckCache = 60
    params.WhiteCidrs = "127.0.0.1/22,193.5.0.1/92"
    params.BlackCidrs = "127.0.0.1/22,193.5.0.1/92"
    params.ApiOnly = true

    btManager.CreateAccessKey(params)

    params.Id = "KeyID"
    btManager.UpdateAccessKey(params)
```

#### Show/Delete Access Keys
```
    btManager.ShowAllAccessKeys("org")
    btManager.ShowAccessKey("org", "KeyID")
    btManager.DeleteAccessKey("org", "KeyID")
```

#### Sign URL
```
    params := url.NewURLParams()
    params.PathDetails, err = utils.CreatePathDetails("subject/repository/file-path")
    // Check for errors
    params.Expiry = time.Now() + time.Hour * 10
    params.ValidFor = 60
    params.CallbackId = "callback-id"
    params.CallbackEmail = "callback-email"
    params.CallbackUrl = "callback-url"
    params.CallbackMethod = "callback-method"

    btManager.SignUrl(params)
```

#### GPG Sign File
```
    path, err := utils.CreatePathDetails("subject/repository/file-path")
    // Check for errors

    btManager.GpgSignFile(path, "passphrase") 
```
	
#### GPG Sign Version
```
    path, err := versions.CreatePath("subject/repo/pkg/version")
    // Check for errors

    btManager.GpgSignVersion(path, "passphrase")
```

#### List Logs
```
    path, err := versions.CreatePath("subject/repo/pkg/version")
    // Check for errors

    btManager.LogsList(versionPath)
```

#### Download Logs
```
    path, err := versions.CreatePath("subject/repo/pkg/version")
    // Check for errors

    btManager.DownloadLog(path, "logName")
```

#### Tests
To run tests execute the following command: 
````
go test -v github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services
````
Optional flags:

| Flag | Description |
| --- | --- |
| `-rt.url` | [Default: http://localhost:8081/artifactory] Artifactory URL. |
| `-rt.user` | [Default: admin] Artifactory username. |
| `-rt.password` | [Default: password] Artifactory password. |
| `-rt.apikey` | [Optional] Artifactory API key. |
| `-rt.sshKeyPath` | [Optional] Ssh key file path. Should be used only if the Artifactory URL format is ssh://[domain]:port |
| `-rt.sshPassphrase` | [Optional] Ssh key passphrase. |
| `-log-level` | [Default: INFO] Sets the log level. |


* Running the tests will create the repository: `jfrog-cli-tests-repo1`.<br/>
  Once the tests are completed, the content of this repository will be deleted.