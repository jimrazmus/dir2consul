package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/jimrazmus/dir2consul/kv"
)

// conditionally compile in or out the debug prints
const debug = true

// ConsulKeyPrefix is the path prefix to prepend to all consul keys
var ConsulKeyPrefix = getenv("D2C_CONSUL_KEY_PREFIX", "")

// ConsulServerURL is the URL of the Nomad server that will handle job submissions
var ConsulServerURL = getenv("D2C_CONSUL_SERVER", "http://localhost:8500")

// Directory is the directory we should walk
var Directory = getenv("D2C_DIRECTORY", "local")

// IgnoreDirs is a comma delimited list of directories to ignore when walking the files
var IgnoreDirs = strings.Split(getenv("D2C_IGNORE_DIRS", ".git"), ",")

// IgnoreTypes is a comma delimited list of file suffixes to ignore when walking the files
var IgnoreTypes = strings.Split(getenv("D2C_IGNORE_TYPES", ""), ",")

// VaultToken is the token value used to access the Consul server
var VaultToken = getenv("VAULT_TOKEN", "")

func main() {
	log.Println("dir2consul starting with configuration:")
	log.Println("D2C_CONSUL_KEY_PREFIX:", ConsulKeyPrefix)
	log.Println("D2C_CONSUL_SERVER:", ConsulServerURL)
	log.Println("D2C_DIRECTORY:", Directory)
	log.Println("D2C_IGNORE_DIRS:", IgnoreDirs)
	log.Println("D2C_IGNORE_TYPES:", IgnoreTypes)

	os.Chdir(Directory)

	// GO Get KVs from Files
	fileKeyValues := kv.NewList()
	err := LoadKeyValuesFromDisk(fileKeyValues)
	if err != nil {
		log.Fatal(err)
	}
	if debug {
		fileKeys := fileKeyValues.Keys()
		log.Println("fileKeys:", fileKeys)
	}

	// GO Get KVs from Consul
	consulKeyValues := kv.NewList()
	err = LoadKeyValuesFromConsul(consulKeyValues)
	if err != nil {
		log.Fatal(err)
	}
	if debug {
		consulKeys := consulKeyValues.Keys()
		log.Println("consulKeys:", consulKeys)
	}

	// Add or update data in Consul
	for _, key := range fileKeyValues.Keys() {
		_, fb, _ := fileKeyValues.Get(key, nil)
		_, cb, _ := consulKeyValues.Get(key, nil)
		if bytes.Compare(fb, cb) != 0 {
			log.Println("Upsert:", key)
		}
	}

	// Delete extra data from Consul
	for _, key := range consulKeyValues.Keys() {
		// TBD
	}

}

// LoadKeyValuesFromConsul queries Consul and loads the results into a kv.List
func LoadKeyValuesFromConsul(kv *kv.List) error {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return err
	}
	consulKv := client.KV()
	kvPairs, _, err := consulKv.List(ConsulKeyPrefix, nil)
	if err != nil {
		return err
	}
	for _, kvPair := range kvPairs {
		kv.Set(kvPair.Key, kvPair.Value)
	}
	return nil
}

// LoadKeyValuesFromDisk walks the file system and loads file contents into a kv.List
func LoadKeyValuesFromDisk(kv *kv.List) error {
	return filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Mode().IsDir() && ignoreDir(path, IgnoreDirs) {
			return filepath.SkipDir
		}

		if info.Mode().IsDir() || !info.Mode().IsRegular() || ignoreFile(path, IgnoreTypes) {
			return nil
		}

		if debug {
			log.Println("path:", path)
		}

		elemKey := strings.TrimSuffix(path, filepath.Ext(path))
		elemVal, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		switch strings.ToLower(filepath.Ext(path)) {
		case ".hcl":
			return loadHclFile()
		case ".ini":
			return loadIniFile()
		case ".json":
			return loadJsonFile()
		case ".properties":
			return loadPropertiesFile()
		case ".toml":
			return loadTomlFile()
		case ".yaml":
			return loadYamlFile()
		default:
			kv.Set(elemKey, elemVal)
		}

		return nil
	})
}

// ignoreDir returns true if the directory should be ignored. Reference filepath.Match for pattern syntax
func ignoreDir(path string, ignoreDirs []string) bool {
	for _, dir := range ignoreDirs {
		match, err := filepath.Match(dir, path)
		if err != nil {
			log.Fatal(err) // xxx: better error message
		}
		if match {
			return true
		}
	}
	return false
}

// ignoreFile returns true if the file should be ignored based on file extension matching
func ignoreFile(path string, ignoreExtensions []string) bool {
	pathExtension := filepath.Ext(path)
	if pathExtension == "" {
		return false
	}
	for _, ignoreExtension := range ignoreExtensions {
		if pathExtension == ignoreExtension {
			return true
		}
	}
	return false
}

func loadHclFile() error {
	return nil
}

func loadIniFile() error {
	return nil
}

func loadJsonFile() error {
	// https://github.com/laszlothewiz/golang-snippets-examples/blob/master/walk-JSON-tree.go
	return nil
}

func loadPropertiesFile() error {
	return nil
}

func loadTomlFile() error {
	return nil
}

func loadYamlFile() error {
	return nil
}
