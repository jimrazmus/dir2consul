// Package kv provides a simple key-value library
package kv

import (
	"errors"
	"sync"
)

// ErrNxKey error is returned when key does not exist
var ErrNxKey = errors.New("key does not exist")

// List provides key-value storage and a sync mutex
type List struct {
	sync.RWMutex
	kvs map[string][]byte
}

// NewList returns a new List
func NewList() *List {
	k := &List{kvs: make(map[string][]byte)}
	return k
}

// Get returns the value associated with provided key or ErrNxKey
func (k *List) Get(key string, _ []byte) (string, []byte, error) {
	k.RLock()
	defer k.RUnlock()
	value, ok := k.kvs[key]
	if !ok {
		return key, nil, ErrNxKey
	}
	return key, value, nil
}

// Keys returns a list of the keys
func (k *List) Keys() []string {
	k.RLock()
	defer k.RUnlock()
	keys := make([]string, len(k.kvs))
	i := 0
	for k := range k.kvs {
		keys[i] = k
		i++
	}
	return keys
}

// IsEmpty returns true if k is empty
func (k *List) IsEmpty() bool {
	k.RLock()
	defer k.RUnlock()
	if k.kvs == nil || len(k.kvs) == 0 {
		return true
	}
	return false
}

// Set stores the value associated with key
func (k *List) Set(key string, value []byte) (string, []byte, error) {
	k.Lock()
	defer k.Unlock()
	k.kvs[key] = value
	return key, value, nil
}

// Values returns a list of the values
func (k *List) Values() [][]byte {
	k.RLock()
	defer k.RUnlock()
	values := make([][]byte, len(k.kvs))
	i := 0
	for _, v := range k.kvs {
		values[i] = v
		i++
	}
	return values
}
