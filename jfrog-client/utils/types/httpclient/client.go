package httpclient

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/errors/httperrors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/types"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/httputils"
)

type HttpClient struct {
	Client *http.Client
}

func NewDefaultJforgHttpClient() *HttpClient {
	return &HttpClient{Client: &http.Client{}}
}

func NewJforgHttpClient(client *http.Client) *HttpClient {
	return &HttpClient{Client: client}
}

func (jc *HttpClient) sendGetLeaveBodyOpen(url string, allowRedirect bool, httpClientsDetails httputils.HttpClientDetails) (*http.Response, []byte, string, error) {
	return jc.Send("GET", url, nil, allowRedirect, false, httpClientsDetails)
}

func (jc *HttpClient) sendGetForFileDownload(url string, allowRedirect bool, httpClientsDetails httputils.HttpClientDetails) (*http.Response, string, error) {
	resp, _, redirectUrl, err := jc.sendGetLeaveBodyOpen(url, allowRedirect, httpClientsDetails)
	return resp, redirectUrl, err
}

func (jc *HttpClient) Stream(url string, httpClientsDetails httputils.HttpClientDetails) (*http.Response, []byte, string, error) {
	return jc.sendGetLeaveBodyOpen(url, true, httpClientsDetails)
}

func (jc *HttpClient) SendGet(url string, allowRedirect bool, httpClientsDetails httputils.HttpClientDetails) (*http.Response, []byte, string, error) {
	return jc.Send("GET", url, nil, allowRedirect, true, httpClientsDetails)
}

func (jc *HttpClient) SendPost(url string, content []byte, httpClientsDetails httputils.HttpClientDetails) (resp *http.Response, body []byte, err error) {
	resp, body, _, err = jc.Send("POST", url, content, true, true, httpClientsDetails)
	return
}

func (jc *HttpClient) SendPatch(url string, content []byte, httpClientsDetails httputils.HttpClientDetails) (resp *http.Response, body []byte, err error) {
	resp, body, _, err = jc.Send("PATCH", url, content, true, true, httpClientsDetails)
	return
}

func (jc *HttpClient) SendDelete(url string, content []byte, httpClientsDetails httputils.HttpClientDetails) (resp *http.Response, body []byte, err error) {
	resp, body, _, err = jc.Send("DELETE", url, content, true, true, httpClientsDetails)
	return
}

func (jc *HttpClient) SendHead(url string, httpClientsDetails httputils.HttpClientDetails) (resp *http.Response, body []byte, err error) {
	resp, body, _, err = jc.Send("HEAD", url, nil, true, true, httpClientsDetails)
	return
}

func (jc *HttpClient) SendPut(url string, content []byte, httpClientsDetails httputils.HttpClientDetails) (resp *http.Response, body []byte, err error) {
	resp, body, _, err = jc.Send("PUT", url, content, true, true, httpClientsDetails)
	return
}

func (jc *HttpClient) Send(method string, url string, content []byte, allowRedirect bool, closeBody bool, httpClientsDetails httputils.HttpClientDetails) (*http.Response, []byte, string, error) {
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

	return jc.doRequest(req, allowRedirect, closeBody, httpClientsDetails)
}

func (jc *HttpClient) doRequest(req *http.Request, allowRedirect bool, closeBody bool, httpClientsDetails httputils.HttpClientDetails) (resp *http.Response, respBody []byte, redirectUrl string, err error) {
	req.Close = true
	setAuthentication(req, httpClientsDetails)
	addUserAgentHeader(req)
	copyHeaders(httpClientsDetails, req)

	client := jc.Client
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

	err = errorutils.CheckError(err)
	if err != nil {
		return
	}
	if closeBody {
		defer resp.Body.Close()
		respBody, _ = ioutil.ReadAll(resp.Body)
	}
	return
}

func copyHeaders(httpClientsDetails httputils.HttpClientDetails, req *http.Request) {
	if httpClientsDetails.Headers != nil {
		for name := range httpClientsDetails.Headers {
			req.Header.Set(name, httpClientsDetails.Headers[name])
		}
	}
}

func setRequestHeaders(httpClientsDetails httputils.HttpClientDetails, size int64, req *http.Request) {
	copyHeaders(httpClientsDetails, req)
	length := strconv.FormatInt(size, 10)
	req.Header.Set("Content-Length", length)
}

func (jc *HttpClient) UploadFile(f *os.File, url string, httpClientsDetails httputils.HttpClientDetails) (*http.Response, []byte, error) {
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

	client := jc.Client
	resp, err := client.Do(req)
	if errorutils.CheckError(err) != nil {
		return nil, nil, err
	}

	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	return resp, body, nil
}

func (jc *HttpClient) DownloadFile(downloadPath, localPath, fileName string, httpClientsDetails httputils.HttpClientDetails) (*http.Response, error) {
	resp, _, err := jc.downloadFile(downloadPath, localPath, fileName, true, httpClientsDetails)
	return resp, err
}

func (jc *HttpClient) DownloadFileNoRedirect(downloadPath, localPath, fileName string, httpClientsDetails httputils.HttpClientDetails) (*http.Response, string, error) {
	return jc.downloadFile(downloadPath, localPath, fileName, false, httpClientsDetails)
}

func (jc *HttpClient) downloadFile(downloadPath, localPath, fileName string, allowRedirect bool,
	httpClientsDetails httputils.HttpClientDetails) (resp *http.Response, redirectUrl string, err error) {
	resp, redirectUrl, err = jc.sendGetForFileDownload(downloadPath, allowRedirect, httpClientsDetails)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	if err = httperrors.CheckResponseStatus(resp, nil, 200); err != nil {
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

func (jc *HttpClient) DownloadFileConcurrently(flags ConcurrentDownloadFlags, logMsgPrefix string, httpClientsDetails httputils.HttpClientDetails) error {
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
		if i == flags.SplitCount-1 {
			end += mod
		}
		requestClientDetails := httpClientsDetails.Clone()
		go func(start, end int64, i int) {
			e := jc.downloadFileRange(flags, start, end, i, logMsgPrefix, *requestClientDetails)
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
		tempFilePath, err := fileutils.GetTempDirPath()
		if err != nil {
			return err
		}
		tempFilePath += "/" + flags.FileName + "_" + strconv.Itoa(i)
		fileutils.AppendFile(tempFilePath, destFile)
	}
	log.Info(logMsgPrefix + "Done downloading.")
	return nil
}

func (jc *HttpClient) downloadFileRange(flags ConcurrentDownloadFlags, start, end int64, currentSplit int, logMsgPrefix string,
	httpClientsDetails httputils.HttpClientDetails) error {

	tempLocalPath, err := fileutils.GetTempDirPath()
	if err != nil {
		return err
	}
	if !flags.Flat {
		tempLocalPath += "/" + flags.LocalPath
	}
	if httpClientsDetails.Headers == nil {
		httpClientsDetails.Headers = make(map[string]string)
	}
	httpClientsDetails.Headers["Range"] = "bytes=" + strconv.FormatInt(start, 10) + "-" + strconv.FormatInt(end-1, 10)

	resp, _, err := jc.sendGetForFileDownload(flags.DownloadPath, false, httpClientsDetails)
	err = errorutils.CheckError(err)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Info(logMsgPrefix+"["+strconv.Itoa(currentSplit)+"]:", resp.Status+"...")
	os.MkdirAll(tempLocalPath, 0777)
	filePath := tempLocalPath + "/" + flags.FileName + "_" + strconv.Itoa(currentSplit)

	out, err := os.Create(filePath)
	err = errorutils.CheckError(err)
	defer out.Close()
	if err != nil {
		return err
	}
	_, err = io.Copy(out, resp.Body)
	err = errorutils.CheckError(err)
	return err
}

func (jc *HttpClient) GetRemoteFileDetails(downloadUrl string, httpClientsDetails httputils.HttpClientDetails) (*fileutils.FileDetails, error) {
	resp, body, err := jc.SendHead(downloadUrl, httpClientsDetails)
	if err != nil {
		return nil, err
	}

	if err = httperrors.CheckResponseStatus(resp, body, 200); err != nil {
		return nil, err
	}

	fileSize, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	err = errorutils.CheckError(err)
	if err != nil {
		return nil, err
	}

	fileDetails := new(fileutils.FileDetails)
	fileDetails.Checksum.Md5 = resp.Header.Get("X-Checksum-Md5")
	fileDetails.Checksum.Sha1 = resp.Header.Get("X-Checksum-Sha1")
	fileDetails.Size = fileSize
	fileDetails.AcceptRanges = types.CreateBoolEnum()
	fileDetails.AcceptRanges.SetValue(resp.Header.Get("Accept-Ranges") == "bytes")
	return fileDetails, nil
}

func setAuthentication(req *http.Request, httpClientsDetails httputils.HttpClientDetails) {
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
	req.Header.Set("User-Agent", cliutils.CliAgent+"/"+cliutils.GetVersion())
}

type ConcurrentDownloadFlags struct {
	DownloadPath string
	FileName     string
	LocalPath    string
	FileSize     int64
	SplitCount   int
	Flat         bool
}
