# Changelog

All notable changes to this project will be documented in this file.

### v0.1.1

> 23 October 2022

- Fixed incompatible `ILIKE` query when using SQLite3 as DB ([#5](https://github.com/slaveofcode/hansip/issues/5))

### v0.1.0

> 22 October 2022

- Updated [Hansip Webserver v0.1.0](https://github.com/slaveofcode/hansip-webserver/releases/tag/v0.1.0)
- Support **S3** Storage
- Support **SQlite3** as a default database with [**WAL**](https://www.sqlite.org/wal.html) mode enabled.
- Support multi configuration paths, by default hansip will looking the configuration ordered like below.
  - `./config.yaml` (current directory, the directory where the binary runs)
  - `$HOME/.hansip/config.yaml` (home directory for current user)
  - `/etc/hansip/config.yaml` (system wide configuration)

### v0.1.0-alpha

> 18 September 2022

- Share file and get the shortened URL for sharing
- File encryption support using age-encryption
- Share to specific user
- ZIP password
- Download page with password protection
