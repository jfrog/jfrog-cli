package command

import (
    "time"
    "encoding/json"
    "io/ioutil"
    "net/http"
    "github.com/JFrogDev/bintray-cli-go/client"
)

type GetRepos struct {
}

func (cmd GetRepos) Execute(bt *client.Bintray) (result interface{}, err error) {
    res, err := http.Get(bt.ApiUrl + "repos/" + bt.Flags["subject"])
    perror(err)
    defer res.Body.Close()
    body, err := ioutil.ReadAll(res.Body)
    perror(err)
    var repos []Repository
    err = json.Unmarshal(body, &repos)
    return repos, err
}

type Repository struct {
    Name               string     `json:"name"`
    Owner              string     `json:"owner"`
    Type               string     `json:"type"`
    Private            bool       `json:"private"`
    Premium            string     `json:"premium"`
    Description        string     `json:"desc"`
    Labels             []string   `json:"labels"`
    Created            time.Time  `json:"created"`
    PackageCount       int        `json:"package_count`
}
