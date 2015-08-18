package command

import (
    "com.jfrog/bintray/cli/client"
    "log"
    "net/http"
)

const defaultApiUrl = "https://bintray.com/api/v1/"

type Command interface {
    Execute(bt client.Bintray) (result interface{}, err error)
}

func updateRequestAuth(req *http.Request, bt *client.Bintray) {
    req.SetBasicAuth(bt.Username, bt.ApiKey)
}

func perror(err error) {
    if err != nil {
        log.Fatal(err)
    }
}