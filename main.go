package main

// import "github.com/jimrazmus/dir2consul/cmd"

// func main() {
// 	cmd.Execute()
// }

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/jimrazmus/kv"
)

func main() {
	// Get Keys from Disk
	diskList := kv.NewList()
	err := LoadKeyValuesFromDisk(viper.GetString("dir"), diskList, viper.GetBool("expand"), viper.GetStringSlice("fileTypes"), viper.GetStringSlice("ignoreDirs"))
	if err != nil {
		return err
	}
	diskKeys := diskList.Keys()
	fmt.Println(diskKeys)

	// Get Keys from Consul

	// Subtract Consul Keys from Disk Keys

	// Add any remaining Disk Keys to Consul

}

// LoadKeyValuesFromDisk walks the file system and loads file contents into a List
func LoadKeyValuesFromDisk(dir string, kv *kv.List, expand bool, filetypes []string, ignoredirs []string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Mode().IsDir() && ignoreDir(path, ignoredirs) {
			return filepath.SkipDir
		}

		if info.Mode().IsRegular() && isFileWanted(info.Name(), filetypes) {
			fmt.Println("file name:", info.Name())

			// if expand is on, do special handling for hcl, ini, json, toml, and yml
			if expand {
				fileExt := filepath.Ext(info.Name())
				switch fileExt {
				case ".hcl":
					loadHclFile()
				case ".ini":
					loadIniFile()
				case ".json":
					loadJsonFile()
				case ".toml":
					loadTomlFile()
				case ".yaml":
					loadYamlFile()
				case ".yml":
					loadYamlFile()
				default:
					loadFile()
				}
			} else {
				// else, just read the file into the kv list

			}
		}

		return nil
	})
}

// ignoreDir returns true if the directory should be ignored
func ignoreDir(path string, ignoredirs []string) bool {
	for _, dir := range ignoredirs {
		if strings.HasPrefix(path, dir) {
			return true
		}
	}
	return false
}

// isFileWanted returns true if the file extension matches one in the extensions list
func isFileWanted(filename string, extentions []string) bool {
	fileExt := filepath.Ext(filename)
	for _, wantExt := range extentions {
		if fileExt == wantExt {
			return true
		}
	}
	return false
}

func loadFile() (string, error) {
	b, err := ioutil.ReadFile("file.txt")
	if err != nil {
		return "", err
	}
	s := string(b)
	return s, nil
}

// parseKey returns the key string
func parseKey() {

}
