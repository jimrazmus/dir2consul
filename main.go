package main

import (
	"bytes"
	"errors"
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
		_, _, err = consulKeyValues.Set(consulKVPair.Key, consulKVPair.Value)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Add or update data in Consul when it doesn't match the file data
	addOrUpdateConsulData(fileKeyValues, consulKeyValues, consulClient)

	// Delete data from Consul that doesn't exist in the file data
	deleteExtraConsulData(fileKeyValues, consulKeyValues, consulClient)

}

func setupEnvironment() {
	envDefaults := map[string]string{
		"CONSUL_KEY_PREFIX":   "dir2consul",
		"DEFAULT_CONFIG_TYPE": "",
		"DIRECTORY":           "local/repo",
		"DRYRUN":              "false",
		"IGNORE_DIR_REGEX":    `a^`,
		"IGNORE_FILE_REGEX":   `README.md`,
		"VERBOSE":             "false",
	}

	viper.SetEnvPrefix("D2C")

	for key, val := range envDefaults {
		viper.SetDefault(key, val)
	}

	viper.AutomaticEnv()

	for key, _ := range envDefaults {
		err := viper.BindEnv(key)
		if err != nil {
			log.Fatalf("Error setting up environment: %s", err)
		}
	}
}

func startupMessage() string {
	banner := "\n------------\n dir2consul \n------------\n"

	config := "Configuration" + "\n\tD2C_CONSUL_KEY_PREFIX: " + viper.GetString("CONSUL_KEY_PREFIX") + "\n\tD2C_DEFAULT_CONFIG_TYPE: " + viper.GetString("DEFAULT_CONFIG_TYPE") + "\n\tD2C_DIRECTORY: " + viper.GetString("DIRECTORY") + "\n\tD2C_DRYRUN: " + viper.GetString("DRYRUN") + "\n\tD2C_IGNORE_DIR_REGEX: " + viper.GetString("IGNORE_DIR_REGEX") + "\n\tD2C_IGNORE_FILE_REGEX: " + viper.GetString("IGNORE_FILE_REGEX") + "\n\tD2C_VERBOSE: " + viper.GetString("VERBOSE")

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

	// Store where we are currently
	curWD, err := os.Getwd()
	if err != nil {
		log.Fatal("Couldn't get current working directory!")
	}
	// Check if the DIRECTORY environment variable is an absolute path...
	if filepath.IsAbs(viper.GetString("DIRECTORY")) {
		// Our root directory is an absolute path and we can just move along...
		err := os.Chdir(viper.GetString("DIRECTORY"))
		if err != nil {
			log.Fatal("Couldn't change directory to:", viper.GetString("DIRECTORY"))
		}
	} else {
		// Our root directory is NOT an absolute path, so do some trickery here...
		if strings.HasSuffix(curWD, viper.GetString("DIRECTORY")) {
			// Our current working directory is already in our current path, do nothing
			// NOTE: using HasSuffix.  CurWD should be an absolute path. If CurWD
			// NOTE: *ends* with the contents of DIRECTORY, we've already moved to
			// NOTE: where we want, and we don't have to chdir
		} else {
			// Our current working directory doesn't appear to be where we want to be
			// so try and move there...
			err := os.Chdir(viper.GetString("DIRECTORY"))
			if err != nil {
				log.Fatal("Couldn't change directory to:", viper.GetString("DIRECTORY"))
			}
		}
	}
	// We should now be where we want to be, hopefully...

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

		// Skip over "default" files
		if filepath.Base(path) == "default" ||
			filepath.Base(path) == "default"+filepath.Ext(path) {
			// We have a default file with some extension... skipping
			// NOTE: This does not compare the extension to anything, so
			// NOTE: default.txt will be treated as a default file.  This
			// NOTE: is not necessarily right...
			if viper.GetBool("VERBOSE") {
				log.Printf("Skipping default file: %s...", path)
			}
			return nil
		}

		elemKey := strings.TrimSuffix(path, filepath.Ext(path))

		filetype := strings.TrimPrefix((strings.ToLower(filepath.Ext(path))), ".")

		if viper.GetBool("VERBOSE") {
			log.Println("\n\n" + path + "\n  - " + elemKey + "\n")
		}

		// Find default files in the paths between where we started and where this file is.

		// The path of the file we have just hit.
		pathPath := filepath.Dir(path)

		// The path we started at
		pathRoot := viper.GetString("DIRECTORY")

		// Call findDefaults and generate a list of all defaults between the "root" and where we are now,
		// including default files at the same level of the directory hierarchy as we currently are.
		defaultList, err := findDefaults(pathPath, pathRoot)
		if err != nil {
			log.Printf("Error processing path %s: %s", pathRoot+"/"+pathPath, err)
			return err
		}

		// This is the "full" path between where we started -- or the absolute root of the filesystem -- and
		// where we are now
		comboPath := pathRoot + "/" + path

		// This should be an absolute path to the file we're at right now.
		var pathFull string

		if filepath.IsAbs(comboPath) {
			// if the comboPath is an absolute path, set pathFull to it
			pathFull = comboPath
		} else {
			// if the comboPath is relative (ie, if the DIRECTORY environment variable is relative)
			// construct an absolute path and assign it to pathFull
			curWD, err := os.Getwd()
			if err != nil {
				return err
			}

			if strings.HasSuffix(curWD, pathRoot) {
				// We already have the pathRoot in our WD
				pathFull = curWD + "/" + path
			} else {
				err = nil
				pathFull, err = filepath.Abs(pathPath + "/" + path)
				if err != nil {
					return err
				}
			}
		}

		// Construct a list of files we care about. Start with the list of defaults we found...
		var filesToParse []string

		filesToParse = append(filesToParse, defaultList...)

		// Check the type of the file we're parsing (ie, not the defaults)
		switch filetype {
		case "hcl", "ini", "json", "properties", "toml", "yaml", "yml":
			// If we understand the filetype, let Viper parse it...
			if !strings.HasPrefix(filepath.Base(path), "default") {
				filesToParse = append(filesToParse, pathFull)
			}

			if viper.GetBool("VERBOSE") {
				for idx, p := range filesToParse {
					log.Printf("    %d    %s", idx, p)
				}
			}

			// Load & merge all the configuration files, in order of precedence (ie, all defaults
			// from the top of the hierarchy down to the file we are looking at, then the file
			// we're looking at.  The results of all the properties in all those files should come
			// to us in the viper object 'v'.
			v, err := mergeConfiguration(filesToParse)
			if err != nil {
				if viper.GetBool("VERBOSE") {
					log.Printf("Error merging configs! %s", err)
				}
				return nil
			}

			// iterate over keys within the merged viper object, and set them in the 'kv' store
			for _, key := range v.AllKeys() {
				if viper.GetBool("VERBOSE") {
					log.Printf("%s=%s", elemKey+"/"+key, v.GetString(key))
				}
				_, _, err = kv.Set(viper.GetString("CONSUL_KEY_PREFIX")+"/"+elemKey+"/"+key, []byte(v.GetString(key)))
				if err != nil {
					log.Fatal(err)
				}
			}
		default:
			// If we don't recognize the file's type (ie, it's something like bob.txt, instead of a
			// proper configuration format, or it's just called 'default' with no extension...

			// If we have a DEFAULT_CONFIG_TYPE set in the environment, use that
			defaultType := viper.GetString("DEFAULT_CONFIG_TYPE")

			if defaultType != "" {
				// if we have a default type, add the file to our "to be parsed list", as usual.
				// mergeConfiguration will treat it as that specified default type automagically
				if !strings.HasPrefix(filepath.Base(path), "default") {
					filesToParse = append(filesToParse, pathFull)
					if viper.GetBool("VERBOSE") {
						log.Printf("Adding %s...", pathFull)
					}
				} else {
					if viper.GetBool("VERBOSE") {
						log.Printf("Skipping %s...", pathFull)
					}
				}
			}

			if viper.GetBool("VERBOSE") {
				for idx, p := range filesToParse {
					log.Printf("+++ %d    %s", idx, p)
				}
			}

			// Load & merge all the configuration files, in order
			// NOTE:  If we don't have a default type, this list will only be the defaults files
			// NOTE:  Not our file of interest...
			v, err := mergeConfiguration(filesToParse)
			if err != nil {
				if viper.GetBool("VERBOSE") {
					log.Printf("Error merging configs! %s", err)
				}
				return nil
			}

			// iterate over keys within the merged viper configuration object
			// NOTE:  If we don't have a default type set in the environment, this will only be a merged
			// NOTE:  property file of all the defaults
			for _, key := range v.AllKeys() {
				if viper.GetBool("VERBOSE") {
					log.Printf("%s=%s", elemKey+"/"+key, v.GetString(key))
				}
				_, _, err = kv.Set(viper.GetString("CONSUL_KEY_PREFIX")+"/"+elemKey+"/"+key, []byte(v.GetString(key)))
				if err != nil {
					log.Fatal(err)
				}
			}

			// If we did *NOT* have a default type set, now snarf the untyped/unrecognized file into our
			// kv set automagically as a single blob.
			if defaultType == "" {
				// Now that the default files are absorbed, absorb this whole file as a single property.
				if info.Size() > 512000 {
					if viper.GetBool("VERBOSE") {
						log.Printf("Skipping %s: size exceeds Consul's 512KB limit", elemKey)
					}
					return nil
				}

				elemVal, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}
				if viper.GetBool("VERBOSE") {
					log.Printf("%s=%s", elemKey, []byte(elemVal))
				}
				_, _, err = kv.Set(viper.GetString("CONSUL_KEY_PREFIX")+"/"+elemKey, elemVal)
				if err != nil {
					log.Fatal(err)
				}
			}
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

func findDefaults(path string, rootProvided string) ([]string, error) {
	// Starting with a root, and a path, walk that path from the top to the bottom, looking for
	// files named 'default.extension' and return an array of them.

	var results []string
	var fullPath string
	var root string

	if filepath.IsAbs(rootProvided) {
		// We have an aboslute path, do the obvious
		root = rootProvided
	} else {
		// We have a relative path.  Is that relative path part of the current path?

		curWD, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		if strings.HasSuffix(curWD, rootProvided) {
			// our WD already includes our provided root
			root = curWD
		} else {
			// our WD does *not* include our provided root
			root, err = filepath.Abs(rootProvided)
			if err != nil {
				return nil, err
			}
		}
	}

	if viper.GetBool("VERBOSE") {
		log.Printf("At findDefaults with:\n  Path: %s\n  Root: %s\n   Abs: %s\n", path, rootProvided, root)
	}

	fullPath = root + "/" + path

	fullPathInfo, err := os.Stat(fullPath)

	if viper.GetBool("VERBOSE") {
		log.Printf("At findDefaults with:\n  Full Root: %s\n  Full Path: %s\n", root, fullPath)
	}

	if os.IsNotExist(err) {
		// Our path doesn't exist
		return nil, err
	}

	if fullPathInfo.IsDir() {

		// scan each file path entry for files, then scan them for `default` files
		pathFiles, err := ioutil.ReadDir(root)
		if err != nil {
			return nil, err
		}

		defaultIndex := 0
		for _, file := range pathFiles {
			if file.IsDir() {
				// skip
			} else {
				if file.Name() == "default" ||
					file.Name() == "default"+filepath.Ext(file.Name()) {
					results = append(results, root+"/"+file.Name())
					if viper.GetBool("VERBOSE") {
						log.Printf(" --- %s", root+"/"+file.Name())
					}
					defaultIndex++
				}
			}
		}
		// If we have more than one file named "default" or "default.<ext>" at a given level in the
		// directory hierarchy, the precedence of applying them is uncertain.  Fail.
		// NOTE:  This will also die if we don't know what kind of files they are, like if they are named
		// NOTE:  "default.txt" or "default.excel" or whatever...
		// NOTE:  This should probably be much smoother.
		if defaultIndex > 1 {
			return nil, fmt.Errorf("Multiple default files found in %s", fullPath)
		}

		// Take our path and split it up into component parts.  Then iterate over them, checking each level
		// for default files.
		dirElements := strings.Split(path, "/")

		var dirConcat string

		for idx, a := range dirElements {
			// Iterate over path elements, looking for a `default` file in each one...
			if idx == 0 {
				dirConcat = a
			} else {
				dirConcat = dirConcat + "/" + a
			}

			// Full path of the element we're currently at
			aPath := root + "/" + dirConcat

			// Does this exist?  it really should, but worse things happen at sea...
			aPathInfo, err := os.Stat(aPath)
			if os.IsNotExist(err) {
				// You should never get here
				return nil, err
			}

			if aPathInfo.IsDir() {
				pathFiles, err := ioutil.ReadDir(aPath)
				if err != nil {
					return nil, err
				}

				defaultIndex := 0

				for _, file := range pathFiles {
					if file.IsDir() {
						// skip
					} else {
						if file.Name() == "default" ||
							file.Name() == "default"+filepath.Ext(file.Name()) {
							results = append(results, aPath+"/"+file.Name())
							if viper.GetBool("VERBOSE") {
								log.Printf(" ... %s", aPath+"/"+file.Name())
							}
							defaultIndex++
						}
					}
				}

				if defaultIndex > 1 {
					return nil, fmt.Errorf("Multiple default files found in %s", aPath)
				}
			} else {
				// We found a file, not a directory.  You should never get here
			}

		} // end of range

	} else {
		// We want to be parsing a path.  We should be called with a root and a path and that's it.
		err := errors.New("findDefaults called with file instead of path")
		return nil, err
	}

	return results, nil
}

func mergeConfiguration(files []string) (config *viper.Viper, err error) {
	// Take a list of files, return a single Viper configuration object containing
	// the properties present in each individual file, in the same order as they are
	// in the list.
	//
	// In other words, if you set a property in the file in the first element of the array,
	// then later override it with the same property name, but a different value, in the
	// file in the third element of the array, you would end up with the value from that
	// third file.  It would override the value in the first.

	// Make a viper object to hold the merged config
	zfinal := viper.NewWithOptions(viper.KeyDelimiter("/"))

	for _, z := range files {

		// For each file in our list, read it into a new viper object
		zv, err := loadFile(z)

		if err != nil {
			return nil, fmt.Errorf("Fatal error config file %s: %s", z, err)
		}

		zvSettings := zv.AllSettings()

		// Merge in the settings in the newly loaded viper object into our
		// merged viper object
		err = zfinal.MergeConfigMap(zvSettings)
		if err != nil {
			return nil, fmt.Errorf("Unable to merge configuration! %s", err)
		}
	}

	return zfinal, nil
}

func loadFile(path string) (*viper.Viper, error) {
	// If given a file, load it into a viper object

	// Normally this is straightforward.  We have some special behavior if it's not a "real" property file.
	// In that case, we load it as a single "blob" (assuming it's small enough to be held as a blob in consul).

	results := viper.NewWithOptions(viper.KeyDelimiter("/"))

	elemKey := strings.TrimSuffix(path, filepath.Ext(path))
	filetype := strings.TrimPrefix((strings.ToLower(filepath.Ext(path))), ".")

	// If we have a default type set in the environment, and we can't otherwise type the file,
	// treat it as this type.  If we don't have it set, just go ahead and blob it.
	// This lets you have default properties files just named "default" with no extension, if you want.
	// By default, we do *NOT* provide a DEFAULT_CONFIG_TYPE
	defaultType := viper.GetString("DEFAULT_CONFIG_TYPE")

	switch filetype {
	case "hcl", "ini", "json", "properties", "toml", "yaml", "yml":
		// The file type is well undersood.  Load away.
		fileDir := filepath.Dir(path)
		fileFile := filepath.Base(path)

		results.SetConfigName(fileFile)
		results.AddConfigPath(fileDir)
		results.SetConfigType(filetype)

		err := results.ReadInConfig()
		if err != nil {
			return nil, fmt.Errorf("Fatal error config file %s: %s", path, err)
		}
	default:
		if defaultType == "" {
			// We have no default type, and we don't know the file's type.

			// Read it in as a blob, unless it's too big

			// We already have this from earlier in the program.  It would behoove us to figure
			// out how to get it down to here in the callstack w/out restat-ing the file...
			aFile, err := os.Open(path)
			if err != nil {
				log.Printf("Unable to open file %s: %s", path, err)
				return nil, err
			}

			info, err := aFile.Stat()
			if err != nil {
				log.Printf("Error stat-ing file: %s", err)
				return nil, err
			}

			// If the file is too big to fit into a consul value, error out.
			if info.Size() > 512000 {
				if viper.GetBool("VERBOSE") {
					log.Printf("Skipping %s: size exceeds Consul's 512KB limit", elemKey)
				}
				return nil, fmt.Errorf("Skipping %s: size exceeds Consul's 512KB limit", elemKey)
			}

			elemVal, err := ioutil.ReadFile(path)
			if err != nil {
				return nil, err
			}
			if viper.GetBool("VERBOSE") {
				log.Printf("%s=%s", elemKey, []byte(elemVal))
			}

			// Load the value into our viper object
			results.Set(elemKey, elemVal)
		} else {
			// We have a default type.  Load this file using viper, as a file of that type, into our viper object
			fileDir := filepath.Dir(path)
			fileFile := filepath.Base(path)

			results.SetConfigName(fileFile)
			results.AddConfigPath(fileDir)
			results.SetConfigType(defaultType)

			err := results.ReadInConfig()
			if err != nil {
				return nil, fmt.Errorf("Fatal error config file %s: %s", path, err)
			}
		}
	}

	return results, nil
}
