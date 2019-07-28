package buildinfo

import (
	"reflect"
	"testing"
)

var envVars = map[string]string{"KeY": "key_val", "INClUdEd_VaR": "included_var", "EXCLUDED_pASSwoRd_var": "excluded_var"}

func TestIncludeAllPattern(t *testing.T) {
	includeFilter := createIncludeFilter("*")
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
	includeFilter := createIncludeFilter("*ED_V*;EXC*SwoRd_var")
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
	includeFilter := createIncludeFilter("*Ed_v*")
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
	excludeFilter := createExcludeFilter("*paSSword*;*seCrEt*;*kEy*;*token*")
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
