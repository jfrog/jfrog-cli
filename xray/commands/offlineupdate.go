package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/utils/io/httputils"
	"encoding/json"
	"io/ioutil"
	"os"
	"archive/zip"
	"io"
	"errors"
	"path/filepath"
	"strings"
	"strconv"
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
)

const VULNERABILITY = "__vuln"
const COMPONENT = "__comp"
const JXRAY_BASR_URL = "https://jxray.jfrog.io/api/v1/updates/"
const ONBOARDING_URL = JXRAY_BASR_URL + "onboarding"
const BUNDLES_URL = JXRAY_BASR_URL + "bundles?from=%v&to=%v"

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
		if err := saveData(xrayTempDir, "vuln", zipSuffix, "", vulnerabilities); err != nil {
			return err
		}
	} else {
		log.Info("There aren't new vulnerabilities.")
	}

	if len(components) > 0 {
		log.Info("Downloading components...")
		if err := saveData(xrayTempDir, "comp", zipSuffix, "", components); err != nil {
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
		cliutils.CheckError(err)
		return
	}
	if from > to {
		err = errors.New("Invalid dates range.")
		cliutils.CheckError(err)
		return
	}
	return
}

func getXrayTempDir() (string, error) {
	tempDir := os.TempDir()
	xrayDir := tempDir + "/jfrog/xray/"
	if err := os.MkdirAll(xrayDir, 0777); err != nil {
		cliutils.CheckError(err)
		return "", nil
	}
	return xrayDir, nil
}

func saveData(xrsyTmpdir, filesPrefix, zipSuffix, logMsgPrefix string, urlsList []string) (err error) {
	dataDir, err := ioutil.TempDir(xrsyTmpdir, filesPrefix)
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
		log.Info(logMsgPrefix, "Downloading", url)
		httputils.DownloadFile(url, dataDir, fileName, httputils.HttpClientDetails{})
	}
	log.Info("Zipping files.")
	err = zipFolderFiles(dataDir, filesPrefix + zipSuffix + ".zip")
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
		cliutils.CheckError(err)
		return nil, nil, 0, err
	}
	if resp.StatusCode != 200 {
		err := errors.New("Response: " + resp.Status)
		cliutils.CheckError(err)
		return nil, nil, 0, err
	}
	var urls FilesList
	json.Unmarshal(body, &urls)
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

func zipFolderFiles(source, target string) (err error) {
	zipfile, err := os.Create(target)
	if err != nil {
		cliutils.CheckError(err)
		return
	}
	defer func() {
		if cerr := zipfile.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	archive := zip.NewWriter(zipfile)
	defer func() {
		if cerr := archive.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	filepath.Walk(source, func(path string, info os.FileInfo, err error) (currentErr error) {
		if info.IsDir() {
			return
		}

		if err != nil {
			currentErr = err
			return
		}

		header, currentErr := zip.FileInfoHeader(info)
		if currentErr != nil {
			cliutils.CheckError(currentErr)
			return
		}

		header.Method = zip.Deflate
		writer, currentErr := archive.CreateHeader(header)
		if currentErr != nil {
			cliutils.CheckError(currentErr)
			return
		}

		file, currentErr := os.Open(path)
		if currentErr != nil {
			cliutils.CheckError(currentErr)
			return
		}
		defer func() {
			if cerr := file.Close(); cerr != nil && currentErr == nil {
				currentErr = cerr
			}
		}()
		_, currentErr = io.Copy(writer, file)
		return
	})
	return
}

type OfflineUpdatesFlags struct {
	License string
	From    int64
	To      int64
	Version   string
}

type FilesList struct {
	Last_update int64
	Urls        []string
}