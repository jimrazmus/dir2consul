package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/consul/sdk/testutil"
	"github.com/jimrazmus/dir2consul/kv"
)

// go test -update
var update = flag.Bool("update", false, "update .golden files")

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
			"skip_some",
			"project-a",
			`.gitx|skipme`,
			`README.md`,
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d_%s", i, tc.name), func(t *testing.T) {
			var err error
			dirIgnoreRe, err = regexp.Compile(tc.dre)
			if err != nil {
				t.Fatal(err)
			}
			fileIgnoreRe = regexp.MustCompile(tc.fre)
			if err != nil {
				t.Fatal(err)
			}
			old, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			err = os.Chdir(fmt.Sprintf("testdata/%s", tc.dir))
			if err != nil {
				t.Fatal(err)
			}
			actual := kv.NewList()
			err = LoadKeyValuesFromDisk(actual)
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
