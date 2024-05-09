package main

import (
	"errors"
	"sync"
)

// ----Store operations (TODO: move to store.go)
// Store will start as a simple map of string:string (Later --> struct with mutex )
// Not valid for concurrency-->var store = make(map[string]string)
// Custom map struct with locks
type lockableMap struct {
	sync.RWMutex //Allow to lock map for write or read
	m            map[string]string
}

// Store is just a map with locking capabilites ( of type lockableMap)
var store = lockableMap{m: make(map[string]string)}

// Define Custom error
var ErrorNoSuchKey = errors.New("no Such Key")

// Define Put function to add elements to store
func Put(key, value string) error {
	// Writing lock
	store.Lock()
	//Always unlock on func exit
	defer store.Unlock()
	//Idempotent so no checks for already existing value with given key ( ALWAYS OVERWRITE)
	store.m[key] = value

	return nil
}

// Define Get function to retrieve element by key
func Get(key string) (string, error) {
	// Read Lock
	store.RLock()
	defer store.RUnlock()
	value, ok := store.m[key]

	if !ok {
		return "", ErrorNoSuchKey
	}

	return value, nil
}

// Define Delete function to delete element by key
func Delete(key string) {
	store.Lock()
	defer store.Unlock()
	delete(store.m, key)

}
