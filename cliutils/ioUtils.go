package cliutils

import (
    "io"
 	"os"
 	"fmt"
    "sync"
 	"bytes"
 	"bufio"
 	"os/user"
 	"strings"
 	"runtime"
 	"strconv"
 	"net/http"
 	"io/ioutil"
    "crypto/md5"
    "crypto/sha1"
    "encoding/hex"
    "path/filepath"
 )

var tempDirPath string

func GetFileSeperator() string {
    if runtime.GOOS == "windows" {
        return "\\"
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

func GetFileAndDirFromPath(path string) (fileName, dir string) {
    index := strings.LastIndex(path, "/")
    if index == -1 {
        index = strings.LastIndex(path, "\\")
    }
    if index != -1 {
        fileName = path[index+1:]
        dir = path[:index]
        return
    }
    fileName = path
    dir = ""
    return
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
    addUserAgentHeader(req)
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

func DownloadFileConcurrently(flags ConcurrentDownloadFlags, logMsgPrefix string) {
    tempLoclPath := GetTempDirPath() + "/" + flags.LocalPath

    var wg sync.WaitGroup
    chunkSize := flags.FileSize / int64(flags.SplitCount)
    mod := flags.FileSize % int64(flags.SplitCount)

    for i := 0; i < flags.SplitCount ; i++ {
        wg.Add(1)
        start := chunkSize * int64(i)
        end := chunkSize * (int64(i) + 1)
        if i == flags.SplitCount-1 {
            end += mod
        }
        go func(start, end int64, i int) {
            headers := make(map[string]string)
            headers["Range"] = "bytes=" + strconv.FormatInt(start, 10) +"-" + strconv.FormatInt(end-1, 10)
            resp, body := SendGet(flags.DownloadPath, headers, flags.User, flags.Password)

            fmt.Println(logMsgPrefix + " [" + strconv.Itoa(i) + "]:", resp.Status + "...")

            os.MkdirAll(tempLoclPath ,0777)
            filePath := tempLoclPath + "/" + flags.FileName + "_" + strconv.Itoa(i)

            createFileWithContent(filePath, body)
            wg.Done()
        }(start, end, i)
    }
    wg.Wait()

    if !flags.Flat && flags.LocalPath != "" {
        os.MkdirAll(flags.LocalPath ,0777)
        flags.FileName = flags.LocalPath + "/" + flags.FileName
    }

    if IsPathExists(flags.FileName) {
        err := os.Remove(flags.FileName)
        CheckError(err)
    }

    destFile, err := os.Create(flags.FileName)
    CheckError(err)
    defer destFile.Close()
    for i := 0; i < flags.SplitCount; i++ {
        tempFilePath := GetTempDirPath() + "/" + flags.FileName + "_" + strconv.Itoa(i)
        AppendFile(tempFilePath, destFile)
    }
    fmt.Println(logMsgPrefix + " Done downloading.")
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

func ScanFromConsole(scanInto *string, defaultValue string) {
    fmt.Scanln(scanInto)
    if *scanInto == "" {
        *scanInto = defaultValue
    }
}

func GetFileDetails(filePath string) *FileDetails {
    details := new(FileDetails)
    details.Md5 = calcMd5(filePath)
    details.Sha1 = calcSha1(filePath)

    file, err := os.Open(filePath)
    CheckError(err)
    defer file.Close()

    fileInfo, err := file.Stat()
    CheckError(err)
    details.Size = fileInfo.Size()

    return details
}

func GetRemoteFileDetails(downloadUrl, user, password string) *FileDetails {
    resp := SendHead(downloadUrl, user, password)
    fileSize, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
    CheckError(err)

    fileDetails := new(FileDetails)

    fileDetails.Md5 = resp.Header.Get("X-Checksum-Md5")
    fileDetails.Sha1 = resp.Header.Get("X-Checksum-Sha1")
    fileDetails.Size = fileSize
    fileDetails.AcceptRanges = resp.Header.Get("Accept-Ranges") == "bytes"
    return fileDetails
}

func createFileWithContent(filePath string, content []byte) {
    out, err := os.Create(filePath)
    CheckError(err)
    defer out.Close()
    out.Write(content)
}

func addUserAgentHeader(req *http.Request) {
    req.Header.Set("User-Agent", "jfrog-cli-go/" + GetVersion())
}

func calcSha1(filePath string) string {
    file, err := os.Open(filePath)
    CheckError(err)
    defer file.Close()

    var resSha1 []byte
    hashSha1 := sha1.New()
    _, err = io.Copy(hashSha1, file)
    CheckError(err)
    return hex.EncodeToString(hashSha1.Sum(resSha1))
}

func calcMd5(filePath string) string {
    file, err := os.Open(filePath)
    CheckError(err)
    defer file.Close()

    var resMd5 []byte
    hashMd5 := md5.New()
    _, err = io.Copy(hashMd5, file)
    CheckError(err)
    return hex.EncodeToString(hashMd5.Sum(resMd5))
}

type ConcurrentDownloadFlags struct {
     DownloadPath string
     FileName string
     LocalPath string
     FileSize int64
     SplitCount int
     Flat bool
     User string
     Password string
 }

 type FileDetails struct {
     Md5 string
     Sha1 string
     Size int64
     AcceptRanges bool
 }