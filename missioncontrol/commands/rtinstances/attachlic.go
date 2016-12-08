package rtinstances

import (
	"github.com/jfrogdev/jfrog-cli-go/missioncontrol/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"encoding/json"
	"io/ioutil"
	"errors"
	"fmt"
	"os"
)

func AttachLic(instanceName string, flags *AttachLicFlags) error {
	prepareLicenseFile(flags.LicensePath, flags.Override)
	postContent := utils.LicenseRequestContent{
		Name: 	  	 instanceName,
		NodeID:	     flags.NodeId,
		Deploy:	     flags.Deploy}
	requestContent, err := json.Marshal(postContent)
	if err != nil {
		return cliutils.CheckError(errors.New("Failed to marshal json. " + cliutils.GetDocumentationMessage()))
	}
	missionControlUrl := flags.MissionControlDetails.Url + "api/v1/buckets/" + flags.BucketId + "/licenses";
	httpClientDetails := utils.GetMissionControlHttpClientDetails(flags.MissionControlDetails)
	resp, body, err := ioutils.SendPost(missionControlUrl, requestContent, httpClientDetails)
    if err != nil {
        return err
    }
	if resp.StatusCode != 200 {
		if flags.LicensePath != "" {
			os.Remove(flags.LicensePath)
		}
		return cliutils.CheckError(errors.New(resp.Status + ". " + utils.ReadMissionControlHttpMessage(body)))
	}
	fmt.Println("Mission Control response: " + resp.Status)
	if flags.LicensePath == "" {
	    var m Message
	    m, err = extractJsonValue(body)
		if err != nil {
		    return err
		}
		requestContent, err = json.Marshal(m)
		err = cliutils.CheckError(err)
		if err != nil {
		    return err
		}
		fmt.Println(string(requestContent))
	} else {
	    var licenseKey []byte
		licenseKey, err = getLicenseFromJson(body)
		if err != nil {
		    return err
		}
		err = saveLicense(flags.LicensePath, licenseKey)
	}
	return nil
}

func getLicenseFromJson(body []byte) (licenseKey []byte, err error) {
    var m Message
    m, err = extractJsonValue(body)
    if err != nil {
        return
    }
	licenseKey = []byte(m.LicenseKey)
	return
}

func extractJsonValue(body []byte) (m Message, err error) {
	data := &Data{}
	err = json.Unmarshal(body, &data);
	err = cliutils.CheckError(err)
	if err != nil {
	    return
	}
	m = data.Data
	return
}

func prepareLicenseFile(filepath string, overrideFile bool) (err error) {
	if filepath == "" {
		return
	}
	var dir bool
	dir, err = ioutils.IsDir(filepath)
	if err != nil {
	    return
	}
	if dir {
		err = cliutils.CheckError(errors.New(filepath + " is a directory."))
        if err != nil {
            return
        }
	}
	var exists bool
	exists, err = ioutils.IsFileExists(filepath)
	if err != nil {
	    return
	}
	if !overrideFile && exists {
		err = cliutils.CheckError(errors.New("File already exist, in case you wish to override the file use --override flag"))
        if err != nil {
            return
        }
	}
	_, directory := ioutils.GetFileAndDirFromPath(filepath)
	isPathExists := ioutils.IsPathExists(directory)
	if !isPathExists {
		os.MkdirAll(directory, 0700)
	}
	err = ioutil.WriteFile(filepath, nil, 0777)
	err = cliutils.CheckError(err)
	return
}

func saveLicense(filepath string, content []byte) (err error) {
	if filepath == "" {
		return
	}
	err = ioutil.WriteFile(filepath, content, 0777)
	err = cliutils.CheckError(err)
	return
}

type AttachLicFlags struct {
	MissionControlDetails *config.MissionControlDetails
	LicensePath 	      string
	NodeId 			      string
	BucketKey 			  string
	BucketId 			  string
	Override 			  bool
	Deploy 			  	  bool
}

type Message struct {
	LicenseKey string `json:"licenseKey,omitempty"`
}

type Data struct {
	Data Message
}
