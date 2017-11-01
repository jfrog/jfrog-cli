# Jfrog Client

## General
    This section includes a few usage examples of the Jfrog client APIs from your application code.

### Setting up Artifactory details
 ```
    rtDetails := new(auth.ArtifactoryDetails)
    rtDetails.Url = artifactoryDetails.Url
    rtDetails.SshKeysPath = artifactoryDetails.SshKeyPath
    rtDetails.ApiKey = artifactoryDetails.ApiKey
    rtDetails.User = artifactoryDetails.User
    rtDetails.Password = artifactoryDetails.Password
 ```

### Setting up Artifactory service manager
```
    serviceConfig, _ := (&artifactory.ArtifactoryServicesConfigBuilder{}).
    SetArtDetails(rtDetails).
    SetCertifactesPath(certPath).
    SetMinChecksumDeploy(minChecksumDeploySize).
    SetSplitCount(splitCount).
    SetMinSplitSize(minSplitSize).
    SetNumOfThreadPerOperation(threads).
    SetDryRun(false).
    SetLogger(logger).
    Build()

    rtManager := artifactory.NewArtifactoryService(serviceConfig)
```

### Services Execution:

#### Upload
```
    params := new(utils.ArtifactoryCommonParams)
    params.Pattern = "filePattern"
    params.Target = "UploadTarget"
    uploadParam := &services.UploadParamsImp{}
    uploadParam.ArtifactoryCommonParams = params
    rtManager.UploadFiles(uploadParamImp)
```

#### Download
```
    params := new(utils.ArtifactoryCommonParams)
    params.Pattern = "filePattern"
    params.Target = "DownloadTarget"
    downloadParams := &services.DownloadParamsImp{}
    downloadParams.ArtifactoryCommonParams = params
    rtManager.DownloadFiles(downloadParamsImpl)
```

#### Search
```
    params := new(utils.ArtifactoryCommonParams)
    params.Pattern = "filePattern"
    params.Target = "DownloadTarget"
    searchParams := &services.SearchParamsImpl{}
    searchParams.ArtifactoryCommonParams = params
    rtManager.DownloadFiles(searchParamsImpl)
```

#### Delete
```
    params := new(utils.ArtifactoryCommonParams)
    params.Pattern = "filePattern"
    params.Target = "DownloadTarget"
    deleteParams := &services.DeleteParamsImpl{}
    deleteParams.ArtifactoryCommonParams = params
    rtManager.DownloadFiles(deleteParamsImpl)
```

#### Get Unreferenced Git Lfs Files
```
    gitLfsCleanParams := &services.GitLfsCleanParamsImpl{}
    gitLfsCleanParams.Refs = "refs/remotes/origin/master"
    gitLfsCleanParams.Repo = "my-project-lfs"
    rtManager.DownloadFiles(gitLfsCleanParamsImpl)
```

#### Move
```
    params := new(utils.ArtifactoryCommonParams)
    params.Pattern = "filePattern"
    params.Target = "TargetPath"
    moveCopyParams := &services.MoveCopyParamsImpl{}
    moveCopyParams.ArtifactoryCommonParams = params
    rtManager.Move(moveCopyParamsImpl)
```

#### Copy
```
    params := new(utils.ArtifactoryCommonParams)
    params.Pattern = "filePattern"
    params.Target = "TargetPath"
    moveCopyParams := &services.MoveCopyParamsImpl{}
    moveCopyParams.ArtifactoryCommonParams = params
    rtManager.Copy(moveCopyParamsImpl)
```

#### Distribute
```
    distributionParams := new(services.BuildDistributionParamsImpl)
    distributionParams.SourceRepos = "sourceRepo"
    distributionParams.TargetRepo = "targetRepo"
    distributionParams.GpgPassphrase = "GpgPassphrase"
    distributionParams.Publish = false
    distributionParams.OverrideExistingFiles = false
    distributionParams.Async = true
    distributionParams.BuildName = "buildName"
    distributionParams.BuildNumber = "10"
    distributionParams.Pattern = "filePattern"
    distributionParams.Target = "DownloadTarget"
    rtManager.DistributeBuild(distributionParams)
```

#### Promote
```
    promotionParams := new(services.PromotionParamsImpl)
    promotionParams.BuildName = "buildName"
    promotionParams.BuildNumber = "10"
    promotionParams.TargetRepo "targetRepo"
    promotionParams.Status = "status"
    promotionParams.Comment ="comment"
    promotionParams.Copy = true
    promotionParams.IncludeDependencies = false
    promotionParams.SourceRepo ="sourceRepo"
    rtManager.DownloadFiles(downloadParamsImpl)
```

#### Set Properties
```
    params := new(utils.ArtifactoryCommonParams)
    params.Pattern = "filePattern"
    params.Target = "TargetPath"
    item, err := rtManager.Search(&clientutils.SearchParamsImpl{ArtifactoryCommonParams: params})
    var items []clientutils.ResultItem
    items = append(items, item)
    prop := "key=value"
    setPropsParams := &services.SetPropsParamsImpl{Items:items, Props:prop}
    rtManager.SetProps(setPropsParams)
```

#### Xray Scan
```
    xrayScanParams := new(services.XrayScanParamsImpl)
    xrayScanParams.BuildName = buildName
    xrayScanParams.BuildNumber = buildNumber
    rtManager.XrayScanBuild(params)
```