package utils

import (
	"os"
	"bytes"
	"net/http"
	"io/ioutil"
)

func DownloadFile(url, user, password string) *http.Response {
    fileName := GetFileNameFromUrl(url)
    out, err := os.Create(fileName)
    CheckError(err)
    defer out.Close()
    resp, body := SendGet(url, nil, user, password)
    out.Write(body)
    CheckError(err)
    return resp
}

func SendGet(url string, headers map[string]string, user, password string) (*http.Response, []byte) {
    return Send("GET", url, nil, headers, user, password)
}

func SendPost(url string, content []byte, user string, password string) (*http.Response, []byte) {
    return Send("POST", url, content, nil, user, password)
}

func Send(method string, url string, content []byte, headers map[string]string, user, password string) (*http.Response, []byte) {
    var req *http.Request
    var err error

    if content != nil {
        req, err = http.NewRequest(method, url, bytes.NewBuffer(content))
    } else {
        req, err = http.NewRequest(method, url, nil)
    }
    CheckError(err)
    req.Close = true
    if user != "" && password != "" {
	    req.SetBasicAuth(user, password)
    }
    if headers != nil {
        for name := range headers {
            req.Header.Set(name, headers[name])
        }
    }
    client := &http.Client{}
    resp, err := client.Do(req)
    CheckError(err)
    defer resp.Body.Close()
    body, _ := ioutil.ReadAll(resp.Body)
    return resp, body
}