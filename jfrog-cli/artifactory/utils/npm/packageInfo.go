package npm

import (
	"encoding/json"
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type PackageInfo struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
	Scope   string
}

func ReadPackageInfoFromPackageJson(packageJsonDirectory string) (*PackageInfo, error) {
	log.Debug("Reading info from package.json file:", filepath.Join(packageJsonDirectory, "package.json"))
	packageJson, err := ioutil.ReadFile(filepath.Join(packageJsonDirectory, "package.json"))
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	return ReadPackageInfo(packageJson)
}

func ReadPackageInfo(data []byte) (*PackageInfo, error) {
	parsedResult := new(PackageInfo)
	if err := json.Unmarshal(data, parsedResult); err != nil {
		return nil, errorutils.CheckError(err)
	}
	return splitScopeFromName(parsedResult), nil
}

func (pi *PackageInfo) BuildInfoModuleId() string {
	nameBase := fmt.Sprintf("%s:%s", pi.Name, pi.Version)
	if pi.Scope == "" {
		return nameBase
	}
	return fmt.Sprintf("%s:%s", strings.TrimPrefix(pi.Scope, "@"), nameBase)
}

func (pi *PackageInfo) GetDeployPath() string {
	fileName := fmt.Sprintf("%s-%s.tgz", pi.Name, pi.Version)
	if pi.Scope == "" {
		return fmt.Sprintf("%s/-/%s", pi.Name, fileName)
	}
	return fmt.Sprintf("%s/%s/-/%s/%s", pi.Scope, pi.Name, pi.Scope, fileName)
}

func (pi *PackageInfo) GetExpectedPackedFileName() string {
	fileNameBase := fmt.Sprintf("%s-%s.tgz", pi.Name, pi.Version)
	if pi.Scope == "" {
		return fileNameBase
	}
	return fmt.Sprintf("%s-%s", strings.TrimPrefix(pi.Scope, "@"), fileNameBase)
}

func splitScopeFromName(packageInfo *PackageInfo) *PackageInfo {
	if strings.HasPrefix(packageInfo.Name, "@") && strings.Contains(packageInfo.Name, "/") {
		splitValues := strings.Split(packageInfo.Name, "/")
		packageInfo.Scope = splitValues[0]
		packageInfo.Name = splitValues[1]
	}
	return packageInfo
}
