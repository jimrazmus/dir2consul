package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
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

func TestFindDefaults(t *testing.T) {
	cases := []struct {
		name   string
		path   string
		root   string
		expect bool
		data   []string
	}{
		{
			"check_file",
			"b.hcl",
			"testdata/project-c/a",
			false,
			nil,
		},
		{
			"check_path",
			"project-c/a",
			"testdata",
			true,
			[]string{
				0: "testdata/project-c/default.hcl",
				1: "testdata/project-c/a/default.hcl"},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d_%s", i, tc.name), func(t *testing.T) {
			os.Clearenv()
			setupEnvironment()

			results, err := findDefaults(tc.path, tc.root)

			if err != nil {
				if tc.expect {
					t.Fatal(err)
				} else {
					// We expect to fail, which is technically a pass
				}
			}
			curWD, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			// Did we find the files we expected?
			for idx, x := range results {
				if strings.HasPrefix(x, curWD) &&
					strings.HasSuffix(x, tc.data[idx]) {
					// We did!
				} else {
					// If we don't sit under the current working directory AND
					// have the path to the expected file as the last elements
					// of your path, you do not match!
					t.Errorf("findDefaults: %s does not match %s", x, tc.data[idx])
				}
			}
		})
	}
}

func TestMergeConfigurations(t *testing.T) {
	testValues := map[string]string{
		"def_one":        "1",
		"def_two":        "2",
		"def_three":      "3",
		"default_four":   "4",
		"default_five":   "5",
		"default_six":    "6",
		"override_one":   "a",
		"override_two":   "b",
		"override_three": "c",
		"b_one":          "1",
	}

	fileList := []string{
		"testdata/project-c/default.hcl",
		"testdata/project-c/a/default.hcl",
		"testdata/project-c/a/b.hcl",
	}

	os.Clearenv()
	setupEnvironment()

	curWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	for idx, x := range fileList {
		x_abs := curWD + "/" + x
		fileList[idx] = x_abs
	}

	v, err := mergeConfiguration(fileList)
	if err != nil {
		t.Fatal(err)
	}

	// You have to loop over both sets of keys, because otherwise you might have values in
	// whatever you didn't loop over that you never check.  You don't just want all keys in a
	// to have matching values in b, you also don't want any keys in b with any values that
	// don't also appear in a.  Simplest way to approach that -- iterate over both sets of keys
	// and compare values to the other.  You could make a merged list of keys, and then only
	// do value comparisons once, but the differences would be marginal.

	// check values of all keys in loaded data vs. test values, above.
	for _, key := range v.AllKeys() {
		lv := v.GetString(key)
		dv := testValues[key]

		if lv != dv {
			t.Errorf("For key %s, %s does not equal %s", key, lv, dv)
		} else {
			// Everybody matches, no error
		}
	}

	// Check values of all keys in test values are in loaded values
	for key, dv := range testValues {
		lv := v.GetString(key)

		if lv != dv {
			t.Errorf("for key %s, %s does not equal %s", key, dv, lv)
		} else {
			// Everybody matches, no error
		}
	}
}

func TestLoadFile(t *testing.T) {
	testValues := map[string]string{
		"def_one":        "1",
		"def_two":        "2",
		"def_three":      "3",
		"override_one":   "a",
		"override_two":   "a",
		"override_three": "a",
	}

	os.Clearenv()
	setupEnvironment()

	curWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	testFile := "testdata/project-c/default.hcl"

	v, err := loadFile(curWD + "/" + testFile)
	if err != nil {
		t.Fatal(err)
	}

	// You have to loop over both sets of keys, because otherwise you might have values in
	// whatever you didn't loop over that you never check.  You don't just want all keys in a
	// to have matching values in b, you also don't want any keys in b with any values that
	// don't also appear in a.  Simplest way to approach that -- iterate over both sets of keys
	// and compare values to the other.  You could make a merged list of keys, and then only
	// do value comparisons once, but the differences would be marginal.

	// check values of all keys in loaded data vs. test values, above.
	for _, key := range v.AllKeys() {
		lv := v.GetString(key)
		dv := testValues[key]

		if lv != dv {
			t.Errorf("For key %s, %s does not equal %s", key, lv, dv)
		} else {
			// Keys values match, which we desire
		}
	}

	// Check values of all keys in test values are in loaded values
	for key, dv := range testValues {
		lv := v.GetString(key)

		if lv != dv {
			t.Errorf("for key %s, %s does not equal %s", key, dv, lv)
		} else {
			// Keys values match, which is what we want.
		}
	}

}

