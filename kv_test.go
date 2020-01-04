package main

import (
	"testing"
	"testing/quick"
)

func TestSet(t *testing.T) {
	k := NewKVList()
	err := quick.CheckEqual(k.Set, k.Get, nil)
	if err != nil {
		t.Error(err)
	}
}
