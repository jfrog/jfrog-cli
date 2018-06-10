package lru

import (
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	var getTests = []struct {
		name       string
		keyToAdd   string
		keyToGet   string
		expectedOk bool
	}{
		{"string_hit", "myKey", "myKey", true},
		{"string_miss", "myKey", "nonsense", false},
	}
	for _, tt := range getTests {
		c := New(0)
		c.Add(tt.keyToAdd, 1234)
		val, ok := c.Get(tt.keyToGet)
		if ok != tt.expectedOk {
			t.Fatalf("%s: cache hit = %v; want %v", tt.name, ok, !ok)
		} else if ok && val != 1234 {
			t.Fatalf("%s expected get to return 1234 but got %v", tt.name, val)
		}
	}
}

func TestEviction(t *testing.T) {
	c := New(3)
	c.Add("e1", true)
	c.Add("e2", true)
	c.Add("e3", false)
	c.Add("e4", false)

	_, ok := c.Get("e1")
	if ok {
		t.Fatal("Did not expect to find element e1 in cache after adding e4")
	}
	_, ok = c.Get("e2")
	if !ok {
		t.Fatal("Expected to find element e2 in cache after adding e4")
	}

	c.Add("e5", true)
	_, ok = c.Get("e2")
	if !ok {
		t.Fatal("Expected to find element e2 in cache after adding e5 because it was accessed recently")
	}
	_, ok = c.Get("e3")
	if ok {
		t.Fatal("Did not expect to find element e3 in cache after adding e5 since e2 was accessed before it")
	}
}

func TestRemove(t *testing.T) {
	c := New(0)
	c.Add("myKey", 1234)
	if val, ok := c.Get("myKey"); !ok {
		t.Fatal("TestRemove returned no match")
	} else if val != 1234 {
		t.Fatalf("TestRemove failed.  Expected %d, got %v", 1234, val)
	}

	c.Remove("myKey")
	if _, ok := c.Get("myKey"); ok {
		t.Fatal("TestRemove returned a removed entry")
	}
}

func TestPurge(t *testing.T) {
	c := New(2)
	l := c.Len()
	if l != 0 {
		t.Fatalf("Expected length to be 1 but got %d", l)
	}
	c.Add("e1", 1)
	l = c.Len()
	if l != 1 {
		t.Fatalf("Expected length to be 1 but got %d", l)
	}
	c.Add("e2", 2)
	l = c.Len()
	if l != 2 {
		t.Fatalf("Expected length to be 2 but got %d", l)
	}
	c.Add("e3", 3)
	l = c.Len()
	if l != 2 {
		t.Fatalf("Expected length to be 2 but got %d", l)
	}
	if _, ok := c.Get("e1"); ok {
		t.Fatal("Expected not to get value for e1 but it was not found")
	}
	if _, ok := c.Get("e2"); !ok {
		t.Fatal("Expected to get value for e2 but it was not found")
	}
	if _, ok := c.Get("e3"); !ok {
		t.Fatal("Expected to get value for e2 but it was not found")
	}

	c.Clear()
	l = c.Len()
	if _, ok := c.Get("e2"); ok {
		t.Fatal("Expected not to get value for e2 but it was found")
	}
	if _, ok := c.Get("e3"); ok {
		t.Fatal("Expected not to get value for e3 but it was found")
	}
	if l != 0 {
		t.Fatalf("Expected length to be 0 after clearing cache, but got %d", l)
	}
}

func TestExpiry(t *testing.T) {
	c := New(3, WithExpiry(100*time.Millisecond))
	c.Add("e1", 1)
	c.Add("e2", 2)
	time.Sleep(50 * time.Millisecond)
	c.Add("e3", 3)
	if _, ok := c.Get("e1"); !ok {
		t.Fatal("Expected to get value for e1 but it was not found")
	}
	if _, ok := c.Get("e2"); !ok {
		t.Fatal("Expected to get value for e2 but it was not found")
	}
	if _, ok := c.Get("e3"); !ok {
		t.Fatal("Expected to get value for e3 but it was not found")
	}
	l := c.Len()
	if l != 3 {
		t.Fatalf("Expected length to be 3 but got %d", l)
	}
	time.Sleep(60 * time.Millisecond)
	if _, ok := c.Get("e1"); ok {
		t.Fatal("Expected not to get value for e2 but it was found")
	}
	if _, ok := c.Get("e2"); ok {
		t.Fatal("Expected not to get value for e3 but it was found")
	}
	if _, ok := c.Get("e3"); !ok {
		t.Fatal("Expected to get value for e3 but it was not found")
	}
	l = c.Len()
	if l != 1 {
		t.Fatalf("Expected length to be 1 but got %d", l)
	}
}
