package golang

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"github.com/magiconair/properties/assert"
)

func TestOutputToMap(t *testing.T) {
	content := `go: finding github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db
go: finding github.com/nwaples/rardecode v0.0.0-20171029023500-e06696f847ae
go: finding github.com/pierrec/lz4 v2.0.5+incompatible
go: finding github.com/ulikunitz/xz v0.5.4
go: finding github.com/dsnet/compress v0.0.0-20171208185109-cc9eb1d7ad76
go: finding github.com/mholt/archiver v2.1.0+incompatible
go: finding rsc.io/quote v1.5.2
go: finding golang.org/x/tools v0.0.0-20181006002542-f60d9635b16a
go: finding golang.org/x/text v0.3.1-0.20180807135948-17ff2d5776d2
go: finding rsc.io/sampler v1.3.0
go: finding golang.org/x/text v0.0.0-20170915032832-14c0d48ead0c
github.com/you/hello github.com/dsnet/compress@v0.0.0-20171208185109-cc9eb1d7ad76
github.com/you/hello github.com/golang/snappy@v0.0.0-20180518054509-2e65f85255db
github.com/you/hello github.com/mholt/archiver@v2.1.0+incompatible
github.com/you/hello github.com/nwaples/rardecode@v0.0.0-20171029023500-e06696f847ae
github.com/you/hello github.com/pierrec/lz4@v2.0.5+incompatible
github.com/you/hello github.com/ulikunitz/xz@v0.5.4
github.com/you/hello golang.org/x/text@v0.3.1-0.20180807135948-17ff2d5776d2
github.com/you/hello golang.org/x/tools@v0.0.0-20181006002542-f60d9635b16a
github.com/you/hello rsc.io/quote@v1.5.2
rsc.io/quote@v1.5.2 rsc.io/sampler@v1.3.0
rsc.io/sampler@v1.3.0 golang.org/x/text@v0.0.0-20170915032832-14c0d48ead0c
	`

	actual := outputToMap(content)
	expected := map[string]bool{
		"github.com/dsnet/compress@v0.0.0-20171208185109-cc9eb1d7ad76":    true,
		"github.com/golang/snappy@v0.0.0-20180518054509-2e65f85255db":     true,
		"github.com/mholt/archiver@v2.1.0+incompatible":                   true,
		"github.com/nwaples/rardecode@v0.0.0-20171029023500-e06696f847ae": true,
		"github.com/pierrec/lz4@v2.0.5+incompatible":                      true,
		"github.com/ulikunitz/xz@v0.5.4":                                  true,
		"golang.org/x/text@v0.3.1-0.20180807135948-17ff2d5776d2":          true,
		"golang.org/x/tools@v0.0.0-20181006002542-f60d9635b16a":           true,
		"rsc.io/quote@v1.5.2":                                             true,
		"rsc.io/sampler@v1.3.0":                                           true,
		"golang.org/x/text@v0.0.0-20170915032832-14c0d48ead0c":            true,
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expecting: \n%v \nGot: \n%v", expected, actual)
	}
}

func TestGetProjectDir(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	defer os.Chdir(wd)

	// CD into a directory with a go.mod file.
	projectRoot := filepath.Join("testdata", "project")
	err = os.Chdir(projectRoot)
	if err != nil {
		t.Error(err)
	}

	// Make projectRoot an absolute path.
	projectRoot, err = os.Getwd()
	if err != nil {
		t.Error(err)
	}

	// Get the project root.
	root, err := GetProjectRoot()
	if err != nil {
		t.Error(err)
	}
	if root != projectRoot {
		t.Error("Expecting", projectRoot, "got:", root)
	}

	// CD back to the original directory.
	if err := os.Chdir(wd); err != nil {
		t.Error(err)
	}

	// CD into a sub directory in the same project, and expect to get the same project root.
	os.Chdir(wd)
	projectSubDirectory := filepath.Join("testdata", "project", "dir")
	err = os.Chdir(projectSubDirectory)
	if err != nil {
		t.Error(err)
	}
	root, err = GetProjectRoot()
	if err != nil {
		t.Error(err)
	}
	if root != projectRoot {
		t.Error("Expecting", projectRoot, "got:", root)
	}

	// CD back to the original directory.
	if err := os.Chdir(wd); err != nil {
		t.Error(err)
	}

	// Now CD into a directory outside the project, and expect to get a different project root.
	noProjectRoot := filepath.Join("testdata", "noproject")
	err = os.Chdir(noProjectRoot)
	if err != nil {
		t.Error(err)
	}
	root, err = GetProjectRoot()
	if err != nil {
		t.Error(err)
	}
	if root == projectRoot {
		t.Error("Expecting a different value than", root)
}
