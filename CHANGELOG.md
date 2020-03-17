<a name="unreleased"></a>
## [Unreleased]


<a name="v1.4.1"></a>
## [v1.4.1] - 2020-03-17
### Fix
- Remove the D2C_ prefix from viper calls

### Refactor
- rm unused files.

### Pull Requests
- Merge pull request [#19](https://github.com/jimrazmus/dir2consul/issues/19) from jimrazmus/fix-config


<a name="v1.4.0"></a>
## [v1.4.0] - 2020-03-17
### Build
- include modules in coverage check
- Remove debug logic. Use IDE debugger.

### Chore
- update changelog
- update and tidy modules

### Docs
- reference version 1.4.0
- use quotes consistently
- add link to regex docs
- fix link
- dry run requires v1.3.0

### Feat
- Switch to Viper and add verbose option.

### Refactor
- mv consul mutation to functions for future testing
- make function private
- relocate chdir for better context
- Seperate and test regex compilation

### Pull Requests
- Merge pull request [#18](https://github.com/jimrazmus/dir2consul/issues/18) from jimrazmus/overhaul
- Merge pull request [#17](https://github.com/jimrazmus/dir2consul/issues/17) from jimrazmus/delete-debug


<a name="v1.3.0"></a>
## [v1.3.0] - 2020-03-11
### Chore
- delete unused makefile

### Docs
- Make README more usable.

### Feat
- Add dry run capability. Closes [#15](https://github.com/jimrazmus/dir2consul/issues/15)

### Pull Requests
- Merge pull request [#16](https://github.com/jimrazmus/dir2consul/issues/16) from jimrazmus/dry-run


<a name="v1.2.0"></a>
## [v1.2.0] - 2020-03-10
### Build
- switch to push only ci
- update ci to use golang 1.14

### Chore
- add changelog
- go get -u && go mod tidy

### Docs
- Note the new default skip logic.
- fix formatting of regexs

### Feat
- skip hidden dirs and files
- update to golang 1.14.0

### Test
- Revise test fixtures for skip feature
- Revise tests for skipping feature
- add test for big files
- update golden files
- remove extra line
- fix and pretty up the serializer
- remove unused import
- add golden test files
- add a sample proj repo for testing
- add testing for LoadKeyValuesFromDisk
- add an ugly serializer function to support other testing

### Pull Requests
- Merge pull request [#14](https://github.com/jimrazmus/dir2consul/issues/14) from jimrazmus/skip-hidden
- Merge pull request [#13](https://github.com/jimrazmus/dir2consul/issues/13) from jimrazmus/test-big-file
- Merge pull request [#11](https://github.com/jimrazmus/dir2consul/issues/11) from jimrazmus/add-testing
- Merge pull request [#12](https://github.com/jimrazmus/dir2consul/issues/12) from jimrazmus/update-ci


<a name="v1.1.0"></a>
## [v1.1.0] - 2020-02-24
### Feat
- Use regexp for ignore functions.

### Pull Requests
- Merge pull request [#10](https://github.com/jimrazmus/dir2consul/issues/10) from jimrazmus/better-ignore


<a name="v1.0.1"></a>
## [v1.0.1] - 2020-02-24
### Fix
- Docker copy missed subdirs.


<a name="v1.0.0"></a>
## v1.0.0 - 2020-02-24
### Build
- Add labels to the docker image.

### Docs
- Add Go Report badge.
- Add contributing and licensing verbiage.

### Fix
- Be consistent with key prefix.
- Prevent exceeding Consul size limit.

### Pull Requests
- Merge pull request [#9](https://github.com/jimrazmus/dir2consul/issues/9) from jimrazmus/use-consul-api
- Merge pull request [#8](https://github.com/jimrazmus/dir2consul/issues/8) from jimrazmus/stop-large-values
- Merge pull request [#7](https://github.com/jimrazmus/dir2consul/issues/7) from jimrazmus/add-consul-kv-functions
- Merge pull request [#1](https://github.com/jimrazmus/dir2consul/issues/1) from jimrazmus/simple


[Unreleased]: https://github.com/jimrazmus/dir2consul/compare/v1.4.1...HEAD
[v1.4.1]: https://github.com/jimrazmus/dir2consul/compare/v1.4.0...v1.4.1
[v1.4.0]: https://github.com/jimrazmus/dir2consul/compare/v1.3.0...v1.4.0
[v1.3.0]: https://github.com/jimrazmus/dir2consul/compare/v1.2.0...v1.3.0
[v1.2.0]: https://github.com/jimrazmus/dir2consul/compare/v1.1.0...v1.2.0
[v1.1.0]: https://github.com/jimrazmus/dir2consul/compare/v1.0.1...v1.1.0
[v1.0.1]: https://github.com/jimrazmus/dir2consul/compare/v1.0.0...v1.0.1
