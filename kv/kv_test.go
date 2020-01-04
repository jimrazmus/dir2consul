package kv

import (
	"testing"
	"testing/quick"
)

func TestGetAndSet(t *testing.T) {
	k := NewKVList()
	err := quick.CheckEqual(k.Set, k.Get, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestIsEmpty(t *testing.T) {
	k := NewKVList()
	if k.IsEmpty() != true {
		t.Error("New KVList should be empty")
	}
	_, _, _ = k.Set("foo", "bar")
	if k.IsEmpty() == true {
		t.Error("KVList should NOT be empty")
	}
}
