package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/httpclient"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	Vulnerability       = "__vuln"
	Component           = "__comp"
	JxrayDefaultBaseUrl = "https://jxray.jfrog.io/"
	JxrayApiBundles     = "api/v1/updates/bundles"
	JxrayApiOnboarding  = "api/v1/updates/onboarding"
)

func OfflineUpdate(flags *OfflineUpdatesFlags) error {
	updatesUrl, err := buildUpdatesUrl(flags)
	if err != nil {
		return err
	}
	vulnerabilities, components, lastUpdate, err := getFilesList(updatesUrl, flags)
	if err != nil {
		return err
	}
	zipSuffix := "_" + strconv.FormatInt(lastUpdate, 10)
	xrayTempDir, err := getXrayTempDir()
	if err != nil {
		return err
	}

	if flags.Target != "" && (len(vulnerabilities) > 0 || len(components) > 0) {
		err = os.MkdirAll(flags.Target, 0777)
		if errorutils.CheckError(err) != nil {
			return err
		}
	}

	if len(vulnerabilities) > 0 {
		log.Info("Downloading vulnerabilities...")
		err := saveData(xrayTempDir, "vuln", zipSuffix, flags.Target, vulnerabilities)
		if err != nil {
			return err
		}
	} else {
		log.Info("There are no new vulnerabilities.")
	}

	if len(components) > 0 {
		log.Info("Downloading components...")
		err := saveData(xrayTempDir, "comp", zipSuffix, flags.Target, components)
		if err != nil {
			return err
		}
	} else {
		log.Info("There are no new components.")
	}

	return nil
}

func getUpdatesBaseUrl(datesSpecified bool) string {
	jxRayBaseUrl := os.Getenv("JFROG_CLI_JXRAY_BASE_URL")
	jxRayBaseUrl = utils.AddTrailingSlashIfNeeded(jxRayBaseUrl)
	if jxRayBaseUrl == "" {
		jxRayBaseUrl = JxrayDefaultBaseUrl
	}
	if datesSpecified {
		return jxRayBaseUrl + JxrayApiBundles
	}
	return jxRayBaseUrl + JxrayApiOnboarding
}

func buildUpdatesUrl(flags *OfflineUpdatesFlags) (string, error) {
	var queryParams string
	datesSpecified := flags.From > 0 && flags.To > 0
	if datesSpecified {
		if err := validateDates(flags.From, flags.To); err != nil {
			return "", err
		}
		queryParams += fmt.Sprintf("from=%v&to=%v", flags.From, flags.To)
	}
	if flags.Version != "" {
		if queryParams != "" {
			queryParams += "&"
		}
		queryParams += fmt.Sprintf("version=%v", flags.Version)
	}
	url := getUpdatesBaseUrl(datesSpecified)
	if queryParams != "" {
		url += "?" + queryParams
	}
	return url, nil
}

func validateDates(from, to int64) error {
	if from < 0 || to < 0 {
		err := errors.New("Invalid dates")
		return errorutils.CheckError(err)
	}
	if from > to {
		err := errors.New("Invalid dates range.")
		return errorutils.CheckError(err)
	}
	return nil
}

func getXrayTempDir() (string, error) {
	xrayDir := filepath.Join(cliutils.GetCliPersistentTempDirPath(), "jfrog", "xray")
	if err := os.MkdirAll(xrayDir, 0777); err != nil {
		errorutils.CheckError(err)
		return "", nil
	}
	return xrayDir, nil
}

func saveData(xrayTmpDir, filesPrefix, zipSuffix, targetPath string, urlsList []string) error {
	dataDir, err := ioutil.TempDir(xrayTmpDir, filesPrefix)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := os.RemoveAll(dataDir); cerr != nil && err == nil {
			err = cerr
		}
	}()
	for _, url := range urlsList {
		fileName, err := createXrayFileNameFromUrl(url)
		if err != nil {
			return err
		}
		log.Info("Downloading", url)
		client, err := httpclient.ClientBuilder().Build()
		if err != nil {
			return err
		}

		details := &httpclient.DownloadFileDetails{
			FileName:      fileName,
			DownloadPath:  url,
			LocalPath:     dataDir,
			LocalFileName: fileName}
		_, err = client.DownloadFile(details, "", httputils.HttpClientDetails{}, 3, false)
		if err != nil {
			return err
		}
	}
	log.Info("Zipping files.")
	err = fileutils.ZipFolderFiles(dataDir, filepath.Join(targetPath, filesPrefix+zipSuffix+".zip"))
	if err != nil {
		return err
	}
	log.Info("Done zipping files.")
	return nil
}

func createXrayFileNameFromUrl(url string) (fileName string, err error) {
	originalUrl := url
	index := strings.Index(url, "?")
	if index != -1 {
		url = url[:index]
	}
	index = strings.Index(url, ";")
	if index != -1 {
		url = url[:index]
	}

	sections := strings.Split(url, "/")
	length := len(sections)
	if length < 2 {
		err = errorutils.CheckError(errors.New(fmt.Sprintf("Unexpected URL format: %s", originalUrl)))
		return
	}
	fileName = fmt.Sprintf("%s__%s", sections[length-2], sections[length-1])
	return
}

func getFilesList(updatesUrl string, flags *OfflineUpdatesFlags) (vulnerabilities []string, components []string, lastUpdate int64, err error) {
	log.Info("Getting updates...")
	headers := make(map[string]string)
	headers["X-Xray-License"] = flags.License
	httpClientDetails := httputils.HttpClientDetails{
		Headers: headers,
	}
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		return
	}
	resp, body, _, err := client.SendGet(updatesUrl, false, httpClientDetails)
	if errorutils.CheckError(err) != nil {
		return
	}
	if err = errorutils.CheckResponseStatus(resp, http.StatusOK); err != nil {
		err = errorutils.CheckError(errors.New("Response: " + err.Error()))
		return
	}

	var urls FilesList
	err = json.Unmarshal(body, &urls)
	if err != nil {
		err = errorutils.CheckError(errors.New("Failed parsing json response: " + string(body)))
		return
	}

	for _, v := range urls.Urls {
		if strings.Contains(v, Vulnerability) {
			vulnerabilities = append(vulnerabilities, v)
		} else if strings.Contains(v, Component) {
			components = append(components, v)
		}
	}
	lastUpdate = urls.Last_update
	return
}

type OfflineUpdatesFlags struct {
	License string
	From    int64
	To      int64
	Version string
	Target  string
}

type FilesList struct {
	Last_update int64
	Urls        []string
}
