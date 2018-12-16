package golang

import (
	"math/rand"
	"testing"
)

func TestSuccessValues(t *testing.T) {
	cache := DependenciesCache{}
	var success int
	cache.IncrementSuccess()
	success++
	cache.IncrementSuccess()
	success++
	cache.IncrementSuccess()
	success++
	if cache.GetSuccesses() != success {
		t.Error("Expected to get", success, ", got:", cache.GetSuccesses())
	}
}

func TestFailureValues(t *testing.T) {
	cache := DependenciesCache{}
	var failures int
	cache.IncrementFailures()
	failures++
	cache.IncrementFailures()
	failures++
	cache.IncrementFailures()
	failures++
	if cache.GetFailures() != failures {
		t.Error("Expected to get", failures, ", got:", cache.GetSuccesses())
	}
}

func TestTotalValue(t *testing.T) {
	cache := DependenciesCache{}
	var total int
	cache.IncrementTotal(1)
	total++
	if cache.GetTotal() != total {
		t.Error("Expected to get", total, ", got:", cache.GetTotal())
	}

	cache.IncrementTotal(1)
	total++
	cache.IncrementTotal(1)
	total++
	cache.IncrementTotal(1)
	total++

	if cache.GetTotal() != total {
		t.Error("Expected to get", total, ", got:", cache.GetTotal())
	}

	randValue := rand.Intn(100)
	cache.IncrementTotal(randValue)
	total += randValue
	if cache.GetTotal() != total {
		t.Error("Expected to get", total, ", got:", cache.GetTotal())
	}
}
