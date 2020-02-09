package main

import "os"

// getenv returns the environment value for the given key or the default value when not found
func getenv(key string, _default string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return _default
	}
	return val
}
