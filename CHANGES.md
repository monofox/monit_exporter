# CHANGES

This file lists changes made to the Ansible role. It follows semantic versioning
guidelines. The content is sorted in reverse chronological order and formatted
to allow easy grepping by scripts.

The headers are:
- bugs
- changes
- enhancements
- features

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
