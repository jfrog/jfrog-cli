package commands

import (
    "github.com/jfrogdev/jfrog-cli-go/bintray/tests"
	"testing"
)

func TestShowAndCreateEntitlements(t *testing.T) {
    repo := CreateVersionDetailsForEntitlements("my-subject/my-repo")
    pkg := CreateVersionDetailsForEntitlements("my-subject/my-repo/my-pkg")
    version := CreateVersionDetailsForEntitlements("my-subject/my-repo/my-pkg/ver-1.9.1")

    bintrayDetails := tests.CreateBintrayDetails()

    url := BuildEntitlementsUrl(bintrayDetails, repo)
    expected := "https://api.bintray.com/repos/my-subject/my-repo/entitlements"
    if expected != url {
        t.Error("Got unexpected url from BuildEntitlementsUrl. Expected: " + expected + " Got " + url)
    }

    url = BuildEntitlementsUrl(bintrayDetails, pkg)
    expected = "https://api.bintray.com/packages/my-subject/my-repo/my-pkg/entitlements"
    if expected != url {
        t.Error("Got unexpected url from BuildEntitlementsUrl. Expected: " + expected + " Got " + url)
    }

    url = BuildEntitlementsUrl(bintrayDetails, version)
    expected = "https://api.bintray.com/packages/my-subject/my-repo/my-pkg/versions/ver-1.9.1/entitlements"
    if expected != url {
        t.Error("Got unexpected url from BuildEntitlementsUrl. Expected: " + expected + " Got " + url)
    }
}

func TestShowUpdateAndDeleteEntitlement(t *testing.T) {
    repo := CreateVersionDetailsForEntitlements("my-subject/my-repo")
    pkg := CreateVersionDetailsForEntitlements("my-subject/my-repo/my-pkg")
    version := CreateVersionDetailsForEntitlements("my-subject/my-repo/my-pkg/ver-1.9.1")

    bintrayDetails := tests.CreateBintrayDetails()

    url := BuildEntitlementUrl(bintrayDetails, repo, "my-ent-id")
    expected := "https://api.bintray.com/repos/my-subject/my-repo/entitlements/my-ent-id"
    if expected != url {
        t.Error("Got unexpected url from BuildEntitlementsUrl. Expected: " + expected + " Got " + url)
    }

    url = BuildEntitlementUrl(bintrayDetails, pkg, "my-ent-id")
    expected = "https://api.bintray.com/packages/my-subject/my-repo/my-pkg/entitlements/my-ent-id"
    if expected != url {
        t.Error("Got unexpected url from BuildEntitlementsUrl. Expected: " + expected + " Got " + url)
    }

    url = BuildEntitlementUrl(bintrayDetails, version, "my-ent-id")
    expected = "https://api.bintray.com/packages/my-subject/my-repo/my-pkg/versions/ver-1.9.1/entitlements/my-ent-id"
    if expected != url {
        t.Error("Got unexpected url from BuildEntitlementsUrl. Expected: " + expected + " Got " + url)
    }
}