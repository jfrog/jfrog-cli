package dependencies

import (
	"github.com/gomods/athens/pkg/gomod/bytes"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"reflect"
	"testing"
)

func TestParseListOutput(t *testing.T) {
	content := []byte(`MODULE                  VERSION
github.com/vgo-example  -
golang.org/x/text       v0.0.0-20170915032832-14c0d48ead0c
rsc.io/quote            v1.5.2
rsc.io/sampler          v1.3.0
	`)

	actual, err := parseListOutput(content)
	if err != nil {
		t.Error(err)
	}
	expected := map[string]string{
		"golang.org/x/text": "v0.0.0-20170915032832-14c0d48ead0c",
		"rsc.io/quote":      "v1.5.2",
		"rsc.io/sampler":    "v1.3.0",
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expecting: \n%s \nGot: \n%s", expected, actual)
	}
}

func TestGomod(t *testing.T) {
	content := []byte(`module "github.com/romang/vgo-example"

require "rsc.io/quote" v1.5.2
	`)
	parser := bytes.NewContentParser(content)

	log.Output(parser.ModuleName())

}
