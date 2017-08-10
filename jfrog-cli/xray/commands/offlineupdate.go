package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"encoding/json"
	"io/ioutil"
	"os"
	"errors"
	"strings"
	"strconv"
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/errors/httperrors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
)

const VULNERABILITY = "__vuln"
const COMPONENT = "__comp"
const JXRAY_BASR_URL = "https://jxray.jfrog.io/api/v1/updates/"
const ONBOARDING_URL = JXRAY_BASR_URL + "onboarding"

var updatesUrl = ONBOARDING_URL

func OfflineUpdate(flags *OfflineUpdatesFlags) error {
	if err := buildUpdatesUrl(flags); err != nil {
		return err
	}
	vulnerabilities, components, last_update, err := getFilesList(flags)
	if err != nil {
		return err
	}
	zipSuffix := "_" + strconv.FormatInt(last_update, 10)
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
		log.Info("There aren't new vulnerabilities.")
	}

	if len(components) > 0 {
		log.Info("Downloading components...")
		err := saveData(xrayTempDir, "comp", zipSuffix, components)
		if err != nil {
			return err
		}
	} else {
		log.Info("There aren't new components.")
	}

	return nil
}

func buildUpdatesUrl(flags *OfflineUpdatesFlags) (err error) {
	var queryParams string
	if flags.From > 0 && flags.To > 0 {
		if err = validateDates(flags.From, flags.To); err != nil {
			return
		}
		queryParams += fmt.Sprintf("from=%v&to=%v", flags.From, flags.To)
	}
	if flags.Version != "" {
		if queryParams != "" {
			queryParams += "&"
		}
		queryParams += fmt.Sprintf("version=%v", flags.Version)
	}
	if queryParams != "" {
		updatesUrl += "?" + queryParams
	}

	return
}

func validateDates(from, to int64) (err error) {
	if from < 0 || to < 0 {
		err = errors.New("Invalid dates")
		errorutils.CheckError(err)
		return
	}
	if from > to {
		err = errors.New("Invalid dates range.")
		errorutils.CheckError(err)
		return
	}
	return
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
	for i, url := range urlsList {
		fileName := filesPrefix + strconv.Itoa(i) + ".json"
		log.Info("Downloading", url)
		_, err := httputils.DownloadFile(url, dataDir, fileName, httputils.HttpClientDetails{})
		if err != nil {
			return err
		}
	}
	log.Info("Zipping files.")
	err = fileutils.ZipFolderFiles(dataDir, filesPrefix + zipSuffix + ".zip")
	if err != nil {
		return err
	}
	log.Info("Done zipping files.")
	return nil
}

func getFilesList(flags *OfflineUpdatesFlags) ([]string, []string, int64, error) {
	log.Info("Getting updates...")
	headers := make(map[string]string)
	headers["X-Xray-License"] = flags.License
	httpClientDetails := httputils.HttpClientDetails{
		Headers: headers,
	}
	resp, body, _, err := httputils.SendGet(updatesUrl, false, httpClientDetails)
	if err != nil {
		errorutils.CheckError(err)
		return nil, nil, 0, err
	}
	if err = httperrors.CheckResponseStatus(resp, body, 200); err != nil {
		errorutils.CheckError(errors.New("Response: " + err.Error()))
		return nil, nil, 0, err
	}

	var urls FilesList
	err = json.Unmarshal(body, &urls)
	if err != nil {
		err = errorutils.CheckError(errors.New("Failed parsing json response: " + string(body)))
		return nil, nil, 0, err
	}

	var vulnerabilities, components []string
	for _, v := range urls.Urls {
		if strings.Contains(v, VULNERABILITY) {
			vulnerabilities = append(vulnerabilities, v)
		} else if strings.Contains(v, COMPONENT) {
			components = append(components, v)
		}
	}
	return vulnerabilities, components, urls.Last_update, nil
}

type OfflineUpdatesFlags struct {
	License   string
	From      int64
	To        int64
	Version   string
}

type FilesList struct {
	Last_update int64
	Urls        []string
}