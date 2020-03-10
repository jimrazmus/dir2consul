<a name="unreleased"></a>
## [Unreleased]


<a name="v1.2.0"></a>
## [v1.2.0] - 2020-03-10
### Build
- switch to push only ci
- update ci to use golang 1.14

### Chore
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


[Unreleased]: https://github.com/jimrazmus/dir2consul/compare/v1.2.0...HEAD
[v1.2.0]: https://github.com/jimrazmus/dir2consul/compare/v1.1.0...v1.2.0
[v1.1.0]: https://github.com/jimrazmus/dir2consul/compare/v1.0.1...v1.1.0
[v1.0.1]: https://github.com/jimrazmus/dir2consul/compare/v1.0.0...v1.0.1
