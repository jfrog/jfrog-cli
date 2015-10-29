package command

import (
    "github.com/JFrogDev/bintray-cli-go/client"
    "log"
    "net/http"
)

const defaultApiUrl = "https://bintray.com/api/v1/"

type Command interface {
    Execute(bt *client.Bintray, args interface {}) (result interface{}, err error)
}

func updateRequestAuth(req *http.Request, bt *client.Bintray) {
    req.SetBasicAuth(bt.Username, bt.ApiKey)
}

func perror(err error) {
    if err != nil {
        log.Fatal(err)
    }
}