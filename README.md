# dir2consul

[![Go Report Card](https://goreportcard.com/badge/github.com/jimrazmus/dir2consul)](https://goreportcard.com/report/github.com/jimraazmus/dir2consul)
![Build](https://github.com/jimrazmus/dir2consul/workflows/Go/badge.svg?branch=master)
[![CodeCov](https://codecov.io/gh/jimrazmus/dir2consul/branch/master/graph/badge.svg)](https://codecov.io/gh/jimrazmus/dir2consul)
[![License](http://img.shields.io/:license-mit-blue.svg?style=flat-square)](http://badges.mit-license.org)

## Summary

dir2consul mirrors a file directory to a Consul Key-Value (KV) Store

A files path and name, with the file extension removed, becomes the Consul Key while the contents of the file are the Value. Note that mirroring is exact which includes removing any Consul Keys that are not present in the source files. Hidden files and directories, those beginning with ".", are always skipped.

## Configuration

dir2consul uses environment variables to override default configuration values. The variables are:

* D2C_CONSUL_KEY_PREFIX is the path prefix to prepend to all consul keys. Default: ""
* D2C_DIRECTORY is the directory we should walk. Default: local
* D2C_IGNORE_DIR_REGEX is a PCRE regular expression that matches directories we ignore when walking the file system. Default: "a^"
* D2C_IGNORE_FILE_REGEX is a PCRE regular expression that matches files we ignore when walking the file system. Default: "README.md"

Consul specific configuration variables are documented [here](https://www.consul.io/docs/commands/index.html#environment-variables).

## Running with Docker

```
docker run --env-file=.env jimrazmus/dir2consul:v1.2.0
```

## Vault Policy

dir2consul needs a Vault policy that allows the service to modify Consul KV data.

*example policy TBD*

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests to us.

## Author

Jim Razmus II

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.
