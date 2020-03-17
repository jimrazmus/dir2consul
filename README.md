# dir2consul

[![Go Report Card](https://goreportcard.com/badge/github.com/jimrazmus/dir2consul)](https://goreportcard.com/report/github.com/jimrazmus/dir2consul)
![Build](https://github.com/jimrazmus/dir2consul/workflows/Go/badge.svg?branch=master)
[![CodeCov](https://codecov.io/gh/jimrazmus/dir2consul/branch/master/graph/badge.svg)](https://codecov.io/gh/jimrazmus/dir2consul)
[![License](http://img.shields.io/:license-mit-blue.svg?style=flat-square)](http://badges.mit-license.org)

## Summary

dir2consul mirrors a filesystem directory to a Consul Key-Value (KV) Store

A files path and name, with the file extension removed, becomes the Consul Key while the contents of the file are the Value. Note that mirroring is exact which includes *deleting* any Consul Keys that are not present in the source files. Hidden files and directories, those beginning with ".", are always skipped.

## Configuration

dir2consul uses environment variables to override default configuration values. The variables are:

* D2C_CONSUL_KEY_PREFIX is the path to prepend to all Consul keys. Default: "dir2consul"
* D2C_DIRECTORY is the directory dir2consul will walk. Default: "local/repo"
* D2C_DRYRUN is a flag that prevents all Consul data modification. Set it to any truthy value to enable. Default: "false"
* D2C_IGNORE_DIR_REGEX is a PCRE regular expression that matches directories we ignore when walking the file system. The default value is impossible to match. Default: "a^"
* D2C_IGNORE_FILE_REGEX is a PCRE regular expression that matches files we ignore when walking the file system. Default: "README.md"
* D2C_VERBOSE is a flag that increases log output. Set it to any truthy value to enable. Default: "false"

Consul specific configuration variables are documented [here](https://www.consul.io/docs/commands/index.html#environment-variables) and may be used to customize dir2consul connectivity to a Consul server.

Read more about [regular expression syntax](https://github.com/google/re2/wiki/Syntax) to get the desired behavior with the D2C_IGNORE_DIR_REGEX and D2C_IGNORE_FILE_REGEX configuration options.

## Running with Docker

The following command does a dry run of mirroring the present working directory (PWD) to the Consul server KV store under the path "some/specific/kv/path".

```
docker run -v $(PWD):/local \
  --env CONSUL_HTTP_ADDR=consul.example.com:8500 \
  --env D2C_CONSUL_KEY_PREFIX=some/specific/kv/path \
  --env D2C_DRYRUN=true \
  jimrazmus/dir2consul:v1.4.1
```

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests to us.

## Author

Jim Razmus II

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.
