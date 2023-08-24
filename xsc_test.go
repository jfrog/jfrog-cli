package main

import (
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/v2/xray/commands/utils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"testing"
)

func initXscTest(t *testing.T, minVersion string) {
	if !*tests.TestXsc {
		t.Skip("Skipping Xsc test. To run xsc test add the '-test.xsc=true' option.")
	}
	*tests.TestXray = true
	validateXscVersion(t, minVersion)
}
func validateXscVersion(t *testing.T, minVersion string) {
	err := coreutils.ValidateMinimumVersion(coreutils.Xsc, xrayDetails.XscVersion, minVersion)
	if err != nil {
		t.Skip(err)
	}
}
func TestXSCAudit(t *testing.T) {
	initXscTest(t, utils.XscMinVersion)
	testXrayAuditNpm(t, "json")
}
