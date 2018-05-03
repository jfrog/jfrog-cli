package httputils

import (
	"bytes"
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/errors/httperrors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var UserAgent string

func sendGetLeaveBodyOpen(url string, followRedirect bool, httpClientsDetails HttpClientDetails) (*http.Response, []byte, string, error) {
	return Send("GET", url, nil, followRedirect, false, httpClientsDetails)
}

func sendGetForFileDownload(url string, followRedirect bool, httpClientsDetails HttpClientDetails) (*http.Response, string, error) {
	resp, _, redirectUrl, err := sendGetLeaveBodyOpen(url, followRedirect, httpClientsDetails)
	return resp, redirectUrl, err
}

func Stream(url string, httpClientsDetails HttpClientDetails) (*http.Response, []byte, string, error) {
	return sendGetLeaveBodyOpen(url, true, httpClientsDetails)
}

func SendGet(url string, followRedirect bool, httpClientsDetails HttpClientDetails) (*http.Response, []byte, string, error) {
	return Send("GET", url, nil, followRedirect, true, httpClientsDetails)
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

func SendHead(url string, httpClientsDetails HttpClientDetails) (resp *http.Response, body []byte, err error) {
	resp, body, _, err = Send("HEAD", url, nil, true, true, httpClientsDetails)
	return
}

func SendPut(url string, content []byte, httpClientsDetails HttpClientDetails) (resp *http.Response, body []byte, err error) {
	resp, body, _, err = Send("PUT", url, content, true, true, httpClientsDetails)
	return
}

func IsSsh(urlPath string) bool {
	u, err := url.Parse(urlPath)
	if err != nil {
		return false
	}
	return strings.ToLower(u.Scheme) == "ssh"
}

func getHttpClient(transport *http.Transport) *http.Client {
	client := &http.Client{}
	if transport != nil {
		client.Transport = transport
	}
	return client
}

func Send(method string, url string, content []byte, followRedirect, closeBody bool, httpClientsDetails HttpClientDetails) (*http.Response, []byte, string, error) {
	var req *http.Request
	var err error
	if content != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(content))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if errorutils.CheckError(err) != nil {
		return nil, nil, "", err
	}

	return doRequest(req, followRedirect, closeBody, httpClientsDetails)
}

func doRequest(req *http.Request, followRedirect, closeBody bool, httpClientsDetails HttpClientDetails) (resp *http.Response, respBody []byte, redirectUrl string, err error) {
	req.Close = true
	setAuthentication(req, httpClientsDetails)
	addUserAgentHeader(req)
	copyHeaders(httpClientsDetails, req)

	client := getHttpClient(httpClientsDetails.Transport)
	if !followRedirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			redirectUrl = req.URL.String()
			return errors.New("redirect")
		}
	}

	resp, err = client.Do(req)
	if closeBody && resp != nil {
		defer resp.Body.Close()
		respBody, _ = ioutil.ReadAll(resp.Body)
	}
	if !followRedirect && err != nil {
		return
	}
	err = errorutils.CheckError(err)
	return
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
	size, err := fileutils.GetFileSize(f)
	if err != nil {
		return nil, nil, err
	}
	req, err := http.NewRequest("PUT", url, fileutils.GetUploadRequestContent(f))
	if errorutils.CheckError(err) != nil {
		return nil, nil, err
	}
	req.ContentLength = size
	req.Close = true

	setRequestHeaders(httpClientsDetails, size, req)
	setAuthentication(req, httpClientsDetails)
	addUserAgentHeader(req)

	client := getHttpClient(httpClientsDetails.Transport)
	resp, err := client.Do(req)
	if errorutils.CheckError(err) != nil {
		return nil, nil, err
	}

	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	return resp, body, nil
}

func DownloadFile(downloadPath, localPath, fileName string, httpClientsDetails HttpClientDetails) (*http.Response, error) {
	resp, _, err := downloadFile(downloadPath, localPath, fileName, true, httpClientsDetails)
	return resp, err
}

func DownloadFileNoRedirect(downloadPath, localPath, fileName string, httpClientsDetails HttpClientDetails) (*http.Response, string, error) {
	return downloadFile(downloadPath, localPath, fileName, false, httpClientsDetails)
}

func downloadFile(downloadPath, localPath, fileName string, followRedirect bool,
	httpClientsDetails HttpClientDetails) (resp *http.Response, redirectUrl string, err error) {
	resp, redirectUrl, err = sendGetForFileDownload(downloadPath, followRedirect, httpClientsDetails)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	if err = httperrors.CheckResponseStatus(resp, nil, http.StatusOK); err != nil {
		return
	}

	fileName, err = fileutils.CreateFilePath(localPath, fileName)
	if err != nil {
		return
	}

	out, err := os.Create(fileName)
	err = errorutils.CheckError(err)
	if err != nil {
		return
	}

	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	err = errorutils.CheckError(err)
	return
}

func DownloadFileConcurrently(flags ConcurrentDownloadFlags, logMsgPrefix string, httpClientsDetails HttpClientDetails) error {
	var wg sync.WaitGroup
	chunkSize := flags.FileSize / int64(flags.SplitCount)
	mod := flags.FileSize % int64(flags.SplitCount)
	chuckPaths := make([]string, flags.SplitCount)
	errors := make([]error, flags.SplitCount)
	for i := 0; i < flags.SplitCount; i++ {
		wg.Add(1)
		start := chunkSize * int64(i)
		end := chunkSize * (int64(i) + 1)
		if i == flags.SplitCount-1 {
			end += mod
		}
		requestClientDetails := httpClientsDetails.Clone()
		go func(start, end int64, i int) {
			chuckPaths[i], errors[i] = downloadFileRange(flags, start, end, i, logMsgPrefix, *requestClientDetails)
			wg.Done()
		}(start, end, i)
	}
	wg.Wait()

	for _, e := range errors {
		if e != nil {
			return errorutils.CheckError(e)
		}
	}

	if flags.LocalPath != "" {
		os.MkdirAll(flags.LocalPath, 0777)
		flags.FileName = filepath.Join(flags.LocalPath, flags.FileName)
	}

	if fileutils.IsPathExists(flags.FileName) {
		err := os.Remove(flags.FileName)
		err = errorutils.CheckError(err)
		if err != nil {
			return err
		}
	}

	destFile, err := os.Create(flags.FileName)
	err = errorutils.CheckError(err)
	if err != nil {
		return err
	}
	defer destFile.Close()
	for i := 0; i < flags.SplitCount; i++ {
		fileutils.AppendFile(chuckPaths[i], destFile)
	}
	log.Info(logMsgPrefix + "Done downloading.")
	return nil
}

func downloadFileRange(flags ConcurrentDownloadFlags, start, end int64, currentSplit int, logMsgPrefix string,
	httpClientsDetails HttpClientDetails) (string, error) {

	tempLocalPath, err := fileutils.GetTempDirPath()
	if err != nil {
		return "", err
	}

	tempFile, err := ioutil.TempFile(tempLocalPath, strconv.Itoa(currentSplit)+"_")
	if errorutils.CheckError(err) != nil {
		return "", err
	}
	defer tempFile.Close()

	if httpClientsDetails.Headers == nil {
		httpClientsDetails.Headers = make(map[string]string)
	}
	httpClientsDetails.Headers["Range"] = "bytes=" + strconv.FormatInt(start, 10) + "-" + strconv.FormatInt(end-1, 10)
	resp, _, err := sendGetForFileDownload(flags.DownloadPath, false, httpClientsDetails)
	if errorutils.CheckError(err) != nil {
		return "", err
	}
	defer resp.Body.Close()

	log.Info(logMsgPrefix+"["+strconv.Itoa(currentSplit)+"]:", resp.Status+"...")
	os.MkdirAll(tempLocalPath, 0777)

	_, err = io.Copy(tempFile, resp.Body)
	return tempFile.Name(), errorutils.CheckError(err)
}

func GetRemoteFileDetails(downloadUrl string, httpClientsDetails HttpClientDetails) (*fileutils.FileDetails, error) {
	resp, body, err := SendHead(downloadUrl, httpClientsDetails)
	if err != nil {
		return nil, err
	}

	if err = httperrors.CheckResponseStatus(resp, body, http.StatusOK); err != nil {
		return nil, err
	}

	fileSize := int64(0)
	contentLength := resp.Header.Get("Content-Length")
	if len(contentLength) > 0 {
		fileSize, err = strconv.ParseInt(contentLength, 10, 64)
		if err != nil {
			return nil, err
		}
	}

	fileDetails := new(fileutils.FileDetails)
	fileDetails.Checksum.Md5 = resp.Header.Get("X-Checksum-Md5")
	fileDetails.Checksum.Sha1 = resp.Header.Get("X-Checksum-Sha1")
	fileDetails.Size = fileSize
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
	if UserAgent != "" {
		req.Header.Set("User-Agent", UserAgent)
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
