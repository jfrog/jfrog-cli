package cliutils

import (
	"strconv"
	"testing"
)

type testStruct struct {
	AStringField string
	ABoolField   bool
	AnIntField   int
}

func TestSetStructField(t *testing.T) {
	tests := []struct {
		testName  string
		fieldName string
		value     string
		fieldType string
	}{
		{"SetStringTest", "AStringField", "value", "string"},
		{"SetBoolTest", "ABoolField", "true", "bool"},
		{"SetIntTest", "AnIntField", "17", "int"},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			aStruct := new(testStruct)
			err := SetStructField(aStruct, test.fieldName, test.value)
			if err != nil {
				t.Error(err)
			}
			switch test.fieldType {
			case "string":
				if aStruct.AStringField != test.value {
					t.Errorf("Expected string field to be setted to %s and not to %s", test.value, aStruct.AStringField)
				}
			case "bool":
				boolVal, _ := strconv.ParseBool(test.value)
				if aStruct.ABoolField != boolVal {
					t.Errorf("Expected string field to be setted to %s and not to %s", test.value, aStruct.AStringField)
				}
			case "int":
				intVal, _ := strconv.ParseInt(test.value, 10, 32)
				if aStruct.AnIntField != int(intVal) {
					t.Errorf("Expected string field to be setted to %s and not to %s", test.value, aStruct.AStringField)
				}
			}

		})
	}
}
