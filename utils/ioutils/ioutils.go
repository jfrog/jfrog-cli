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
	"runtime"
	"strconv"
	"strings"
	"sync"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/types"
	"golang.org/x/crypto/ssh/terminal"
	"syscall"
	"path"
)

const SYMLINK_FILE_CONTENT = ""
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

func IsFileExists(path string) (bool, error) {
	if !IsPathExists(path) {
		return false, nil
	}
	f, err := os.Stat(path)
	err = cliutils.CheckError(err)
	if err != nil {
		return false, err
	}
	return !f.IsDir(), nil
}

func IsPathSymlink(path string) bool {
	f, _ := os.Lstat(path)
	return f != nil && IsFileSymlink(f)
}

func IsFileSymlink(file os.FileInfo) bool {
	return file.Mode() & os.ModeSymlink != 0
}

func IsDir(path string) (bool, error) {
	if !IsPathExists(path) {
		return false, nil
	}
	f, err := os.Stat(path)
	err = cliutils.CheckError(err)
	if err != nil {
		return false, err
	}
	return f.IsDir(), nil
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
		fileName = path[index + 1:]
		dir = path[:index]
		return
	}
	fileName = path
	dir = ""
	return
}

// Get the local path and filename from original file name and path according to targetPath
func GetLocalPathAndFile(originalFileName, relativePath, targetPath string, flat bool) (localTargetPath, fileName string) {
	targetFileName, targetDirPath := GetFileAndDirFromPath(targetPath)

	localTargetPath = targetDirPath
	if !flat {
		localTargetPath = path.Join(targetDirPath, relativePath)
	}

	fileName = originalFileName
	if targetFileName != "" {
		fileName = targetFileName
	}
	return
}

// Return the recursive list of files and directories in the specified path
func ListFilesRecursiveWalkIntoDirSymlink(path string, walkIntoDirSymlink bool) (fileList []string, err error) {
	fileList = []string{}
	err = Walk(path, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return nil
	}, walkIntoDirSymlink)
	err = cliutils.CheckError(err)
	return
}

// Return the list of files and directories in the specified path
func ListFiles(path string) ([]string, error) {
	sep := GetFileSeperator()
	if !strings.HasSuffix(path, sep) {
		path += sep
	}
	fileList := []string{}
	files, _ := ioutil.ReadDir(path)
	path = strings.TrimPrefix(path, "." + sep)

	for _, f := range files {
		filePath := path + f.Name()
		exists, err := IsFileExists(filePath)
		if err != nil {
			return nil, err
		}
		if exists || IsPathSymlink(filePath) {
			fileList = append(fileList, filePath)
		}
	}
	return fileList, nil
}

func sendGetLeaveBodyOpen(url string, allowRedirect bool, httpClientsDetails HttpClientDetails) (*http.Response, []byte, string, error) {
	return Send("GET", url, nil, allowRedirect, false, httpClientsDetails)
}

func sendGetForFileDownload(url string, allowRedirect bool, httpClientsDetails HttpClientDetails) (*http.Response, string, error) {
	resp, _, redirectUrl, err := sendGetLeaveBodyOpen(url, allowRedirect, httpClientsDetails)
	return resp, redirectUrl, err
}

func Stream(url string, httpClientsDetails HttpClientDetails) (*http.Response,[]byte, string, error) {
	return sendGetLeaveBodyOpen(url, true, httpClientsDetails)
}

func SendGet(url string, allowRedirect bool, httpClientsDetails HttpClientDetails) (*http.Response, []byte, string, error) {
	return Send("GET", url, nil, allowRedirect, true, httpClientsDetails)
}

func SendPost(url string, content []byte, httpClientsDetails HttpClientDetails) (resp *http.Response, body []byte, err error) {
	resp, body, _, err = Send("POST", url, content, true, true, httpClientsDetails)
	return
}

func SendPatch(url string, content []byte, httpClientsDetails HttpClientDetails) (resp *http.Response, body []byte, err error) {
	resp, body, _, err = Send("PATCH", url, content, true, true, httpClientsDetails)
	return
}

func SendDelete(url string, content []byte, httpClientsDetails HttpClientDetails) (resp *http.Response, body []byte, err error) {
	resp, body, _, err = Send("DELETE", url, content, true, true, httpClientsDetails)
	return
}

func SendHead(url string, httpClientsDetails HttpClientDetails) (resp *http.Response, err error) {
	resp, _, _, err = Send("HEAD", url, nil, true, true, httpClientsDetails)
	return
}

func SendPut(url string, content []byte, httpClientsDetails HttpClientDetails) (resp *http.Response, body []byte, err error) {
	resp, body, _, err = Send("PUT", url, content, true, true, httpClientsDetails)
	return
}

func getHttpClient(transport *http.Transport) *http.Client {
	client := &http.Client{}
	if transport != nil {
		client.Transport = transport
	}
	return client
}

func Send(method string, url string, content []byte, allowRedirect bool,
closeBody bool, httpClientsDetails HttpClientDetails) (*http.Response, []byte, string, error) {

	var req *http.Request
	var err error
	if content != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(content))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if cliutils.CheckError(err) != nil {
		return nil, nil, "", err
	}
	return doRequest(req, allowRedirect, closeBody, httpClientsDetails)
}

func doRequest(req *http.Request, allowRedirect bool, closeBody bool, httpClientsDetails HttpClientDetails)  (resp *http.Response, respBody []byte, redirectUrl string, err error)  {
	req.Close = true
	setAuthentication(req, httpClientsDetails)
	addUserAgentHeader(req)
	copyHeaders(httpClientsDetails, req)

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

	err = cliutils.CheckError(err)
	if err != nil {
		return
	}
	if closeBody {
		defer resp.Body.Close()
		respBody, _ = ioutil.ReadAll(resp.Body)
	}
	return
}

func getUploadRequestContent(file *os.File) io.Reader {
	var reqBody io.Reader
	reqBody = file
	if file == nil {
		reqBody = bytes.NewBuffer([]byte(SYMLINK_FILE_CONTENT))
	}
	return reqBody
}

func getFileSize(file *os.File) (int64, error) {
	size := int64(0)
	if file != nil {
		fileInfo, err := file.Stat()
		if cliutils.CheckError(err) != nil {
			return size, err
		}
		size = fileInfo.Size()
	}
	return size, nil
}

func copyHeaders(httpClientsDetails HttpClientDetails, req *http.Request) {
	if httpClientsDetails.Headers != nil {
		for name := range httpClientsDetails.Headers {
			req.Header.Set(name, httpClientsDetails.Headers[name])
		}
	}
}

func setRequestHeaders(httpClientsDetails HttpClientDetails, size int64, req *http.Request) {
	copyHeaders(httpClientsDetails, req)
	length := strconv.FormatInt(size, 10)
	req.Header.Set("Content-Length", length)
}

func UploadFile(f *os.File, url string, httpClientsDetails HttpClientDetails) (*http.Response, []byte, error) {
	size, err := getFileSize(f)
	if err != nil {
		return nil, nil, err
	}
	req, err := http.NewRequest("PUT", url, getUploadRequestContent(f))
	if cliutils.CheckError(err) != nil {
		return nil, nil, err
	}
	req.ContentLength = size
	req.Close = true

	setRequestHeaders(httpClientsDetails, size, req)
	setAuthentication(req, httpClientsDetails)
	addUserAgentHeader(req)

	client := getHttpClient(httpClientsDetails.Transport);
	resp, err := client.Do(req)
	if cliutils.CheckError(err) != nil {
		return nil, nil, err
	}

	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	return resp, body, nil
}

func CreateFilePath(localPath, fileName string) (string, error) {
	if localPath != "" {
		err := os.MkdirAll(localPath, 0777)
		if cliutils.CheckError(err) != nil {
			return "", err
		}
		fileName = localPath + "/" + fileName
	}
	return fileName, nil
}

func DownloadFile(downloadPath, localPath, fileName string, httpClientsDetails HttpClientDetails) (*http.Response, error) {
	resp, _, err := downloadFile(downloadPath, localPath, fileName, true, httpClientsDetails)
	return resp, err
}

func DownloadFileNoRedirect(downloadPath, localPath, fileName string, httpClientsDetails HttpClientDetails) (*http.Response, string, error) {
	return downloadFile(downloadPath, localPath, fileName, false, httpClientsDetails)
}

func downloadFile(downloadPath, localPath, fileName string, allowRedirect bool,
httpClientsDetails HttpClientDetails) (resp *http.Response, redirectUrl string, err error) {

	fileName, err = CreateFilePath(localPath, fileName)
	if err != nil {
		return
	}
	out, err := os.Create(fileName)
	err = cliutils.CheckError(err)
	if err != nil {
		return
	}
	defer out.Close()
	resp, redirectUrl, err = sendGetForFileDownload(downloadPath, allowRedirect, httpClientsDetails)
	defer resp.Body.Close()
	if err == nil {
		_, err = io.Copy(out, resp.Body)
		err = cliutils.CheckError(err)
		if err != nil {
			return
		}
	}
	return
}

func DownloadFileConcurrently(flags ConcurrentDownloadFlags, logMsgPrefix string, httpClientsDetails HttpClientDetails) error {
	var wg sync.WaitGroup
	chunkSize := flags.FileSize / int64(flags.SplitCount)
	mod := flags.FileSize % int64(flags.SplitCount)
	var err error
	for i := 0; i < flags.SplitCount; i++ {
		if err != nil {
			break
		}
		wg.Add(1)
		start := chunkSize * int64(i)
		end := chunkSize * (int64(i) + 1)
		if i == flags.SplitCount - 1 {
			end += mod
		}
		requestClientDetails := httpClientsDetails.Clone()
		go func(start, end int64, i int) {
			e := downloadFileRange(flags, start, end, i, logMsgPrefix, *requestClientDetails)
			if e != nil {
				err = e
			}
			wg.Done()
		}(start, end, i)
	}
	wg.Wait()

	if err != nil {
		return err
	}

	if !flags.Flat && flags.LocalPath != "" {
		os.MkdirAll(flags.LocalPath, 0777)
		flags.FileName = flags.LocalPath + "/" + flags.FileName
	}

	if IsPathExists(flags.FileName) {
		err := os.Remove(flags.FileName)
		err = cliutils.CheckError(err)
		if err != nil {
			return err
		}
	}

	destFile, err := os.Create(flags.FileName)
	err = cliutils.CheckError(err)
	if err != nil {
		return err
	}
	defer destFile.Close()
	for i := 0; i < flags.SplitCount; i++ {
		tempFilePath, err := GetTempDirPath()
		if err != nil {
			return err
		}
		tempFilePath += "/" + flags.FileName + "_" + strconv.Itoa(i)
		AppendFile(tempFilePath, destFile)
	}
	fmt.Println(logMsgPrefix + "Done downloading.")
	return nil
}

func downloadFileRange(flags ConcurrentDownloadFlags, start, end int64, currentSplit int, logMsgPrefix string,
httpClientsDetails HttpClientDetails) error {

	tempLoclPath, err := GetTempDirPath()
	if err != nil {
		return err
	}
	if !flags.Flat {
		tempLoclPath += "/" + flags.LocalPath
	}
	if httpClientsDetails.Headers == nil {
		httpClientsDetails.Headers = make(map[string]string)
	}
	httpClientsDetails.Headers["Range"] = "bytes=" + strconv.FormatInt(start, 10) + "-" + strconv.FormatInt(end - 1, 10)

	resp, _, err :=
			sendGetForFileDownload(flags.DownloadPath, false, httpClientsDetails)
	defer resp.Body.Close()
	err = cliutils.CheckError(err)
	if err != nil {
		return err
	}

	fmt.Println(logMsgPrefix + "[" + strconv.Itoa(currentSplit) + "]:", resp.Status + "...")
	os.MkdirAll(tempLoclPath, 0777)
	filePath := tempLoclPath + "/" + flags.FileName + "_" + strconv.Itoa(currentSplit)

	out, err := os.Create(filePath)
	err = cliutils.CheckError(err)
	defer out.Close()
	if err != nil {
		return err
	}
	_, err = io.Copy(out, resp.Body)
	err = cliutils.CheckError(err)
	return err
}

func GetTempDirPath() (string, error) {
	if tempDirPath == "" {
		err := cliutils.CheckError(errors.New("Function cannot be used before 'tempDirPath' is created."))
		if err != nil {
			return "", err
		}
	}
	return tempDirPath, nil
}

func CreateTempDirPath() error {
	if tempDirPath != "" {
		err := cliutils.CheckError(errors.New("'tempDirPath' has already been initialized."))
		if err != nil {
			return err
		}
	}
	path, err := ioutil.TempDir("", "jfrog.cli.")
	err = cliutils.CheckError(err)
	if err != nil {
		return err
	}
	tempDirPath = path
	return nil
}

func RemoveTempDir() error {
	defer func() {
		tempDirPath = ""
	}()

	exists, err := IsDirExists(tempDirPath)
	if err != nil {
		return err
	}
	if exists {
		return os.RemoveAll(tempDirPath)
	}
	return nil
}

func IsDirExists(path string) (bool, error) {
	if !IsPathExists(path) {
		return false, nil
	}
	f, err := os.Stat(path)
	err = cliutils.CheckError(err)
	if err != nil {
		return false, err
	}
	return f.IsDir(), nil
}

// Reads the content of the file in the source path and appends it to
// the file in the destination path.
func AppendFile(srcPath string, destFile *os.File) error {
	srcFile, err := os.Open(srcPath)
	err = cliutils.CheckError(err)
	if err != nil {
		return err
	}

	defer func() error {
		err := srcFile.Close()
		return cliutils.CheckError(err)
	}()

	reader := bufio.NewReader(srcFile)

	writer := bufio.NewWriter(destFile)
	buf := make([]byte, 1024000)
	for {
		n, err := reader.Read(buf)
		if err != io.EOF {
			err = cliutils.CheckError(err)
			if err != nil {
				return err
			}
		}
		if n == 0 {
			break
		}
		_, err = writer.Write(buf[:n])
		err = cliutils.CheckError(err)
		if err != nil {
			return err
		}
	}
	err = writer.Flush()
	return cliutils.CheckError(err)
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

func ReadFile(filePath string) ([]byte, error) {
	content, err := ioutil.ReadFile(filePath)
	err = cliutils.CheckError(err)
	return content, err
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

func GetFileDetails(filePath string) (*FileDetails, error) {
	var err error
	details := new(FileDetails)
	details.Md5, err = CalcMd5(filePath)
	if err != nil {
		return nil, err
	}
	details.Sha1, err = CalcSha1(filePath)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(filePath)
	err = cliutils.CheckError(err)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	err = cliutils.CheckError(err)
	if err != nil {
		return nil, err
	}
	details.Size = fileInfo.Size()
	return details, nil
}

func GetRemoteFileDetails(downloadUrl string, httpClientsDetails HttpClientDetails) (*FileDetails, error) {
	resp, err := SendHead(downloadUrl, httpClientsDetails)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 404 {
		return nil, errors.New("response: " + resp.Status)
	}
	fileSize, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	err = cliutils.CheckError(err)
	if err != nil {
		return nil, err
	}

	fileDetails := new(FileDetails)
	fileDetails.Md5 = resp.Header.Get("X-Checksum-Md5")
	fileDetails.Sha1 = resp.Header.Get("X-Checksum-Sha1")
	fileDetails.Size = fileSize
	fileDetails.AcceptRanges = types.CreateBoolEnum()
	fileDetails.AcceptRanges.SetValue(resp.Header.Get("Accept-Ranges") == "bytes")
	return fileDetails, nil
}

func setAuthentication(req *http.Request, httpClientsDetails HttpClientDetails) {
	//Set authentication
	if httpClientsDetails.ApiKey != "" {
		if httpClientsDetails.User != "" {
			req.SetBasicAuth(httpClientsDetails.User, httpClientsDetails.ApiKey)
		} else {
			req.Header.Set("X-JFrog-Art-Api", httpClientsDetails.ApiKey)
		}
	} else if httpClientsDetails.Password != "" {
		req.SetBasicAuth(httpClientsDetails.User, httpClientsDetails.Password)
	}
}

func addUserAgentHeader(req *http.Request) {
	req.Header.Set("User-Agent", cliutils.CliAgent + "/" + cliutils.GetVersion())
}

func CalcSha1(filePath string) (string, error) {
	file, err := os.Open(filePath)
	cliutils.CheckError(err)
	if err != nil {
		return "", err
	}
	defer file.Close()
	return GetSha1(file)
}

func GetSha1(input io.Reader) (string, error) {
	var resSha1 []byte
	hashSha1 := sha1.New()
	_, err := io.Copy(hashSha1, input)
	err = cliutils.CheckError(err)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hashSha1.Sum(resSha1)), nil
}

func CalcMd5(filePath string) (string, error) {
	var err error
	file, err := os.Open(filePath)
	err = cliutils.CheckError(err)
	if err != nil {
		return "", err
	}
	defer file.Close()
	return GetMd5(file)
}

func GetMd5(input io.Reader) (string, error) {
	var resMd5 []byte
	hashMd5 := md5.New()
	_, err := io.Copy(hashMd5, input)
	err = cliutils.CheckError(err)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hashMd5.Sum(resMd5)), nil
}

func ReadCredentialsFromConsole(details, savedDetails cliutils.Credentials) error {
	if details.GetUser() == "" {
		tempUser := ""
		ScanFromConsole("User", &tempUser, savedDetails.GetUser())
		details.SetUser(tempUser)
	}
	if details.GetPassword() == "" {
		print("Password: ")
		bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
		err = cliutils.CheckError(err)
		if err != nil {
			return err
		}
		details.SetPassword(string(bytePassword))
		if details.GetPassword() == "" {
			details.SetPassword(savedDetails.GetPassword())
		}
	}
	return nil
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
	AcceptRanges *types.BoolEnum
}

