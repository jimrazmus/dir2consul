package main

import (
    "bytes"
    "flag"
    "fmt"
    "io/ioutil"
    "os"
    "regexp"
    "testing"

    "github.com/jimrazmus/dir2consul/kv"
    "github.com/spf13/viper"
)

// go test -update
var update = flag.Bool("update", false, "update .golden files")

func TestSetupEnvironment(t *testing.T) {
    os.Clearenv()
    setupEnvironment()
    if viper.GetString("CONSUL_KEY_PREFIX") != "dir2consul" {
        t.Error("D2C_CONSUL_KEY_PREFIX != dir2consul")
    }
    if viper.GetString("DIRECTORY") != "local/repo" {
        t.Error("D2C_DIRECTORY != local/repo")
    }
    if viper.GetString("DRYRUN") != "false" && !viper.GetBool("DRYRUN") {
        t.Error("D2C_DRYRUN != false")
    }
    if viper.GetString("IGNORE_DIR_REGEX") != "a^" {
        t.Error("D2C_IGNORE_DIR_REGEX != a^")
    }
    if viper.GetString("IGNORE_FILE_REGEX") != "README.md" {
        t.Error("D2C_IGNORE_FILE_REGEX != README.md")
    }
    if viper.GetString("VERBOSE") != "false" && !viper.GetBool("VERBOSE") {
        t.Error("D2C_VERBOSE != false")
    }
}

func TestStartupMessage(t *testing.T) {
    os.Clearenv()
    os.Setenv("TEST", "TestStartupMessage")
    setupEnvironment()
    actual := []byte(startupMessage())
    auFile := "testdata/TestStartupMessage.golden"
    if *update {
        ioutil.WriteFile(auFile, actual, 0644)
    }
    golden, err := ioutil.ReadFile(auFile)
    if err != nil {
        t.Fatal(err)
    }
    if !bytes.Equal(golden, actual) {
        t.Errorf("failed\nexpected:\n%s\ngot:\n%s", string(golden[:]), string(actual[:]))
    }

}

func TestCompileRegexps(t *testing.T) {
    cases := []struct {
        name string
        dre  string
        fre  string
        want error
    }{
        {
            `defaults succeed`,
            `a^`,
            `README.md`,
            nil,
        },
    }

    for i, tc := range cases {
        t.Run(fmt.Sprintf("%d_%s", i, tc.name), func(t *testing.T) {
            _, _, err := compileRegexps(tc.dre, tc.fre)
            if err != tc.want {
                t.Errorf("%s failed (%s)\n", tc.name, err)
            }
        })
    }
}

func TestLoadKeyValuesFromDisk(t *testing.T) {
    cases := []struct {
        name string
        dir  string
        dre  string
        fre  string
    }{
        {
            "skip_everything",
            "project-a",
            `^`,
            `^`,
        },
        {
            "skip_nothing",
            "project-a",
            `a^`,
            `a^`,
        },
        {
            "skip_skipme_dir",
            "project-a",
            `skipme`,
            `a^`,
        },
        {
            "skip_skipme_file",
            "project-a",
            `a^`,
            `skipme`,
        },
        {
            "skip_readme_file",
            "project-a",
            `a^`,
            `README.md`,
        },
        {
            "skip_bigfile",
            "project-b",
            `a^`,
            `a^`,
        },
    }

    for i, tc := range cases {
        t.Run(fmt.Sprintf("%d_%s", i, tc.name), func(t *testing.T) {
            os.Clearenv()
            setupEnvironment()
            err := os.Setenv("D2C_DIRECTORY", fmt.Sprintf("testdata/%s", tc.dir))
            if err != nil {
                t.Fatal(err)
            }
            // err = os.Setenv("D2C_VERBOSE", "true")
            // if err != nil {
            //     t.Fatal(err)
            // }
            var dirIgnoreRe, fileIgnoreRe *regexp.Regexp
            dirIgnoreRe, err = regexp.Compile(tc.dre)
            if err != nil {
                t.Fatal(err)
            }
            fileIgnoreRe, err = regexp.Compile(tc.fre)
            if err != nil {
                t.Fatal(err)
            }
            old, err := os.Getwd()
            if err != nil {
                t.Fatal(err)
            }
            defer os.Chdir(old)
            
            actual := kv.NewList()
            err = loadKeyValuesFromDisk(actual, dirIgnoreRe, fileIgnoreRe)
            if err != nil {
                t.Fatal(err)
            }
            err = os.Chdir(old)
            if err != nil {
                t.Fatal(err)
            }
            auFile := fmt.Sprintf("testdata/%s.golden", tc.name)
            if *update {
                ioutil.WriteFile(auFile, actual.Serialize(), 0644)
            }
            golden, err := ioutil.ReadFile(auFile)
            if err != nil {
                t.Fatal(err)
            }
            if !bytes.Equal(golden, actual.Serialize()) {
                t.Errorf("%s failed\nexpected: %+v\ngot: %+v", tc.name, golden, actual)
            }
        })
    }
}
