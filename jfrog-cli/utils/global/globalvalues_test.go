package global

import (
	"math/rand"
	"testing"
)

func TestSuccessValues(t *testing.T) {
	global1 := GetGlobalVariables()
	var success int
	global1.IncreaseSuccess()
	success++
	global2 := GetGlobalVariables()
	if global2.GetSuccess() != success {
		t.Error("Expected to get", success, ", got:", global2.GetSuccess())
	}

	global2.IncreaseSuccess()
	success++
	global2.IncreaseSuccess()
	success++
	global2.IncreaseSuccess()
	success++

	if global1.GetSuccess() != success {
		t.Error("Expected to get", success, ", got:", global2.GetSuccess())
	}

	randValue := rand.Intn(100)

	i := 0
	for i < randValue {
		global1.IncreaseSuccess()
		success++
		i++
	}

	if global2.GetSuccess() != success {
		t.Error("Expected to get", success, ", got:", global2.GetSuccess())
	}
}

func TestFailureValues(t *testing.T) {
	global1 := GetGlobalVariables()
	var failure int
	global1.IncreaseFailures()
	failure++
	global2 := GetGlobalVariables()
	if global2.GetFailures() != failure {
		t.Error("Expected to get", failure, ", got:", global2.GetFailures())
	}

	global2.IncreaseFailures()
	failure++
	global2.IncreaseFailures()
	failure++
	global2.IncreaseFailures()
	failure++

	if global1.GetFailures() != failure {
		t.Error("Expected to get", failure, ", got:", global2.GetFailures())
	}

	randValue := rand.Intn(100)

	i := 0
	for i < randValue {
		global1.IncreaseFailures()
		failure++
		i++
	}

	if global2.GetFailures() != failure {
		t.Error("Expected to get", failure, ", got:", global2.GetFailures())
	}
}

func TestTotalValue(t *testing.T) {
	global1 := GetGlobalVariables()
	var total int
	global1.IncreaseTotal(1)
	total++
	global2 := GetGlobalVariables()
	if global2.GetTotal() != total {
		t.Error("Expected to get", total, ", got:", global2.GetTotal())
	}

	global2.IncreaseTotal(1)
	total++
	global2.IncreaseTotal(1)
	total++
	global2.IncreaseTotal(1)
	total++

	if global1.GetTotal() != total {
		t.Error("Expected to get", total, ", got:", global2.GetTotal())
	}

	randValue := rand.Intn(100)
	global3 := GetGlobalVariables()
	global3.IncreaseTotal(randValue)
	total += randValue
	if global2.GetTotal() != total {
		t.Error("Expected to get", total, ", got:", global2.GetFailures())
	}
}

func TestGlobalMap(t *testing.T) {
	global1 := GetGlobalVariables()
	globalMap := global1.GetGlobalMap()
	globalMap["key"] = false

	global2 := GetGlobalVariables()
	globalMap2 := global2.GetGlobalMap()
	value, ok := globalMap2["key"]
	if !ok {
		t.Error("Expected key were not found in second map")
	}

	if value {
		t.Error("Expected value false but got true for the second map")
	}

	globalMap2["key"] = true
	value, ok = globalMap["key"]
	if !ok {
		t.Error("Expected key were not found in first map")
	}

	if !value {
		t.Error("Expected value true but got false")
	}
}
