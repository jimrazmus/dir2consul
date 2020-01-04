// Simple Key Value library built on a map[string]string
package kv

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

// NewKVList returns a new KVList
func NewKVList() *KVList {
	k := &KVList{kvs: make(map[string]string)}
	return k
}

// Get returns the value associated with provided key or ErrNxKey
func (k *KVList) Get(key string, _ string) (string, string, error) {
	k.RLock()
	defer k.RUnlock()
	value, ok := k.kvs[key]
	if !ok {
		return key, "", ErrNxKey
	}
	return key, value, nil
}

// Keys returns a list of the keys
func (k *KVList) Keys() []string {
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
func (k *KVList) IsEmpty() bool {
	k.RLock()
	defer k.RUnlock()
	if k.kvs == nil || len(k.kvs) == 0 {
		return true
	}
	return false
}

// Set stores the value associated with key
func (k *KVList) Set(key string, value string) (string, string, error) {
	k.Lock()
	defer k.Unlock()
	k.kvs[key] = value
	return key, value, nil
}

// Values returns a list of the values
func (k *KVList) Values() []string {
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
