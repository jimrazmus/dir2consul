# dir2consul

dir2consul mirrors a file directory to a Consul Key-Value (KV) Store

A files path and name, with the file extension removed, becomes the Consul Key while the contents of the file are the Value. Note that mirroring is exact which includes removing any Consul Keys that are not present in the source files.

## Configuration

dir2consul uses environment variables to override default configuration values. The variables are:

* D2C_CONSUL_KEY_PREFIX is the path prefix to prepend to all consul keys. Default: ""
* D2C_CONSUL_SERVER is the URL of the Consul server. Default: http://localhost:8500
* D2C_DIRECTORY is the directory we should walk. Default: local
* D2C_IGNORE_DIRS is a comma delimited list of directory patterns to ignore when walking the file system. Reference filepath.Match for pattern syntax. Default: .git
* D2C_IGNORE_TYPES is a comma delimited list of file suffixes to ignore when walking the file system. Default: ""
* VAULT_TOKEN is the token value used to access the Consul server. Default: ""

## Running with Docker

TBD

## Vault Policy

dir2consul needs a Vault policy that allows the service to modify Consul KV data.

*example policy TBD*
