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
	"errors"

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
	viper.SetDefault("DEFAULT_CONFIG_TYPE", "")
	viper.SetDefault("DIRECTORY", "local/repo")
	viper.SetDefault("DRYRUN", "false")
	viper.SetDefault("IGNORE_DIR_REGEX", `a^`)
	viper.SetDefault("IGNORE_FILE_REGEX", `README.md`)
	viper.SetDefault("VERBOSE", "false")
	viper.AutomaticEnv()
	viper.BindEnv("CONSUL_KEY_PREFIX")
	viper.BindEnv("DEFAULT_CONFIG_TYPE")
	viper.BindEnv("DIRECTORY")
	viper.BindEnv("DRYRUN")
	viper.BindEnv("IGNORE_DIR_REGEX")
	viper.BindEnv("IGNORE_FILE_REGEX")
	viper.BindEnv("VERBOSE")
}

func startupMessage() string {
	banner := "\n------------\n dir2consul \n------------\n"

	config := "Configuration" + "\n\tD2C_CONSUL_KEY_PREFIX: " + viper.GetString("CONSUL_KEY_PREFIX") + "\n\tD2C_DEFAULT_CONFIG_TYPE: " + viper.GetString("DEFAULT_CONFIG_TYPE") + "\n\tD2C_DIRECTORY: " + viper.GetString("DIRECTORY") + "\n\tD2C_DRYRUN: " + viper.GetString("DRYRUN") + "\n\tD2C_IGNORE_DIR_REGEX: " + viper.GetString("IGNORE_DIR_REGEX") + "\n\tD2C_IGNORE_FILE_REGEX: " + viper.GetString("IGNORE_FILE_REGEX") + "\n\tD2C_VERBOSE: " + viper.GetString("VERBOSE")

	env := os.Environ()
	sort.Strings(env)
	environment := fmt.Sprintf("\nEnvironment\n\t%s", strings.Join(env, "\n\t"))
	// FIXME

	curDir, _ := os.Getwd()
	log.Println(curDir)
	
	files, _ := ioutil.ReadDir(".")

	for _, afile := range files {
		log.Println(afile.Name())
	}

	lfiles, _ := ioutil.ReadDir("local/")

	for _, lfile := range lfiles {
		log.Println(lfile.Name())
	}
	
	// EOFIXME
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

		filetype := strings.TrimPrefix((strings.ToLower(filepath.Ext(path))), ".")
		
		if viper.GetBool("VERBOSE") {
			log.Println("\n\n"+path+"\n  - " + elemKey + "\n")
		}
		
		// Skip processing the root default file...

		// Are we looking at a file, or a default file?
                if filepath.Base(path) == "default" ||
			filepath.Base(path) == "default" + filepath.Ext(path) {
			// Yup ... skipping
			if viper.GetBool("VERBOSE") {
				log.Printf("Skipping default file: %s...", path)
			}
		} else if info.Mode().IsDir() {
			// Skip this, too -- we care about processing files, and only files
			if viper.GetBool("VERBOSE") {
				log.Printf("Skipping directory -- not a file! %s...", path)
			}
		} else {
			if viper.GetBool("VERBOSE") {
				log.Printf("Found non-default file: %s\n    of type: %s", filepath.Base(path), filepath.Ext(path))
			}
			// Find default files
			pathPath := filepath.Dir(path)
			pathRoot := viper.GetString("DIRECTORY")
			defaultList, err := findDefaults(pathPath, pathRoot)
			if err != nil {
				log.Printf("Error processing path %s: %s", pathRoot + "/" + pathPath, err)
				return err
			}
			
			pathFull := pathRoot + "/" + path
			
			// Add our path to the list, making it an ordered list of defaults + our file
			// of interest
			var filesToParse []string

			filesToParse = append(filesToParse, defaultList...)

			switch filetype {
			case "hcl", "ini", "json", "properties", "toml", "yaml", "yml":
				// If we understand the filetype, let Viper parse it...
				if ! strings.HasPrefix(filepath.Base(path), "default") {
					filesToParse = append(filesToParse, pathFull)
				}

				if  viper.GetBool("VERBOSE") {
					for idx, p := range filesToParse {
						log.Printf("    %d    %s", idx, p)
					}
				}

				// Load & merge all the configuration files, in order
				v, err := mergeConfiguration(filesToParse)
				if err != nil {
					log.Printf("Error merging configs! %s", err)
					return err
				}
				

				// iterate over keys within the file
				for _, key := range v.AllKeys() {
					if viper.GetBool("VERBOSE") {
						log.Printf("%s=%s", elemKey+"/"+key, v.GetString(key))
					}
					kv.Set(viper.GetString("CONSUL_KEY_PREFIX")+"/"+elemKey+"/"+key, []byte(v.GetString(key)))
				}
			default:
				// If we don't know the properties file type, load the whole thing as a blob... after the defaults.

				// Revise that:  if we have a default type, load it now...
				defaultType := viper.GetString("DEFAULT_CONFIG_TYPE")

				if defaultType != "" {
					if ! strings.HasPrefix(filepath.Base(path), "default") {
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

				if  viper.GetBool("VERBOSE") {
					for idx, p := range filesToParse {
						log.Printf("+++ %d    %s", idx, p)
					}
				}

				// Load & merge all the configuration files, in order
				v, err := mergeConfiguration(filesToParse)
				if err != nil {
					log.Printf("Error merging configs! %s", err)
					return err
				}
				

				// iterate over keys within the file
				for _, key := range v.AllKeys() {
					if viper.GetBool("VERBOSE") {
						log.Printf("%s=%s", elemKey+"/"+key, v.GetString(key))
					}
					kv.Set(viper.GetString("CONSUL_KEY_PREFIX")+"/"+elemKey+"/"+key, []byte(v.GetString(key)))
				}

				if defaultType == "" {
					// Now that the default files are absorbed, absorb this whole file as a single property.
					if info.Size() > 512000 {
						log.Printf("Skipping %s: size exceeds Consul's 512KB limit", elemKey)
					}
					
					elemVal, err := ioutil.ReadFile(path)
					if err != nil {
						return err
					}
					if viper.GetBool("VERBOSE") {
						log.Printf("%s=%s", elemKey, []byte(elemVal))
					}
					kv.Set(viper.GetString("CONSUL_KEY_PREFIX")+"/"+elemKey, elemVal)
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

func findDefaults(path string, root string) ([]string, error) {
	var results []string
	
	fullPath := root + "/" + path
	
	fullPathInfo, err := os.Stat(fullPath)

	if viper.GetBool("VERBOSE") {
		log.Printf("At findDefaults with:\n  Path: %s\n  Root: %s\n", path, root)
	}
	if os.IsNotExist(err) {
		// Our path doesn't exist
		return nil, err
	} else {
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
						file.Name() == "default" + filepath.Ext(file.Name()) {
						results = append(results, root + "/" + file.Name())
						if viper.GetBool("VERBOSE") {
							log.Printf(" --- %s", root + "/" + file.Name())
						}
						defaultIndex++
					}
				}
			}

			if defaultIndex > 1 {
				return nil, fmt.Errorf("Multiple default files found in %s!", fullPath)
			}

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
				} else {
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
									file.Name() == "default" + filepath.Ext(file.Name()) {
									results = append(results, aPath + "/" + file.Name())
									if viper.GetBool("VERBOSE") {
										log.Printf(" ... %s", aPath + "/" + file.Name())
									}
									defaultIndex++
								}
							}
						}

						if defaultIndex > 1 {
							return nil, fmt.Errorf("Multilple default files found in %s", aPath)
						}
					} else {
						// We found a file, not a directory.  You should never get here
					}
				}
				
			} // end of range
			
		} else {
			err := errors.New("findDefaults called with file instead of path")
			return nil, err
		}
	}
	
	return results, nil
}

func mergeConfiguration(files []string) (config *viper.Viper, err error) {
	
	// Make an array of viper objects
	var viperList []viper.Viper
	
	// Make a viper object to hold the merged config
	zfinal := viper.New()
	
	for _, z := range files {
		
		// For each file in our list, read it into a new viper object
		// zv := viper.New()
		
		// zDir  := filepath.Dir(z)
		// zFile := filepath.Base(z)
		
		// zv.SetConfigName(zFile)
		// zv.AddConfigPath(zDir)
		// zv.SetConfigType("properties")
		
		// err := zv.ReadInConfig()
		zv, err := loadFile(z)
		
		if err != nil {
			return nil, fmt.Errorf("Fatal error config file %s: %s\n", z, err)
		}
		
		// Add this viper object to our array of viper objects
		viperList = append(viperList, *zv)
		
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

func loadFile (path string) (*viper.Viper, error) {
	results := viper.New()
	
	elemKey := strings.TrimSuffix(path, filepath.Ext(path))
	filetype := strings.TrimPrefix((strings.ToLower(filepath.Ext(path))), ".")
	
	defaultType := viper.GetString("DEFAULT_CONFIG_TYPE")
	
	
	switch filetype {
	case "hcl", "ini", "json", "properties", "toml", "yaml", "yml":
		fileDir := filepath.Dir(path)
		fileFile := filepath.Base(path)
		
		results.SetConfigName(fileFile)
		results.AddConfigPath(fileDir)
		results.SetConfigType(filetype)
		
		err := results.ReadInConfig()
		if err != nil {
			return nil, fmt.Errorf("Fatal error config file %s: %s\n", path, err)
		}
	default:
		if defaultType == "" {
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

			if info.Size() > 512000 {
				log.Printf("Skipping %s: size exceeds Consul's 512KB limit", elemKey)
				return nil, fmt.Errorf("Skipping %s: size exceeds Consul's 512KB limit", elemKey)
			}
			
			elemVal, err := ioutil.ReadFile(path)
			if err != nil {
				return nil, err
			}
			if viper.GetBool("VERBOSE") {
				log.Printf("%s=%s", elemKey, []byte(elemVal))
			}
			// kv.Set(viper.GetString("CONSUL_KEY_PREFIX")+"/"+elemKey, elemVal)
			results.Set(elemKey, elemVal)
			// panic(fmt.Sprintf("No default config type, trying to load a file in loadfile..."))
		} else {
			fileDir := filepath.Dir(path)
			fileFile := filepath.Base(path)
			
			results.SetConfigName(fileFile)
			results.AddConfigPath(fileDir)
			results.SetConfigType(defaultType)
			
			err := results.ReadInConfig()
			if err != nil {
				return nil, fmt.Errorf("Fatal error config file %s: %s\n", path, err)
			}
		}
	}
	
	return results, nil
}

// Emacs formatting variables

// Local Variables:
// mode: go
// tab-width: 8
// indent-tabs-mode: t
// standard-indent: 8
// End:
 
