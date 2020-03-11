package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/jimrazmus/dir2consul/kv"
)

// conditionally compile in or out the debug prints
const debug = false

// ConsulKeyPrefix is the path prefix to prepend to all consul keys
var ConsulKeyPrefix = getenv("D2C_CONSUL_KEY_PREFIX", "")

// Directory is the directory we should walk
var Directory = getenv("D2C_DIRECTORY", "local")

// DryRun is a flag to prevent Consul data modifications
var DryRun = getenv("D2C_DRYRUN", "")

// IgnoreDirRegex is a PCRE regular expression that matches directories we ignore when walking the file system
var IgnoreDirRegex = getenv("D2C_IGNORE_DIR_REGEX", `a^`)

// IgnoreFileRegex is a PCRE regular expression that matches files we ignore when walking the file system
var IgnoreFileRegex = getenv("D2C_IGNORE_FILE_REGEX", `README.md`)

var dirIgnoreRe, fileIgnoreRe *regexp.Regexp

func main() {
	log.Println("dir2consul starting with configuration:")
	log.Println("D2C_CONSUL_KEY_PREFIX:", ConsulKeyPrefix)
	log.Println("D2C_DIRECTORY:", Directory)
	log.Println("D2C_DRYRUN:", DryRun)
	log.Println("D2C_IGNORE_DIR_REGEX:", IgnoreDirRegex)
	log.Println("D2C_IGNORE_FILE_REGEX:", IgnoreFileRegex)

	var err error

	// Compile regular expressions
	dirIgnoreRe, err = regexp.Compile(IgnoreDirRegex)
	if err != nil {
		log.Fatal("Ignore Dir Regex failed to compile:", err)
	}
	fileIgnoreRe = regexp.MustCompile(IgnoreFileRegex)
	if err != nil {
		log.Fatal("Ignore File Regex failed to compile:", err)
	}

	os.Chdir(Directory)

	// Establish a Consul client
	// Lots of configuration is encapsulated here.
	// Reference https://github.com/hashicorp/consul/tree/master/api
	consulClient, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatal("Error establishing Consul client:", err)
	}

	// Get KVs from Files
	fileKeyValues := kv.NewList()
	err = LoadKeyValuesFromDisk(fileKeyValues)
	if err != nil {
		log.Fatal(err)
	}
	if debug {
		fileKeys := fileKeyValues.Keys()
		log.Println("fileKeys:", fileKeys)
	}

	// Get KVs from Consul
	consulKeyValues := kv.NewList()
	consulKVPairs, _, err := consulClient.KV().List(ConsulKeyPrefix, nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, consulKVPair := range consulKVPairs {
		consulKeyValues.Set(consulKVPair.Key, consulKVPair.Value)
	}
	if debug {
		consulKeys := consulKeyValues.Keys()
		log.Println("consulKeys:", consulKeys)
	}

	// Add or update data in Consul when it doesn't match the file data
	for _, key := range fileKeyValues.Keys() {
		_, fb, _ := fileKeyValues.Get(key, nil)
		_, cb, _ := consulKeyValues.Get(key, nil)
		if bytes.Compare(fb, cb) != 0 {
			if DryRun != "" {
				log.Printf("SET key: %s value: %s\n", key, string(fb))
				continue
			}
			p := &api.KVPair{Key: key, Value: fb}
			_, err = consulClient.KV().Put(p, nil)
			if err != nil {
				log.Println("Failed Consul KV Put:", err)
			}
		}
	}

	// Delete data from Consul that doesn't exist in the file data
	for _, key := range consulKeyValues.Keys() {
		_, _, err := fileKeyValues.Get(key, nil)
		if err != nil { // xxx: check for the not exist err
			if DryRun != "" {
				log.Printf("DELETE key: %s\n", key)
				continue
			}
			_, err := consulClient.KV().Delete(key, nil)
			if err != nil {
				log.Println("Failed Consul KV Delete:", err)
			}
		}
	}

}

// LoadKeyValuesFromDisk walks the file system and loads file contents into a kv.List
func LoadKeyValuesFromDisk(kv *kv.List) error {
	return filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip dot directory
		if info.Mode().IsDir() && info.Name() == "." {
			return nil
		}

		// Skip over hidden directories
		if info.Mode().IsDir() && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}

		// Skip over directories we want to ignore
		if info.Mode().IsDir() && dirIgnoreRe.MatchString(path) {
			return filepath.SkipDir
		}

		// Skip directories, non-regular files, and dot files
		if info.Mode().IsDir() || !info.Mode().IsRegular() || strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		// Skip files we want to ignore
		if info.Mode().IsRegular() && fileIgnoreRe.MatchString(info.Name()) {
			return nil
		}

		if debug {
			log.Println("path:", path)
		}

		elemKey := strings.TrimSuffix(path, filepath.Ext(path))
		if ConsulKeyPrefix != "" {
			elemKey = strings.Join([]string{ConsulKeyPrefix, elemKey}, "/")
		}

		elemVal, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		if len(elemVal) > 512000 {
			log.Printf("Skipping %s: value length exceeds Consul's 512KB limit", elemKey)
			return nil
		}

		switch strings.ToLower(filepath.Ext(path)) {
		// case ".hcl":
		// 	return loadHclFile()
		// case ".ini":
		// 	return loadIniFile()
		// case ".json":
		// 	return loadJsonFile()
		// case ".properties":
		// 	return loadPropertiesFile()
		// case ".toml":
		// 	return loadTomlFile()
		// case ".yaml":
		// 	return loadYamlFile()
		default:
			kv.Set(elemKey, elemVal)
		}

		return nil
	})
}

// func loadHclFile() error {
// 	return nil
// }

// func loadIniFile() error {
// 	return nil
// }

// func loadJsonFile() error {
// 	// https://github.com/laszlothewiz/golang-snippets-examples/blob/master/walk-JSON-tree.go
// 	return nil
// }

// func loadPropertiesFile() error {
// 	return nil
// }

// func loadTomlFile() error {
// 	return nil
// }

// func loadYamlFile() error {
// 	return nil
// }
