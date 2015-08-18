package client

const defaultApiUrl = "https://bintray.com/api/v1/"

type Bintray struct {
    Username   string
    ApiKey     string
    ApiUrl string
    Flags map[string]string
}

func New(username string, apiKey string, apiUrl string, flags map[string]string) *Bintray {
    if apiUrl == "" { apiUrl = apiUrl }
    if apiUrl == "" {
        apiUrl = defaultApiUrl
    }
    return &Bintray{username, apiKey, apiUrl, flags}
}


/*func (bt *Bintray) SearchFile(name string) []BintrayFile {
    var files []BintrayFile
    response, err := bt.bc.ExecuteGet("search/file?name=" + name)
    if (err != nil) {
        fmt.Printf("%s", err)
    } else {
        defer response.Body.Close()
        contents, err := ioutil.ReadAll(response.Body)
        if err == nil {
            //fmt.Printf("%s\n", string(contents))
            json.Unmarshal(contents, &files)
        } else {
            fmt.Printf("%s", err)
        }
    }
    return files
}

func (bt *Bintray) DownloadAsync(urls []string, targetDir string) []*os.File {
    ch := make(chan *os.File)
    responses := []*os.File{}
    for _, url := range urls {
        go func(url string) {
            fmt.Printf("Downloading %s\n", url)
            i, j := strings.LastIndex(url, "/"), len(url)
            fileName := url[i:j]
            targetFile := targetDir + "/" + fileName
            file := bt.Download(url, targetFile)
            ch <- file
        }(url)
    }

    for {
        select {
        case response := <-ch:
            fmt.Printf("Downloaded %s", response.Name)
            responses = append(responses, response)
            if len(responses) == len(urls) {
                return responses
            }
        case <-time.After(50 * time.Millisecond):
            fmt.Print(".")
        }
    }
}

func (bt *Bintray) Download(url string, target string) *os.File {
    response, err := bt.bc.ExecuteGet(url)
    defer response.Body.Close()
    if err != nil {
        fmt.Printf("Failed to execute download: %s", err)
        return nil
    }
    out, err := os.Create(target)
    defer out.Close()
    if err != nil {
        fmt.Printf("Failed to create file %s", err)
        return nil
    }
    _, err = io.Copy(out, response.Body)
    if err != nil {
        fmt.Printf("Failed to download: %s", err)
        return nil
    }

    return out
}

func (bt *Bintray) GetRepositories(subject string) []Repository {
    var repos []Repository
    response, err := bt.bc.ExecuteGet("repos/" + subject)
    if err == nil {
        defer response.Body.Close()
        contents, err := ioutil.ReadAll(response.Body)
        if err == nil {
            fmt.Printf("%s\n", string(contents))
            json.Unmarshal(contents, &repos)
        } else {
            fmt.Printf("%s", err)
        }
    } else {
        fmt.Printf("%s", err)
    }

    return repos
}

func (bt *Bintray) GetRepository(subject string, repo string) Repository {
    var repoResult Repository
    response, err := bt.bc.ExecuteGet("repos/" + subject + "/" + repo)
    if err == nil {
        defer response.Body.Close()
        contents, err := ioutil.ReadAll(response.Body)
        if err == nil {
            fmt.Printf("%s\n", string(contents))
            json.Unmarshal(contents, &repoResult)
        } else {
            fmt.Printf("%s", err)
        }
    } else {
        fmt.Printf("%s", err)
    }

    return repoResult
}

func (bt *Bintray) RepositorySearch(name string, desc string) []Repository {
    var repos []Repository
    response, err := bt.bc.ExecuteGet("search/repos?name=" + name + "&desc=" + desc)
    if err == nil {
        defer response.Body.Close()
        contents, err := ioutil.ReadAll(response.Body)
        if err == nil {
            fmt.Printf("%s\n", string(contents))
            json.Unmarshal(contents, &repos)
        } else {
            fmt.Printf("%s", err)
        }
    } else {
        fmt.Printf("%s", err)
    }

    return repos
}

func (bt *Bintray) GetPackage(subject string, repo string, packageName string) Pkg {
    var pkgResult Pkg
    response, err := bt.bc.ExecuteGet("packages/" + subject + "/" + repo + "/" + packageName)
    if err == nil {
        defer response.Body.Close()
        contents, err := ioutil.ReadAll(response.Body)
        if err == nil {
            fmt.Printf("%s\n", string(contents))
            json.Unmarshal(contents, &pkgResult)
        } else {
            fmt.Printf("%s", err)
        }
    } else {
        fmt.Printf("%s", err)
    }

    return pkgResult
}

func (bt *Bintray) PackageSearch(name string, desc string, subject string, repo string) []Pkg {
    var packages []Pkg
    response, err := bt.bc.ExecuteGet("search/packages/?name=" + name + "&desc=" + desc + "&subject=" + subject + "&repo=" + repo)
    if err == nil {
        defer response.Body.Close()
        contents, err := ioutil.ReadAll(response.Body)
        if err == nil {
            fmt.Printf("%s\n", string(contents))
            json.Unmarshal(contents, &packages)
        } else {
            fmt.Printf("%s", err)
        }
    } else {
        fmt.Printf("%s", err)
    }

    return packages
}*/
