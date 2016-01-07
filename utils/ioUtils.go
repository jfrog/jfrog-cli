package utils

import (
 	"os"
 	"bytes"
 	"strings"
 	"runtime"
 	"strconv"
 	"net/http"
 	"io/ioutil"
    "path/filepath"
 )

func GetFileSeperator() string {
    if runtime.GOOS == "windows" {
        return "\\\\"
    }
    return "/"
}

func IsPathExists(path string) bool {
 _, err := os.Stat(path)
 return !os.IsNotExist(err)
}

func IsFileExists(path string) bool {
    if !IsPathExists(path) {
        return false
    }
    f, err := os.Stat(path)
    CheckError(err)
    return !f.IsDir()
}

func IsDir(path string) bool {
    if !IsPathExists(path) {
        return false
    }
    f, err := os.Stat(path)
    CheckError(err)
    return f.IsDir()
}

func GetFileNameFromPath(path string) string {
    index := strings.LastIndex(path, "/")
    if index != -1 {
        return path[index+1:]
    }
    index = strings.LastIndex(path, "\\")
    if index != -1 {
        return path[index+1:]
    }
    return path
}

// Return the recursive list of files and directories in the specified path
func ListFilesRecursive(path string) []string {
    fileList := []string{}
    err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
        fileList = append(fileList, path)
        return nil
    })
    CheckError(err)
    return fileList
}

// Return the list of files and directories in the specified path
func ListFiles(path string) []string {
    sep := GetFileSeperator()
    if !strings.HasSuffix(path, sep) {
        path += sep
    }
    fileList := []string{}
    files, _ := ioutil.ReadDir(path)

    for _, f := range files {
        filePath := path + f.Name()
        if IsFileExists(filePath) {
            fileList = append(fileList, filePath)
        }
    }
    return fileList
}

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

func SendPatch(url string, content []byte, user string, password string) (*http.Response, []byte) {
    return Send("PATCH", url, content, nil, user, password)
}

func SendDelete(url string, user string, password string) (*http.Response, []byte) {
    return Send("DELETE", url, nil, nil, user, password)
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

func UploadFile(filePath, url, user, password string) *http.Response {
    file, err := os.Open(filePath)
    CheckError(err)
    defer file.Close()

    fileInfo, err := file.Stat()
    CheckError(err)
    fileSize := fileInfo.Size()

    req, err := http.NewRequest("PUT", url, file)
    CheckError(err)
    req.ContentLength = fileSize
    req.Close = true
    if user != "" && password != "" {
	    req.SetBasicAuth(user, password)
    }
    size := strconv.FormatInt(fileSize, 10)
    req.Header.Set("Content-Length", size)

    client := &http.Client{}
    resp, err := client.Do(req)
    CheckError(err)
    defer resp.Body.Close()
    return resp
}