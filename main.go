package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/jimrazmus/dir2consul/kv"
	"github.com/spf13/viper"
)

func main() {

	setupEnvironment()
	fmt.Println(startupMessage())

	dirIgnoreRe, fileIgnoreRe, err := compileRegexps(viper.GetString("IGNORE_DIR_REGEX"), viper.GetString("IGNORE_FILE_REGEX"))
	if err != nil {
		log.Fatal(err)
	}

	// Establish a Consul client
	// Lots of configuration is encapsulated here.
	// Reference https://github.com/hashicorp/consul/tree/master/api
	consulClient, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatal("Error establishing Consul client:", err)
	}

	// Get KVs from Files
	fileKeyValues := kv.NewList()
	err = loadKeyValuesFromDisk(fileKeyValues, dirIgnoreRe, fileIgnoreRe)
	if err != nil {
		log.Fatal(err)
	}

	// Get KVs from Consul
	consulKeyValues := kv.NewList()
	consulKVPairs, _, err := consulClient.KV().List(viper.GetString("CONSUL_KEY_PREFIX"), nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, consulKVPair := range consulKVPairs {
		consulKeyValues.Set(consulKVPair.Key, consulKVPair.Value)
	}

	// Add or update data in Consul when it doesn't match the file data
	addOrUpdateConsulData(fileKeyValues, consulKeyValues, consulClient)

	// Delete data from Consul that doesn't exist in the file data
	deleteExtraConsulData(fileKeyValues, consulKeyValues, consulClient)

}

func setupEnvironment() {
	viper.SetEnvPrefix("D2C")
	viper.SetDefault("CONSUL_KEY_PREFIX", "dir2consul")
	viper.SetDefault("DIRECTORY", "local/repo")
	viper.SetDefault("DRYRUN", "false")
	viper.SetDefault("IGNORE_DIR_REGEX", `a^`)
	viper.SetDefault("IGNORE_FILE_REGEX", `README.md`)
	viper.SetDefault("VERBOSE", "false")
	viper.AutomaticEnv()
	viper.BindEnv("CONSUL_KEY_PREFIX")
	viper.BindEnv("DIRECTORY")
	viper.BindEnv("DRYRUN")
	viper.BindEnv("IGNORE_DIR_REGEX")
	viper.BindEnv("IGNORE_FILE_REGEX")
	viper.BindEnv("VERBOSE")
}

func startupMessage() string {
	banner := "\n------------\n dir2consul \n------------\n"

	config := "Configuration" + "\n\tD2C_CONSUL_KEY_PREFIX: " + viper.GetString("CONSUL_KEY_PREFIX") + "\n\tD2C_DIRECTORY: " + viper.GetString("DIRECTORY") + "\n\tD2C_DRYRUN: " + viper.GetString("DRYRUN") + "\n\tD2C_IGNORE_DIR_REGEX: " + viper.GetString("IGNORE_DIR_REGEX") + "\n\tD2C_IGNORE_FILE_REGEX: " + viper.GetString("IGNORE_FILE_REGEX") + "\n\tD2C_VERBOSE: " + viper.GetString("VERBOSE")

	env := os.Environ()
	sort.Strings(env)
	environment := fmt.Sprintf("\nEnvironment\n\t%s", strings.Join(env, "\n\t"))

	return banner + config + environment
}

func compileRegexps(dirPcre string, filePcre string) (*regexp.Regexp, *regexp.Regexp, error) {
	var err error
	var dirRe, fileRe *regexp.Regexp

	dirRe, err = regexp.Compile(dirPcre)
	if err != nil {
		return nil, nil, fmt.Errorf("Ignore Dir Regex failed to compile: %v", err)
	}

	fileRe, err = regexp.Compile(filePcre)
	if err != nil {
		return nil, nil, fmt.Errorf("Ignore File Regex failed to compile: %v", err)
	}

	return dirRe, fileRe, nil
}

// loadKeyValuesFromDisk walks the file system and loads file contents into a kv.List
func loadKeyValuesFromDisk(kv *kv.List, dirIgnoreRe *regexp.Regexp, fileIgnoreRe *regexp.Regexp) error {
	// Change directory to where the files are located
	err := os.Chdir(viper.GetString("DIRECTORY"))
	if err != nil {
		log.Fatal("Couldn't change directory to:", viper.GetString("DIRECTORY"))
	}

	// Walk the filesystem
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

		elemKey := strings.TrimSuffix(path, filepath.Ext(path))
		if viper.IsSet("CONSUL_KEY_PREFIX") {
			elemKey = strings.Join([]string{viper.GetString("CONSUL_KEY_PREFIX"), elemKey}, "/")
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

func addOrUpdateConsulData(fileKeyValues *kv.List, consulKeyValues *kv.List, consulClient *api.Client) {
	// Add or update data in Consul when it doesn't match the file data
	for _, key := range fileKeyValues.Keys() {
		_, fb, _ := fileKeyValues.Get(key, nil)
		_, cb, _ := consulKeyValues.Get(key, nil)
		if bytes.Compare(fb, cb) != 0 {
			if viper.GetBool("DRYRUN") {
				continue
			}
			if viper.GetBool("VERBOSE") {
				log.Printf("SET key: %s value: %s\n", key, string(fb))
			}
			p := &api.KVPair{Key: key, Value: fb}
			_, err := consulClient.KV().Put(p, nil)
			if err != nil {
				log.Println("Failed Consul KV Put:", err)
			}
		}
	}
}

func deleteExtraConsulData(fileKeyValues *kv.List, consulKeyValues *kv.List, consulClient *api.Client) {
	// Delete data from Consul that doesn't exist in the file data
	for _, key := range consulKeyValues.Keys() {
		_, _, err := fileKeyValues.Get(key, nil)
		if err != nil { // xxx: check for the not exist err
			if viper.GetBool("DRYRUN") {
				continue
			}
			if viper.GetBool("VERBOSE") {
				log.Printf("DELETE key: %s\n", key)
			}
			_, err := consulClient.KV().Delete(key, nil)
			if err != nil {
				log.Println("Failed Consul KV Delete:", err)
			}
		}
	}
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
