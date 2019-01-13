package services

import (
	"encoding/json"
	"errors"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/missioncontrol/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/httpclient"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io/ioutil"
	"net/http"
	"os"
)

func AttachLic(service_name string, flags *AttachLicFlags) error {
	prepareLicenseFile(flags.LicensePath, flags.Override)
	postContent := utils.LicenseRequestContent{
		Name:             service_name,
		NumberOfLicenses: 1,
		Deploy:           flags.Deploy}
	requestContent, err := json.Marshal(postContent)
	if err != nil {
		return errorutils.CheckError(errors.New("Failed to marshal json. " + cliutils.GetDocumentationMessage()))
	}
	missionControlUrl := flags.MissionControlDetails.Url + "api/v3/attach_lic/buckets/" + flags.BucketId
	httpClientDetails := utils.GetMissionControlHttpClientDetails(flags.MissionControlDetails)
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		return err
	}
	resp, body, err := client.SendPost(missionControlUrl, requestContent, httpClientDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		if flags.LicensePath != "" {
			os.Remove(flags.LicensePath)
		}
		return errorutils.CheckError(errors.New(resp.Status + ". " + utils.ReadMissionControlHttpMessage(body)))
	}
	log.Debug("Mission Control response: " + resp.Status)

	if flags.LicensePath == "" {
		// Print response body to log
		log.Output(clientutils.IndentJson(body))
		return nil
	}

	// Extract license from response
	var licenseKey licenseKey
	err = json.Unmarshal(body, &licenseKey)
	err = errorutils.CheckError(err)
	if err != nil {
		return err
	}

	// Save license to file
	licenseKeyString := []byte(licenseKey.LicenseKey)
	err = saveLicense(flags.LicensePath, licenseKeyString)
	if err != nil {
		return err
	}

	return nil
}

func prepareLicenseFile(filepath string, overrideFile bool) (err error) {
	if filepath == "" {
		return
	}
	var dir bool
	dir, err = fileutils.IsDirExists(filepath, false)
	if err != nil {
		return
	}
	if dir {
		err = errorutils.CheckError(errors.New(filepath + " is a directory."))
		if err != nil {
			return
		}
	}
	var exists bool
	exists, err = fileutils.IsFileExists(filepath, false)
	if err != nil {
		return
	}
	if !overrideFile && exists {
		err = errorutils.CheckError(errors.New("File already exist, in case you wish to override the file use --override flag"))
		if err != nil {
			return
		}
	}
	_, directory := fileutils.GetFileAndDirFromPath(filepath)
	isPathExists := fileutils.IsPathExists(directory, false)
	if !isPathExists {
		os.MkdirAll(directory, 0700)
	}
	err = ioutil.WriteFile(filepath, nil, 0777)
	err = errorutils.CheckError(err)
	return
}

func saveLicense(filepath string, content []byte) (err error) {
	if filepath == "" {
		return
	}
	err = ioutil.WriteFile(filepath, content, 0777)
	err = errorutils.CheckError(err)
	return
}

type AttachLicFlags struct {
	MissionControlDetails *config.MissionControlDetails
	LicensePath           string
	BucketKey             string
	BucketId              string
	Override              bool
	Deploy                bool
}

type licenseKey struct {
	LicenseKey string `json:"license_key,omitempty"`
}
