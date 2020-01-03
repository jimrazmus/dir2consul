// Simple Key Value library built on a map[string]string
package main

import (
	"errors"
	"sync"
)

// ErrNxKey error is returned when key does not exist
var ErrNxKey = errors.New("key does not exist")

// KVList struct provides key-value storage and a sync mutex
type KVList struct {
	sync.RWMutex
	kvs map[string]string
}

// New returns an new KVList
func New() KVList {
	k := KVList{kvs: make(map[string]string)}
	return k
}

// Get returns the value associated with provided key or ErrNxKey
func (k KVList) Get(key string) (string, error) {
	k.RLock()
	defer k.RUnlock()
	value, ok := k.kvs[key]
	if !ok {
		return "", ErrNxKey
	}
	return value, nil
}

func (k KVList) Keys() []string {
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

func (k KVList) IsEmpty() bool {
	k.RLock()
	defer k.RUnlock()
	if k.kvs == nil || len(k.kvs) == 0 {
		return true
	}
	return false
}

func (k KVList) Set(key string, value string) {
	k.Lock()
	defer k.Unlock()
	k.kvs[key] = value
}

func (k KVList) Values() []string {
	k.RLock()
	defer k.RUnlock()
	values := make([]string, len(k.kvs))
	i := 0
	for _, v := range k.kvs {
		values[i] = v
		i++
	}
	return values
}
