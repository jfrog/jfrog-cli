package main

import (
	"github.com/jfrog/jfrog-cli-core/v2/xray/scangraph"
	"github.com/jfrog/jfrog-cli/utils/tests"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"testing"
)

func initXscTest(t *testing.T, minVersion string) {
	if !*tests.TestXsc {
		t.Skip("Skipping Xsc test. To run xsc test add the '-test.xsc=true' option.")
	}
	validateXscVersion(t, minVersion)
}
func validateXscVersion(t *testing.T, minVersion string) {
	err := clientutils.ValidateMinimumVersion(clientutils.Xray, xrayDetails.XscVersion, minVersion)
	if err != nil {
		t.Skip(err)
	}
}
func TestXSCAudit(t *testing.T) {
	initXscTest(t, scangraph.XscMinVersion)
	testXrayAuditNpm(t, "json", true)
}
