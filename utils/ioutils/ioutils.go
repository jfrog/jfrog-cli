package ioutils

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/Godeps/_workspace/src/golang.org/x/crypto/ssh/terminal"
	"syscall"
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
	cliutils.CheckError(err)
	return !f.IsDir()
}

func IsDir(path string) bool {
	if !IsPathExists(path) {
		return false
	}
	f, err := os.Stat(path)
	cliutils.CheckError(err)
	return f.IsDir()
}

func GetFileAndDirFromPath(path string) (fileName, dir string) {
	index1 := strings.LastIndex(path, "/")
	index2 := strings.LastIndex(path, "\\")
	var index int
	if index1 >= index2 {
        index = index1
	} else {
	    index = index2
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
	cliutils.CheckError(err)
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
	path = strings.TrimPrefix(path, "." + sep)

	for _, f := range files {
		filePath := path + f.Name()
		if IsFileExists(filePath) {
			fileList = append(fileList, filePath)
		}
	}
	return fileList
}

func sendGetForFileDownload(url string, allowRedirect bool, httpClientsDetails HttpClientDetails) (*http.Response, string, error) {
	resp, _, redirectUrl, err := Send("GET", url, nil, allowRedirect, false, httpClientsDetails)
	return resp, redirectUrl, err
}
func SendGet(url string, allowRedirect bool, httpClientsDetails HttpClientDetails) (*http.Response, []byte, string, error) {
	return Send("GET", url, nil, allowRedirect, true, httpClientsDetails)
}

func SendPost(url string, content []byte, httpClientsDetails HttpClientDetails) (*http.Response, []byte) {
	resp, body, _, err := Send("POST", url, content, true, true, httpClientsDetails)
	cliutils.CheckError(err)
	return resp, body
}

func SendPatch(url string, content []byte, httpClientsDetails HttpClientDetails) (*http.Response, []byte) {
	resp, body, _, err := Send("PATCH", url, content, true, true, httpClientsDetails)
	cliutils.CheckError(err)
	return resp, body
}

func SendDelete(url string, content []byte, httpClientsDetails HttpClientDetails) (*http.Response, []byte) {
	resp, body, _, err := Send("DELETE", url, content, true, true, httpClientsDetails)
	cliutils.CheckError(err)
	return resp, body
}

func SendHead(url string, httpClientsDetails HttpClientDetails) *http.Response {
	resp, _, _, err := Send("HEAD", url, nil, true, true, httpClientsDetails)
	cliutils.CheckError(err)
	return resp
}

func SendPut(url string, content []byte, httpClientsDetails HttpClientDetails) (*http.Response, []byte) {
	resp, body, _, err := Send("PUT", url, content, true, true, httpClientsDetails)
	cliutils.CheckError(err)
	return resp, body
}

func getHttpClient(transport *http.Transport) *http.Client {
	client := &http.Client{}
	if transport != nil{
		client.Transport = transport
	}
	return client
}

func Send(method string, url string, content []byte, allowRedirect bool,
	closeBody bool, httpClientsDetails HttpClientDetails) (resp *http.Response, respBody []byte, redirectUrl string, err error) {

	var req *http.Request
	if content != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(content))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	cliutils.CheckError(err)
	req.Close = true
	if httpClientsDetails.User != "" && httpClientsDetails.Password != "" {
		req.SetBasicAuth(httpClientsDetails.User, httpClientsDetails.Password)
	}
	addUserAgentHeader(req)
	if httpClientsDetails.Headers != nil {
		for name := range httpClientsDetails.Headers {
			req.Header.Set(name, httpClientsDetails.Headers[name])
		}
	}

	client := getHttpClient(httpClientsDetails.Transport);
	if !allowRedirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			redirectUrl = req.URL.String()
			return errors.New("redirect")
		}
	}
	resp, err = client.Do(req)
	if !allowRedirect && err != nil {
		return
	}

	cliutils.CheckError(err)
	if closeBody {
	    defer resp.Body.Close()
	    respBody, _ = ioutil.ReadAll(resp.Body)
	}
	return
}

func UploadFile(f *os.File, url string, httpClientsDetails HttpClientDetails) *http.Response {
	fileInfo, err := f.Stat()
	cliutils.CheckError(err)
	size := fileInfo.Size()

	req, err := http.NewRequest("PUT", url, f)
	cliutils.CheckError(err)
	req.ContentLength = size
	req.Close = true

	if httpClientsDetails.Headers != nil {
		for name := range httpClientsDetails.Headers {
			req.Header.Set(name, httpClientsDetails.Headers[name])
		}
	}
	if httpClientsDetails.User != "" && httpClientsDetails.Password != "" {
		req.SetBasicAuth(httpClientsDetails.User, httpClientsDetails.Password)
	}
	addUserAgentHeader(req)

	length := strconv.FormatInt(size, 10)
	req.Header.Set("Content-Length", length)

	client := getHttpClient(httpClientsDetails.Transport);
	resp, err := client.Do(req)
	cliutils.CheckError(err)
	defer resp.Body.Close()
	return resp
}

func DownloadFile(downloadPath, localPath, fileName string, flat bool, httpClientsDetails HttpClientDetails) *http.Response {
	resp, _, _ := downloadFile(downloadPath, localPath, fileName, flat, true, httpClientsDetails)
	return resp
}

func DownloadFileNoRedirect(downloadPath, localPath, fileName string, flat bool,
	httpClientsDetails HttpClientDetails) (*http.Response, string, error) {

    return downloadFile(downloadPath, localPath, fileName, flat, false, httpClientsDetails)
}

func downloadFile(downloadPath, localPath, fileName string, flat, allowRedirect bool,
	httpClientsDetails HttpClientDetails) (resp *http.Response, redirectUrl string, err error) {
	if !flat && localPath != "" {
		os.MkdirAll(localPath, 0777)
		fileName = localPath + "/" + fileName
	}

	out, err := os.Create(fileName)
	cliutils.CheckError(err)
	defer out.Close()
	resp, redirectUrl, err = sendGetForFileDownload(downloadPath, allowRedirect, httpClientsDetails)
    defer resp.Body.Close()
	if err == nil {
        _, err = io.Copy(out, resp.Body)
	cliutils.CheckError(err)
	}
	return
}

func DownloadFileConcurrently(flags ConcurrentDownloadFlags, logMsgPrefix string, httpClientsDetails HttpClientDetails) {
	var wg sync.WaitGroup
	chunkSize := flags.FileSize / int64(flags.SplitCount)
	mod := flags.FileSize % int64(flags.SplitCount)
	for i := 0; i < flags.SplitCount; i++ {
		wg.Add(1)
		start := chunkSize * int64(i)
		end := chunkSize * (int64(i) + 1)
		if i == flags.SplitCount-1 {
			end += mod
		}
		requestClientDetails := httpClientsDetails.Clone()
		go func(start, end int64, i int) {
			downloadFileRange(flags, start, end, i, logMsgPrefix, *requestClientDetails)
			wg.Done()
		}(start, end, i)
	}
	wg.Wait()

	if !flags.Flat && flags.LocalPath != "" {
		os.MkdirAll(flags.LocalPath, 0777)
		flags.FileName = flags.LocalPath + "/" + flags.FileName
	}

	if IsPathExists(flags.FileName) {
		err := os.Remove(flags.FileName)
		cliutils.CheckError(err)
	}

	destFile, err := os.Create(flags.FileName)
	cliutils.CheckError(err)
	defer destFile.Close()
	for i := 0; i < flags.SplitCount; i++ {
		tempFilePath := GetTempDirPath() + "/" + flags.FileName + "_" + strconv.Itoa(i)
		AppendFile(tempFilePath, destFile)
	}
	fmt.Println(logMsgPrefix + "Done downloading.")
}

func downloadFileRange(flags ConcurrentDownloadFlags, start, end int64, currentSplit int, logMsgPrefix string, httpClientsDetails HttpClientDetails) {
	tempLoclPath := GetTempDirPath()
    if !flags.Flat {
        tempLoclPath += "/" + flags.LocalPath
    }
	if httpClientsDetails.Headers == nil {
		httpClientsDetails.Headers = make(map[string]string)
	}
	httpClientsDetails.Headers["Range"] = "bytes=" + strconv.FormatInt(start, 10) + "-" + strconv.FormatInt(end-1, 10)

	resp, _, err :=
		sendGetForFileDownload(flags.DownloadPath, false, httpClientsDetails)
    defer resp.Body.Close()
	cliutils.CheckError(err)

	fmt.Println(logMsgPrefix + "[" + strconv.Itoa(currentSplit)+"]:", resp.Status+"...")
	os.MkdirAll(tempLoclPath, 0777)
	filePath := tempLoclPath + "/" + flags.FileName + "_" + strconv.Itoa(currentSplit)

	out, err := os.Create(filePath)
	cliutils.CheckError(err)
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	cliutils.CheckError(err)
}

func GetTempDirPath() string {
	if tempDirPath == "" {
		cliutils.Exit(cliutils.ExitCodeError, "Function cannot be used before 'tempDirPath' is created.")
	}
	return tempDirPath
}

func CreateTempDirPath() {
	if tempDirPath != "" {
		cliutils.Exit(cliutils.ExitCodeError, "'tempDirPath' has already been initialized.")
	}
	path, err := ioutil.TempDir("", "jfrog.cli.")
	cliutils.CheckError(err)
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
	cliutils.CheckError(err)
	return f.IsDir()
}

// Reads the content of the file in the source path and appends it to
// the file in the destination path.
func AppendFile(srcPath string, destFile *os.File) {
	srcFile, err := os.Open(srcPath)
	cliutils.CheckError(err)

	defer func() {
		err := srcFile.Close()
		cliutils.CheckError(err)
	}()

	reader := bufio.NewReader(srcFile)

	writer := bufio.NewWriter(destFile)
	buf := make([]byte, 1024000)
	for {
		n, err := reader.Read(buf)
		if err != io.EOF {
			cliutils.CheckError(err)
		}
		if n == 0 {
			break
		}
		_, err = writer.Write(buf[:n])
		cliutils.CheckError(err)
	}
	err = writer.Flush()
	cliutils.CheckError(err)
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
	return ""
}

func ReadFile(filePath string) []byte {
	content, err := ioutil.ReadFile(filePath)
	cliutils.CheckError(err)
	return content
}

func ScanFromConsole(caption string, scanInto *string, defaultValue string) {
    if defaultValue != "" {
        print(caption + " [" + defaultValue + "]: ")
    } else {
        print(caption + ": ")
    }
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
	cliutils.CheckError(err)
	defer file.Close()

	fileInfo, err := file.Stat()
	cliutils.CheckError(err)
	details.Size = fileInfo.Size()

	return details
}

func GetRemoteFileDetails(downloadUrl string, httpClientsDetails HttpClientDetails) *FileDetails {
	resp := SendHead(downloadUrl, httpClientsDetails)
	fileSize, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	cliutils.CheckError(err)

	fileDetails := new(FileDetails)
	fileDetails.Md5 = resp.Header.Get("X-Checksum-Md5")
	fileDetails.Sha1 = resp.Header.Get("X-Checksum-Sha1")
	fileDetails.Size = fileSize
	fileDetails.AcceptRanges = resp.Header.Get("Accept-Ranges") == "bytes"
	return fileDetails
}

func addUserAgentHeader(req *http.Request) {
	req.Header.Set("User-Agent", "jfrog-cli-go/" + cliutils.GetVersion())
}

func calcSha1(filePath string) string {
	file, err := os.Open(filePath)
	cliutils.CheckError(err)
	defer file.Close()

	var resSha1 []byte
	hashSha1 := sha1.New()
	_, err = io.Copy(hashSha1, file)
	cliutils.CheckError(err)
	return hex.EncodeToString(hashSha1.Sum(resSha1))
}

func calcMd5(filePath string) string {
	file, err := os.Open(filePath)
	cliutils.CheckError(err)
	defer file.Close()

	var resMd5 []byte
	hashMd5 := md5.New()
	_, err = io.Copy(hashMd5, file)
	cliutils.CheckError(err)
	return hex.EncodeToString(hashMd5.Sum(resMd5))
}

func ReadCredentialsFromConsole(details, savedDetails cliutils.Credentials) {
	if details.GetUser() == "" {
		tempUser := ""
		ScanFromConsole("User", &tempUser, savedDetails.GetUser())
		details.SetUser(tempUser)
	}
	if details.GetPassword() == "" {
		print("Password: ")
		bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
		cliutils.CheckError(err)
		details.SetPassword(string(bytePassword))
		if details.GetPassword() == "" {
			details.SetPassword(savedDetails.GetPassword())
		}
	}
}

type ConcurrentDownloadFlags struct {
	DownloadPath string
	FileName     string
	LocalPath    string
	FileSize     int64
	SplitCount   int
	Flat         bool
}

type FileDetails struct {
	Md5          string
	Sha1         string
	Size         int64
	AcceptRanges bool
}

