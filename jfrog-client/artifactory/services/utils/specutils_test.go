package utils

import (
	"encoding/json"
	"testing"
)

func TestAqlUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		itemsFind string
		expected  string
	}{
		{"test1", "{\"items.find\":{}}", "{}"},
		{"test2", "{\"items.find\": {}}", "{}"},
		{"test3", "  {  \"items.find\"\n  :    {}  }  ", "{}  "},
		{"test4", "  {  \"items.find\"\n  :    {  \"inside\"  :  \"something\"  }  }  ", "{  \"inside\"  :  \"something\"  }  "},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			aql := &Aql{}
			err := json.Unmarshal([]byte(test.itemsFind), aql)
			if err != nil {
				t.Error(err)
			}
			if aql.ItemsFind != test.expected {
				t.Error("Test:", test.name, "Expected:", "'"+test.expected+"'", "got:", "'"+aql.ItemsFind+"'")
			}
		})
	}
}
