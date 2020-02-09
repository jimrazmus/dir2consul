package main

import "testing"

func TestIgnoreDir(t *testing.T) {
	fixtures := []struct {
		dir    string
		dirs   []string
		result bool
	}{
		{".git", []string{".git", "foo", "bar"}, true},
		{".Git", []string{".git", "foo", "bar"}, false},
		{"foo", []string{".git", "foo", "bar"}, true},
		{"fOo", []string{".git", "foo", "bar"}, false},
		{"bar", []string{".git", "foo", "bar"}, true},
		{"baR", []string{".git", "foo", "bar"}, false},
		{".git/a", []string{".git/*", "foo", "bar"}, true},
		{".git/a", []string{".git", "foo", "bar"}, false},
		{"a/.git", []string{".git", "foo", "bar"}, false},
		{"a/.git/", []string{".git", "foo", "bar"}, false},
		{"foo/a", []string{".git", "foo/a", "bar"}, true},
		{"fo/oa", []string{".git", "foo", "bar"}, false},
		{"a/foo", []string{".git", "foo", "bar"}, false},
		{"a/foo/", []string{".git", "foo", "bar"}, false},
		{"bar/a", []string{".git", "foo", "bar"}, false},
		{"ba/ra", []string{".git", "foo", "bar"}, false},
		{"a/bar", []string{".git", "foo", "bar"}, false},
		{"a/bar/", []string{".git", "foo", "bar"}, false},
		{"b/.git/a", []string{".git", "foo", "bar"}, false},
		{"b/foo/a", []string{".git", "foo", "bar"}, false},
		{"b/bar/a", []string{".git", "foo", "bar"}, false},
		{".bar", []string{".git", "foo", "bar"}, false},
		{"foo.bar", []string{".git", "foo", "bar"}, false},
		{"foo.bar", []string{".git", "*foo", "bar"}, false},
		{"foo.bar", []string{".git", "*foo*", "bar"}, true},
		{"foo.bar", []string{".git", "foo/*", "bar"}, false},
		{"foo.pdf", []string{".git", "foo", "bar"}, false},
		{"bar.xyz.xyz", []string{".git", "foo", "bar"}, false},
	}

	for _, s := range fixtures {
		if r := ignoreDir(s.dir, s.dirs); r != s.result {
			t.Errorf("ignoreDir Failed on %s and %s", s.dir, s.dirs)
		}
	}
}

func TestIgnoreFile(t *testing.T) {
	fixtures := []struct {
		filename  string
		ignoreExt []string
		result    bool
	}{
		{"a.db", []string{".db", ".foo", ".bar"}, true},
		{"a.foo", []string{".db", ".foo", ".bar"}, true},
		{"a.bar", []string{".db", ".foo", ".bar"}, true},
		{"a/a.db", []string{".db", ".foo", ".bar"}, true},
		{"foo", []string{".db", ".foo", ".bar"}, false},
		{"a.foo/foo", []string{".db", ".foo", ".bar"}, false},
		{"a/a.foo/foo", []string{".db", ".foo", ".bar"}, false},
		{"a.db/a.foo/a.bar/bar", []string{".db", ".foo", ".bar"}, false},
		{"foo.txt", []string{".db", ".foo", ".bar"}, false},
		{"a.foo/foo.txt", []string{".db", ".foo", ".bar"}, false},
		{"a/a.foo/foo.txt", []string{".db", ".foo", ".bar"}, false},
		{"a.db/a.foo/a.bar/bar.txt", []string{".db", ".foo", ".bar"}, false},
	}

	for _, s := range fixtures {
		if r := ignoreFile(s.filename, s.ignoreExt); r != s.result {
			t.Errorf("ignoreFile Failed on %s and %s", s.filename, s.ignoreExt)
		}
	}
}
