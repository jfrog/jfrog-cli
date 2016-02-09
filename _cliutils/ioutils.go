package cliutils

import (
    "io"
 	"os"
 	"bytes"
 	"bufio"
 	"os/user"
 	"strings"
 	"runtime"
 	"strconv"
 	"net/http"
 	"io/ioutil"
    "path/filepath"
 )

var tempDirPath string

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

func SendGet(url string, headers map[string]string, user, password string) (*http.Response, []byte) {
    return Send("GET", url, nil, headers, user, password)
}

func SendPost(url string, headers map[string]string, content []byte, user string, password string) (*http.Response, []byte) {
    return Send("POST", url, content, headers, user, password)
}

func SendPatch(url string, content []byte, user string, password string) (*http.Response, []byte) {
    return Send("PATCH", url, content, nil, user, password)
}

func SendDelete(url string, user, password string) (*http.Response, []byte) {
    return Send("DELETE", url, nil, nil, user, password)
}

func SendHead(url string, user, password string) *http.Response {
    resp, _ := Send("HEAD", url, nil, nil, user, password)
    return resp
}

func SendPut(url string, content []byte, headers map[string]string, user, password string) (*http.Response, []byte) {
    return Send("PUT", url, content, headers, user, password)
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

func UploadFile(f *os.File, url, user, password string, headers map[string]string) *http.Response {
    fileInfo, err := f.Stat()
    CheckError(err)
    size := fileInfo.Size()

    req, err := http.NewRequest("PUT", url, f)
    CheckError(err)
    req.ContentLength = size
    req.Close = true

    if headers != nil {
        for name := range headers {
            req.Header.Set(name, headers[name])
        }
    }
    if user != "" && password != "" {
	    req.SetBasicAuth(user, password)
    }
    addUserAgentHeader(req)

    length := strconv.FormatInt(size, 10)
    req.Header.Set("Content-Length", length)

    client := &http.Client{}
    resp, err := client.Do(req)
    CheckError(err)
    defer resp.Body.Close()
    return resp
}

func DownloadFile(downloadPath, localPath, fileName string, flat bool,
    user, password string) *http.Response {
    if !flat && localPath != "" {
        os.MkdirAll(localPath ,0777)
        fileName = localPath + "/" + fileName
    }

    out, err := os.Create(fileName)
    CheckError(err)
    defer out.Close()
    resp, body := SendGet(downloadPath, nil, user, password)
    out.Write(body)
    CheckError(err)
    return resp
}

func GetTempDirPath() string {
    if tempDirPath == "" {
        Exit(ExitCodeError, "Function cannot be used before 'tempDirPath' is created.")
    }
    return tempDirPath
}

func CreateTempDirPath() {
    if tempDirPath != "" {
        Exit(ExitCodeError, "'tempDirPath' has already been initialized.")
    }
    path, err := ioutil.TempDir("", "artifactory.cli.")
    CheckError(err)
    tempDirPath = path
}

func RemoveTempDir() {
    if IsDirExists(tempDirPath) {
        os.RemoveAll(tempDirPath)
    }
}

func IsDirExists(path string) bool {
    if !IsPathExists(path) {
        return false
    }
    f, err := os.Stat(path)
    CheckError(err)
    return f.IsDir()
}

// Reads the content of the file in the source path and appends it to
// the file in the destination path.
func AppendFile(srcPath string, destFile *os.File) {
    srcFile, err := os.Open(srcPath)
    CheckError(err)

    defer func() {
        err := srcFile.Close();
        CheckError(err)
    }()

    reader := bufio.NewReader(srcFile)

    writer := bufio.NewWriter(destFile)
    buf := make([]byte, 1024000)
    for {
        n, err := reader.Read(buf)
        if err != io.EOF {
            CheckError(err)
        }
        if n == 0 {
            break
        }
        _, err = writer.Write(buf[:n])
        CheckError(err)
    }
    err = writer.Flush()
    CheckError(err)
}

func GetFileNameFromUrl(url string) string {
    parts := strings.Split(url, "/")
    size := len(parts)
    if size == 0 {
        return url
    }
    return parts[size-1]
}

func GetHomeDir() string {
    user, err := user.Current()
    if err == nil {
        return user.HomeDir
    }
    home := os.Getenv("HOME")
    if home != "" {
        return home
    }
    return "";
}

func ReadFile(filePath string) []byte {
	content, err := ioutil.ReadFile(filePath)
	CheckError(err)
	return content
}
func addUserAgentHeader(req *http.Request) {
    req.Header.Set("User-Agent", "jfrog-cli-go/" + GetVersion())
}