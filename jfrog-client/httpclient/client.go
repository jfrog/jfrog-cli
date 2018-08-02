package httpclient

import (
	"bytes"
	"errors"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/errors/httperrors"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
	multifilereader "github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
	"github.com/mholt/archiver"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"crypto/sha1"
	"encoding/hex"
	"hash"
)

func (jc *HttpClient) sendGetLeaveBodyOpen(url string, followRedirect bool, httpClientsDetails httputils.HttpClientDetails) (resp *http.Response, respBody []byte, redirectUrl string, err error) {
	return jc.Send("GET", url, nil, followRedirect, false, httpClientsDetails)
}

type HttpClient struct {
	Client *http.Client
}

func NewDefaultHttpClient() *HttpClient {
	return &HttpClient{Client: &http.Client{}}
}

func NewHttpClient(client *http.Client) *HttpClient {
	return &HttpClient{Client: client}
}

func (jc *HttpClient) sendGetForFileDownload(url string, followRedirect bool, httpClientsDetails httputils.HttpClientDetails, currentSplit, retries int) (resp *http.Response, redirectUrl string, err error) {
	for i := 0; i < retries+1; i++ {
		resp, _, redirectUrl, err = jc.sendGetLeaveBodyOpen(url, followRedirect, httpClientsDetails)
		if resp != nil && resp.StatusCode <= 500 {
			// No error and status <= 500
			return
		}
		log.Warn("Download attempt #", i, "of part", currentSplit, "of", url, "failed.")
	}
	return
}

func (jc *HttpClient) Stream(url string, httpClientsDetails httputils.HttpClientDetails) (*http.Response, []byte, string, error) {
	return jc.sendGetLeaveBodyOpen(url, true, httpClientsDetails)
}

func (jc *HttpClient) SendGet(url string, followRedirect bool, httpClientsDetails httputils.HttpClientDetails) (resp *http.Response, respBody []byte, redirectUrl string, err error) {
	return jc.Send("GET", url, nil, followRedirect, true, httpClientsDetails)
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

func (jc *HttpClient) Send(method string, url string, content []byte, followRedirect bool, closeBody bool, httpClientsDetails httputils.HttpClientDetails) (resp *http.Response, respBody []byte, redirectUrl string, err error) {
	var req *http.Request
	if content != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(content))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if errorutils.CheckError(err) != nil {
		return nil, nil, "", err
	}

	return jc.doRequest(req, content, followRedirect, closeBody, httpClientsDetails)
}

func (jc *HttpClient) doRequest(req *http.Request, content []byte, followRedirect bool, closeBody bool, httpClientsDetails httputils.HttpClientDetails) (resp *http.Response, respBody []byte, redirectUrl string, err error) {
	req.Close = true
	setAuthentication(req, httpClientsDetails)
	addUserAgentHeader(req)
	copyHeaders(httpClientsDetails, req)

	if !followRedirect || (followRedirect && req.Method == "POST") {
		jc.Client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			redirectUrl = req.URL.String()
			return errors.New("redirect")
		}
	}

	resp, err = jc.Client.Do(req)
	jc.Client.CheckRedirect = nil

	if err != nil && redirectUrl != "" {
		if !followRedirect {
			log.Debug("Blocking HTTP redirect to ", redirectUrl)
			return
		}
		// Due to security reasons, there's no built in HTTP redirect in the HTTP Client
		// for POST requests. We therefore implement the redirect on our own.
		if req.Method == "POST" {
			log.Debug("HTTP redirecting to ", redirectUrl)
			resp, respBody, err = jc.SendPost(redirectUrl, content, httpClientsDetails)
			redirectUrl = ""
			return
		}
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

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if errorutils.CheckError(err) != nil {
		return nil, nil, err
	}
	return resp, body, nil
}

// Read remote file,
// The caller is responsible to close the io.ReaderCloser
func (jc *HttpClient) ReadRemoteFile(downloadPath string, httpClientsDetails httputils.HttpClientDetails, retries int) (io.ReadCloser, error) {
	resp, _, err := jc.sendGetForFileDownload(downloadPath, true, httpClientsDetails, 0, retries)
	if err != nil {
		return nil, err
	}
	if err = httperrors.CheckResponseStatus(resp, nil, http.StatusOK); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (jc *HttpClient) DownloadFile(downloadFileDetails *DownloadFileDetails, logMsgPrefix string, httpClientsDetails httputils.HttpClientDetails, retries int, isExplode bool) (*http.Response, error) {
	resp, _, err := jc.downloadFile(downloadFileDetails, logMsgPrefix, true, httpClientsDetails, retries, isExplode)
	return resp, err
}

func (jc *HttpClient) DownloadFileNoRedirect(downloadPath, localPath, fileName string, httpClientsDetails httputils.HttpClientDetails) (*http.Response, string, error) {
	downloadFileDetails := &DownloadFileDetails{DownloadPath: downloadPath, LocalPath: localPath, FileName: fileName}
	return jc.downloadFile(downloadFileDetails, "", false, httpClientsDetails, 0, false)
}

func (jc *HttpClient) downloadFile(downloadFileDetails *DownloadFileDetails, logMsgPrefix string, followRedirect bool,
	httpClientsDetails httputils.HttpClientDetails, retries int, isExplode bool) (resp *http.Response, redirectUrl string, err error) {
	resp, redirectUrl, err = jc.sendGetForFileDownload(downloadFileDetails.DownloadPath, followRedirect, httpClientsDetails, 0, retries)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	if err = httperrors.CheckResponseStatus(resp, nil, http.StatusOK); err != nil {
		return
	}

	isZip := fileutils.IsZip(downloadFileDetails.FileName)
	arch := archiver.MatchingFormat(downloadFileDetails.FileName)

	// If explode flag is true and the file is an archive but not zip, extract the file.
	if isExplode && !isZip && arch != nil {
		err = extractFile(downloadFileDetails, arch, resp.Body, logMsgPrefix)
		return
	}

	// Save the file to the file system
	err = saveToFile(downloadFileDetails, resp.Body)
	if err != nil {
		return
	}

	// Extract zip if necessary
	// Extracting zip after download to prevent out of memory issues.
	if isExplode && isZip {
		err = extractZip(downloadFileDetails, logMsgPrefix)
	}
	return
}

func saveToFile(downloadFileDetails *DownloadFileDetails, body io.ReadCloser) error {
	fileName, err := fileutils.CreateFilePath(downloadFileDetails.LocalPath, downloadFileDetails.LocalFileName)
	if err != nil {
		return err
	}

	out, err := os.Create(fileName)
	if errorutils.CheckError(err) != nil {
		return err
	}

	defer out.Close()
	if len(downloadFileDetails.ExpectedSha1) > 0 {
		actualSha1 := sha1.New()
		writer := io.MultiWriter(actualSha1, out)
		_, err = io.Copy(writer, body)
		if hex.EncodeToString(actualSha1.Sum(nil)) != downloadFileDetails.ExpectedSha1 {
			err = errors.New("Checksum mismatch for " + fileName + ", expected: " + downloadFileDetails.ExpectedSha1 + ", actual: " + hex.EncodeToString(actualSha1.Sum(nil)))
		}
	} else {
		_, err = io.Copy(out, body)
	}

	return errorutils.CheckError(err)
}

func extractFile(downloadFileDetails *DownloadFileDetails, arch archiver.Archiver, reader io.Reader, logMsgPrefix string) error {
	log.Info(logMsgPrefix+"Extracting archive:", downloadFileDetails.FileName, "to", downloadFileDetails.LocalPath)
	err := fileutils.CreateDirIfNotExist(downloadFileDetails.LocalPath)
	if err != nil {
		return err
	}
	err = arch.Read(reader, downloadFileDetails.LocalPath)
	return errorutils.CheckError(err)
}

func extractZip(downloadFileDetails *DownloadFileDetails, logMsgPrefix string) error {
	fileName, err := fileutils.CreateFilePath(downloadFileDetails.LocalPath, downloadFileDetails.LocalFileName)
	if err != nil {
		return err
	}
	log.Info(logMsgPrefix+"Extracting archive:", fileName, "to", downloadFileDetails.LocalPath)
	err = archiver.Zip.Open(fileName, downloadFileDetails.LocalPath)
	if errorutils.CheckError(err) != nil {
		return err
	}
	err = os.Remove(fileName)
	return errorutils.CheckError(err)
}

func (jc *HttpClient) DownloadFileConcurrently(flags ConcurrentDownloadFlags, logMsgPrefix string, httpClientsDetails httputils.HttpClientDetails) error {
	chunksPaths := make([]string, flags.SplitCount)

	err := jc.downloadChunksConcurrently(chunksPaths, flags, logMsgPrefix, httpClientsDetails)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if !flags.Flat && flags.LocalPath != "" {
		os.MkdirAll(flags.LocalPath, 0777)
		flags.LocalFileName = filepath.Join(flags.LocalPath, flags.LocalFileName)
	}

	if fileutils.IsPathExists(flags.LocalFileName) {
		err := os.Remove(flags.LocalFileName)
		if errorutils.CheckError(err) != nil {
			return err
		}
	}

	// Explode and merge archive if necessary
	if flags.Explode {
		extracted, err := extractAndMergeChunks(chunksPaths, flags, logMsgPrefix)
		if extracted || err != nil {
			return err
		}
	}

	err = mergeChunks(chunksPaths, flags)
	if errorutils.CheckError(err) != nil {
		return err
	}
	log.Info(logMsgPrefix + "Done downloading.")
	return nil
}

func (jc *HttpClient) GetRemoteFileDetails(downloadUrl string, httpClientsDetails httputils.HttpClientDetails) (*fileutils.FileDetails, error) {
	resp, body, err := jc.SendHead(downloadUrl, httpClientsDetails)
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

func (jc *HttpClient) downloadChunksConcurrently(chunksPaths []string, flags ConcurrentDownloadFlags, logMsgPrefix string, httpClientsDetails httputils.HttpClientDetails) error {
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
			var downloadErr error
			chunksPaths[i], downloadErr = jc.downloadFileRange(flags, start, end, i, logMsgPrefix, *requestClientDetails, flags.Retries)
			if downloadErr != nil {
				err = downloadErr
			}
			wg.Done()
		}(start, end, i)
	}
	wg.Wait()
	return err
}

func mergeChunks(chunksPaths []string, flags ConcurrentDownloadFlags) error {
	destFile, err := os.OpenFile(flags.LocalFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if errorutils.CheckError(err) != nil {
		return err
	}
	defer destFile.Close()
	var writer io.Writer
	var actualSha1 hash.Hash
	if len(flags.ExpectedSha1) > 0 {
		actualSha1 = sha1.New()
		writer = io.MultiWriter(actualSha1, destFile)
	} else {
		writer = io.MultiWriter(destFile)
	}
	for i := 0; i < flags.SplitCount; i++ {
		reader, err := os.Open(chunksPaths[i])
		if err != nil {
			return err
		}
		defer reader.Close()
		_, err = io.Copy(writer, reader)
		if err != nil {
			return err
		}
	}
	if len(flags.ExpectedSha1) > 0 {
		if hex.EncodeToString(actualSha1.Sum(nil)) != flags.ExpectedSha1 {
			err = errors.New("Checksum mismatch for  " + flags.LocalFileName + ", expected: " + flags.ExpectedSha1 + ", actual: " + hex.EncodeToString(actualSha1.Sum(nil)))
		}
	}
	return err
}

func extractAndMergeChunks(chunksPaths []string, flags ConcurrentDownloadFlags, logMsgPrefix string) (bool, error) {
	if fileutils.IsZip(flags.FileName) {
		multiReader, err := multifilereader.NewMultiFileReaderAt(chunksPaths)
		if errorutils.CheckError(err) != nil {
			return false, err
		}
		log.Info(logMsgPrefix+"Extracting archive:", flags.FileName, "to", flags.LocalPath)
		err = fileutils.Unzip(multiReader, multiReader.Size(), flags.LocalPath)
		if errorutils.CheckError(err) != nil {
			return false, err
		}
		return true, nil
	}

	arch := archiver.MatchingFormat(flags.FileName)
	if arch == nil {
		log.Debug(logMsgPrefix+"Not an archive:", flags.FileName, "downloading file without extracting it.")
		return false, nil
	}

	fileReaders := make([]io.Reader, len(chunksPaths))
	var err error
	for k, v := range chunksPaths {
		f, err := os.Open(v)
		fileReaders[k] = f
		if err != nil {
			return false, errorutils.CheckError(err)
		}
		defer f.Close()
	}

	multiReader := io.MultiReader(fileReaders...)
	log.Info(logMsgPrefix+"Extracting archive:", flags.FileName, "to", flags.LocalPath)
	err = arch.Read(multiReader, flags.LocalPath)
	if err != nil {
		return false, errorutils.CheckError(err)
	}
	return true, nil
}

func (jc *HttpClient) downloadFileRange(flags ConcurrentDownloadFlags, start, end int64, currentSplit int, logMsgPrefix string, httpClientsDetails httputils.HttpClientDetails, retries int) (string, error) {
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
	resp, _, err := jc.sendGetForFileDownload(flags.DownloadPath, true, httpClientsDetails, currentSplit, retries)
	if errorutils.CheckError(err) != nil {
		return "", err
	}
	defer resp.Body.Close()

	log.Info(logMsgPrefix+"["+strconv.Itoa(currentSplit)+"]:", resp.Status+"...")
	os.MkdirAll(tempLocalPath, 0777)

	_, err = io.Copy(tempFile, resp.Body)
	return tempFile.Name(), errorutils.CheckError(err)
}

func (jc *HttpClient) IsAcceptRanges(downloadUrl string, httpClientsDetails httputils.HttpClientDetails) (bool, error) {
	resp, body, err := jc.SendHead(downloadUrl, httpClientsDetails)
	if errorutils.CheckError(err) != nil {
		return false, err
	}

	if err = httperrors.CheckResponseStatus(resp, body, http.StatusOK); errorutils.CheckError(err) != nil {
		return false, err
	}
	return resp.Header.Get("Accept-Ranges") == "bytes", nil
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
	req.Header.Set("User-Agent", utils.GetUserAgent())
}

type DownloadFileDetails struct {
	FileName       string `json:"LocalFileName,omitempty"`
	DownloadPath   string `json:"DownloadPath,omitempty"`
	LocalPath      string `json:"LocalPath,omitempty"`
	LocalFileName  string `json:"LocalFileName,omitempty"`
	ExpectedSha1   string `json:"ExpectedSha1,omitempty"`
	Size           int64  `json:"Size,omitempty"`
}

type ConcurrentDownloadFlags struct {
	FileName      string
	DownloadPath  string
	LocalFileName string
	LocalPath     string
	ExpectedSha1  string
	FileSize      int64
	SplitCount    int
	Flat          bool
	Explode       bool
	Retries       int
}
