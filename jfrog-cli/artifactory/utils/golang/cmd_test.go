package golang

import (
	"reflect"
	"testing"
)

func TestOutputToSlice(t *testing.T) {
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
	expected := []string{
		"github.com/dsnet/compress@v0.0.0-20171208185109-cc9eb1d7ad76",
		"github.com/golang/snappy@v0.0.0-20180518054509-2e65f85255db",
		"github.com/mholt/archiver@v2.1.0+incompatible",
		"github.com/nwaples/rardecode@v0.0.0-20171029023500-e06696f847ae",
		"github.com/pierrec/lz4@v2.0.5+incompatible",
		"github.com/ulikunitz/xz@v0.5.4",
		"golang.org/x/text@v0.3.1-0.20180807135948-17ff2d5776d2",
		"golang.org/x/tools@v0.0.0-20181006002542-f60d9635b16a",
		"rsc.io/quote@v1.5.2",
		"rsc.io/sampler@v1.3.0",
		"golang.org/x/text@v0.0.0-20170915032832-14c0d48ead0c",
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expecting: \n%s \nGot: \n%s", expected, actual)
	}
}
