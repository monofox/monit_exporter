# CHANGES

This file lists changes made to the monit exporter. It follows semantic versioning
guidelines. The content is sorted in reverse chronological order and formatted
to allow easy grepping by scripts.

The headers are:
- Bugs
- Changes
- Enhancements
- Features

## 0.3.0 (2024-04-xx)

### Bugs
- Catching scrape errors

### Changes
- Breaking change: changing type names in compliance to Monit service type names (`programPid` => `process`, `programPath` => `program`, `remoteHost` => `host`)
- Earlier, by default the exporter listened on localhost. In future, it will listen on `0.0.0.0:9388` by default. This enhance the use within containerized environment.
- Breaking change: some metrics were renamed. E.g.:
  - `monit_exporter_up` => `monit_up`
  - `monit_exporter_service_check` => `monit_service_check`
- Upgraded third party depenencies
- Upgraded regular go version to 1.21.9

### Enhancements

- Adding Dockerfile for easier deployment.
- Providing workflow for automatic docker image generation in Github container registry.
- Extended documentation via README.md of provided features, installation, configuration.

### Features
- Adding support of environment variables for configuration.
- Added extraction of:
  - port response times
  - unix socket response times
  - CPU usage
  - Memory usage
  - Disk write metrics
  - Disk read metrics
  - I/O service times
  - Network link metrics
- Added option in order to ignore TLS certificate validation (restricted and not recommended)

## 0.2.2 (2023-10-22)

### Enhancements

- Build monit_exporter with Go version 1.21.3

## 0.2.1 (2023-07-27)

### Enhancements

- Build monit_exporter with Go version 1.20.6

## 0.2.0 (2023-04-11)

### Changes

- Rebuild monit_exporter for OpenBSD 7.3

### Enhancements

- Build monit_exporter with Go version 1.20.3
- Resolve GoReleaser deprecation notices \
  `--rm-dist` has been deprecated in favor of `--clean` \
  `replacements` will be removed from the `archives` section

## 0.1.0 (2020-11-29)

### Bugs

- Add promhttp dependency (prometheus.Handler deprecated by promhttp.Handler)
- Improve error handling if close body is not nil

### Enhancements

- Return descriptive error message with HTTP response status codes
- Remove white spaces from service types

## 0.0.2 (2017-10-16)

### Features

- Change deploy key

## 0.0.1 (2017-09-22)

### Features

- First working release
