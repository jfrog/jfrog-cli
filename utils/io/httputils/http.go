package httputils

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
	"github.com/jfrogdev/jfrog-cli-go/utils/io/fileutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/types"
	"github.com/jfrogdev/jfrog-cli-go/errors/httperrors"
)

func sendGetLeaveBodyOpen(url string, allowRedirect bool, httpClientsDetails HttpClientDetails) (*http.Response, []byte, string, error) {
	return Send("GET", url, nil, allowRedirect, false, httpClientsDetails)
}

func sendGetForFileDownload(url string, allowRedirect bool, httpClientsDetails HttpClientDetails) (*http.Response, string, error) {
	resp, _, redirectUrl, err := sendGetLeaveBodyOpen(url, allowRedirect, httpClientsDetails)
	return resp, redirectUrl, err
}

func Stream(url string, httpClientsDetails HttpClientDetails) (*http.Response, []byte, string, error) {
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

func SendHead(url string, httpClientsDetails HttpClientDetails) (resp *http.Response, body []byte, err error) {
	resp, body, _, err = Send("HEAD", url, nil, true, true, httpClientsDetails)
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

func Send(method string, url string, content []byte, allowRedirect bool, closeBody bool, httpClientsDetails HttpClientDetails) (*http.Response, []byte, string, error) {
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

func doRequest(req *http.Request, allowRedirect bool, closeBody bool, httpClientsDetails HttpClientDetails) (resp *http.Response, respBody []byte, redirectUrl string, err error) {
	req.Close = true
	setAuthentication(req, httpClientsDetails)
	addUserAgentHeader(req)
	copyHeaders(httpClientsDetails, req)

	client := getHttpClient(httpClientsDetails.Transport)
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
	if cliutils.CheckError(err) != nil {
		return nil, nil, err
	}
	req.ContentLength = size
	req.Close = true

	setRequestHeaders(httpClientsDetails, size, req)
	setAuthentication(req, httpClientsDetails)
	addUserAgentHeader(req)

	client := getHttpClient(httpClientsDetails.Transport)
	resp, err := client.Do(req)
	if cliutils.CheckError(err) != nil {
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

func downloadFile(downloadPath, localPath, fileName string, allowRedirect bool,
httpClientsDetails HttpClientDetails) (resp *http.Response, redirectUrl string, err error) {
	resp, redirectUrl, err = sendGetForFileDownload(downloadPath, allowRedirect, httpClientsDetails)
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
	err = cliutils.CheckError(err)
	if err != nil {
		return
	}

	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	err = cliutils.CheckError(err)
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

	if fileutils.IsPathExists(flags.FileName) {
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

func downloadFileRange(flags ConcurrentDownloadFlags, start, end int64, currentSplit int, logMsgPrefix string,
httpClientsDetails HttpClientDetails) error {

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
	httpClientsDetails.Headers["Range"] = "bytes=" + strconv.FormatInt(start, 10) + "-" + strconv.FormatInt(end - 1, 10)

	resp, _, err := sendGetForFileDownload(flags.DownloadPath, false, httpClientsDetails)
	err = cliutils.CheckError(err)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Info(logMsgPrefix + "[" + strconv.Itoa(currentSplit) + "]:", resp.Status + "...")
	os.MkdirAll(tempLocalPath, 0777)
	filePath := tempLocalPath + "/" + flags.FileName + "_" + strconv.Itoa(currentSplit)

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

func GetRemoteFileDetails(downloadUrl string, httpClientsDetails HttpClientDetails) (*fileutils.FileDetails, error) {
	resp, body, err := SendHead(downloadUrl, httpClientsDetails)
	if err != nil {
		return nil, err
	}

	if err = httperrors.CheckResponseStatus(resp, body, 200); err != nil {
		return nil, err
	}

	fileSize, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	err = cliutils.CheckError(err)
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

type ConcurrentDownloadFlags struct {
	DownloadPath string
	FileName     string
	LocalPath    string
	FileSize     int64
	SplitCount   int
	Flat         bool
}
