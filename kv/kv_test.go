package kv

import (
	"testing"
	"testing/quick"
)

func TestGetAndSet(t *testing.T) {
	k := NewList()
	err := quick.CheckEqual(k.Set, k.Get, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestIsEmpty(t *testing.T) {
	k := NewList()
	if k.IsEmpty() != true {
		t.Error("New KVList should be empty")
	}
	_, _, _ = k.Set("foo", []byte("bar"))
	if k.IsEmpty() == true {
		t.Error("KVList should NOT be empty")
	}
}

func TestKeys(t *testing.T) {
	k := NewList()
	keys := k.Keys()
	if len(keys) != 0 {
		t.Error("ERROR")
	}
	_, _, _ = k.Set("foo", []byte("bar"))
	_, _, _ = k.Set("goo", []byte("bar"))
	_, _, _ = k.Set("hoo", []byte("bar"))
	keys = k.Keys()
	if len(keys) != 3 {
		t.Error("ERROR")
	}
}

func TestValues(t *testing.T) {
	k := NewList()
	values := k.Values()
	if len(values) != 0 {
		t.Error("ERROR")
	}
	_, _, _ = k.Set("foo", []byte("bar"))
	_, _, _ = k.Set("goo", []byte("bar"))
	_, _, _ = k.Set("hoo", []byte("bar"))
	values = k.Values()
	if len(values) != 3 {
		t.Error("ERROR")
	}
}
