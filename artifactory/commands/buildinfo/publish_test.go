package buildinfo

import (
	"reflect"
	"testing"

	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
)

var envVars = map[string]string{"KeY": "key_val", "INClUdEd_VaR": "included_var", "EXCLUDED_pASSwoRd_var": "excluded_var"}

func TestIncludeAllPattern(t *testing.T) {
	conf := buildinfo.Configuration{EnvInclude: "*"}
	includeFilter := conf.IncludeFilter()
	filteredKeys, err := includeFilter(envVars)
	if err != nil {
		t.Error(err)
	}

	equals := reflect.DeepEqual(envVars, filteredKeys)
	if !equals {
		t.Error("expected:", envVars, "got:", filteredKeys)
	}
}

func TestIncludePartial(t *testing.T) {
	conf := buildinfo.Configuration{EnvInclude: "*ED_V*;EXC*SwoRd_var"}
	includeFilter := conf.IncludeFilter()
	filteredKeys, err := includeFilter(envVars)
	if err != nil {
		t.Error(err)
	}

	expected := map[string]string{"INClUdEd_VaR": "included_var", "EXCLUDED_pASSwoRd_var": "excluded_var"}
	equals := reflect.DeepEqual(expected, filteredKeys)
	if !equals {
		t.Error("expected:", expected, "got:", filteredKeys)
	}
}

func TestIncludePartialIgnoreCase(t *testing.T) {
	conf := buildinfo.Configuration{EnvInclude: "*Ed_v*"}
	includeFilter := conf.IncludeFilter()
	filteredKeys, err := includeFilter(envVars)
	if err != nil {
		t.Error(err)
	}

	expected := map[string]string{"INClUdEd_VaR": "included_var"}
	equals := reflect.DeepEqual(expected, filteredKeys)
	if !equals {
		t.Error("expected:", expected, "got:", filteredKeys)
	}
}

func TestExcludePasswordsPattern(t *testing.T) {
	conf := buildinfo.Configuration{EnvExclude: "*paSSword*;*PsW*;*seCrEt*;*kEy*;*token*"}
	excludeFilter := conf.ExcludeFilter()
	filteredKeys, err := excludeFilter(envVars)
	if err != nil {
		t.Error(err)
	}

	expected := map[string]string{"INClUdEd_VaR": "included_var"}
	equals := reflect.DeepEqual(expected, filteredKeys)
	if !equals {
		t.Error("expected:", expected, "got:", filteredKeys)
	}
}
