package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/errors/httperrors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const VULNERABILITY = "__vuln"
const COMPONENT = "__comp"

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
	if len(vulnerabilities) > 0 {
		log.Info("Downloading vulnerabilities...")
		err := saveData(xrayTempDir, "vuln", zipSuffix, vulnerabilities)
		if err != nil {
			return err
		}
	} else {
		log.Info("There are no new vulnerabilities.")
	}

	if len(components) > 0 {
		log.Info("Downloading components...")
		err := saveData(xrayTempDir, "comp", zipSuffix, components)
		if err != nil {
			return err
		}
	} else {
		log.Info("There are no new components.")
	}

	return nil
}

func getUpdatesBaseUrl() string {
	url := os.Getenv("JFROG_CLI_JXRAY_UPDATES_API_URL")
	if url != "" {
		return url
	}
	return "https://jxray.jfrog.io/api/v1/updates/onboarding"
}

func buildUpdatesUrl(flags *OfflineUpdatesFlags) (string, error) {
	var queryParams string
	if flags.From > 0 && flags.To > 0 {
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
	url := getUpdatesBaseUrl()
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
	tempDir := os.TempDir()
	xrayDir := tempDir + "/jfrog/xray/"
	if err := os.MkdirAll(xrayDir, 0777); err != nil {
		errorutils.CheckError(err)
		return "", nil
	}
	return xrayDir, nil
}

func saveData(xrayTmpDir, filesPrefix, zipSuffix string, urlsList []string) error {
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
		fileName := fileutils.GetFileNameFromUrl(url)
		log.Info("Downloading", url)
		_, err := httputils.DownloadFile(url, dataDir, fileName, httputils.HttpClientDetails{})
		if err != nil {
			return err
		}
	}
	log.Info("Zipping files.")
	err = fileutils.ZipFolderFiles(dataDir, filesPrefix+zipSuffix+".zip")
	if err != nil {
		return err
	}
	log.Info("Done zipping files.")
	return nil
}

func getFilesList(updatesUrl string, flags *OfflineUpdatesFlags) (vulnerabilities []string, components []string, lastUpdate int64, err error) {
	log.Info("Getting updates...")
	headers := make(map[string]string)
	headers["X-Xray-License"] = flags.License
	httpClientDetails := httputils.HttpClientDetails{
		Headers: headers,
	}
	resp, body, _, err := httputils.SendGet(updatesUrl, false, httpClientDetails)
	if err != nil {
		errorutils.CheckError(err)
		return
	}
	if err = httperrors.CheckResponseStatus(resp, body, http.StatusOK); err != nil {
		errorutils.CheckError(errors.New("Response: " + err.Error()))
		return
	}

	var urls FilesList
	err = json.Unmarshal(body, &urls)
	if err != nil {
		err = errorutils.CheckError(errors.New("Failed parsing json response: " + string(body)))
		return
	}

	for _, v := range urls.Urls {
		if strings.Contains(v, VULNERABILITY) {
			vulnerabilities = append(vulnerabilities, v)
		} else if strings.Contains(v, COMPONENT) {
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
}

type FilesList struct {
	Last_update int64
	Urls        []string
}
