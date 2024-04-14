# monit_exporter

<!-- shields.io -->
[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![MIT License][license-shield]][license-url]

## Index

* [About](#about)
  * [Features](#features)
  * [Support](#support)
  * [Dependencies](#dependencies)
* [Setup](#setup)
  * [Requirements](#requirements)
  * [Installation](#installation)
  * [Update](#update)
* [Usage](#usage)
* [License](#license)
* [Credits](#credits)
* [Appendix](#appendix)

## About

monit_exporter periodically scrapes the monit status and provides its data via HTTP to Prometheus.

### Features

#### Exported metrics

These metrics are exported by `monit_exporter`:

| name                                   | description                                                                                                                                                                                                                                        |
|----------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| monit_service_check                    | Monit service check info with following labels provided:<br><dl><dt>`check_name`</dt><dd>Name of monit check</dd><dt>`monitored`</dt><dd>Specifies, if the service is monitored or not, whereas `0` means no and `1` means yes.</dd><dt>`type`</dt><dd>Specifies the type of service.</dd></dl>
| monit_service_cpu_perc                 | Monit service CPU info with following labels:<br><dl><dt>`check_name`</dt><dd>Name of monit check</dd><dt>`type`</dt><dd>Specifies value type whereas value can be `percentage` or `percentage_total`</dd></dl>
| monit_service_mem_bytes                | Monit service mem info with following labels:<br><dl><dt>`check_name`</dt><dd>Name of monit check</dd><dt>`type`</dt><dd>Specifies value type whereas value can be `kilobyte` or `kilobyte_total`</dd></dl>
| monit_service_network_link_state       | Monit service link states<br><dl><dt>`check_name`</dt><dd>Name of monit check</dd></dl><br>Value can be either `-1` = Not available, `0` = down and `1` = up
| monit_service_network_link_statistics  | Monit service link statistics<br><dl><dt>`check_name`</dt><dd>Name of monit check</dd><dt>`direction`</dt><dd>Specifies link direction (upload / download)</dd><dt>`unit`</dt><dd>Spcifies unit of metrics (bytes, errors, packets)</dd><dt>`type`</dt><dd>Specifies the type with either now or total. Whereas now means "per second"</dd></dl>
| monit_service_port_response_times      | Monit service port and unix socket checks response times<br><dl><dt>`check_name`</dt><dd>Name of monit check</dd><dt>`hostname`</dt><dd>Specifies hostname checked</dd><dt>`path`</dt><dd>Specifies a unix socket path</dd><dt>`port`</dt><dd>Specifies port to check</dd><dt>`protocol`</dt><dd>Specifies protocol used for checking service (e.g. POP, IMAP, REDIS, etc.). Default is a RAW check.</dd><dt>`type`</dt><dd>Specifies protocol type (e.g. TCP, UDP, UNIX)</dd><dt>`uri`</dt><dd>Gives full URI for the service check including type, host and port or path.</dd></dl>
| monit_service_read_bytes               | Monit service Disk Read Bytes<dl><dt>`check_name`</dt><dd>Name of monit check</dd><dt>`type`</dt><dd>Specifies type of read / write. Possible values: read_count, read_count_total. Value is given in bytes.</dd></dl>
| monit_service_write_bytes              | Monit service Disk Writes Bytes<dl><dt>`check_name`</dt><dd>Name of monit check</dd><dt>`type`</dt><dd>Specifies type of read / write. Possible values: write_count, write_count_total. Value is given in bytes.</dd></dl>
| monit_up                               | Monit status availability. `0` = not available and `1` = available


#### Service types

Services type provided correspond to the XML structure of monit:

| type id | type name  |
|---------|------------|
| 0       | filesystem |
| 1       | directory  |
| 2       | file       |
| 3       | process    |
| 4       | host       |
| 5       | system     |
| 6       | fifo       |
| 7       | program    |
| 8       | network    |

### Support

### Dependencies

This application is written in Go and has the following dependencies:
* [client_golang](github.com/prometheus/client_golang/prometheus)
* [logrus](github.com/sirupsen/logrus)
* [viper](github.com/spf13/viper)
* [charset](golang.org/x/net/html/charset)

## Setup

### Requirements

This application has the following build requirements:
* Git
* Go

### Installation

#### From source

To build the application from source, simply run the following commands:
```
git clone https://github.com/monofox/monit_exporter.git
cd monit_exporter
go build
```

#### Docker Image

The preferred way to use `monit_exporter` is by running the provided Docker image. It is currently provided on GitHub Container Registry:

- [`ghcr.io/monofox/monit_exporter`](https://github.com/monofox/monit_exporter/pkgs/container/monit_exporter)

The following tags are available:

- `x.y.z` pointing to the release with that version
- `latest` pointing to the most recent released version
- `master` pointing to the latest build from the default branch


#### Scrape configuration

The exporter will query the monit server every time it is scraped by prometheus. 
If you want to reduce load on the monit server you need to change the scrape interval accordingly:

```yml
scrape_configs:
  - job_name: 'monit'
    scrape_interval: 90s
    static_configs:
      - targets: ['localhost:9338']
```

### Update

To rebuild the application with the latest Go release, execute the following commands:
```
export GO_VERSION="1.21.3"
cd ~
curl --location https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz --output ~/go${GO_VERSION}.linux-amd64.tar.gz
tar xzvf ~/go${GO_VERSION}.linux-amd64.tar.gz
sudo rm -rf /usr/local/src/go
sudo mv ~/go /usr/local/src
echo 'PATH=$PATH:/usr/local/src/go/bin:$GOPATH/bin' | sudo tee /etc/profile.d/go.sh
source /etc/profile.d/go.sh

git clone https://github.com/monofox/monit_exporter.git
cd ./monit_exporter/
rm -f go\.mod  go\.sum
sed -i 's@go-version: .*@go-version: ${GO_VERSION}@g' .github/workflows/release.yml
go mod init github.com/monofox/monit_exporter
go mod tidy
```

## Usage

The application will load the `config.toml` file located in the same directory if present. 
Use the `-conf` flag to override the default configuration file name and location.

To run the application, simply execute the Go binary.

### Parameters

Config parameter   | Environment equivalent   | Description                          | Type    | Default
------------------ | ------------------------ | ------------------------------------ | ------- | ----------------------------------------------------
`listen_address`   | `MONIT_LISTEN_ADDRESS`   | address and port to bind             | String  | 0.0.0.0:9388
`metrics_path`     | `MONIT_METRICS_PATH`     | relative path to expose metrics      | String  | /metrics
`ignore_ssl`       | `MONIT_IGNORE_SSL`       | whether of not to ignore ssl errors  | Boolean | false
`monit_scrape_uri` | `MONIT_MONIT_SCRAPE_URI` | uri to get monit status              | String  | http://localhost:2812/_status?format=xml&level=full
`monit_user`       | `MONIT_MONIT_USER`       | user for monit basic auth, if needed | String  | none
`monit_password`   | `MONIT_MONIT_PASSWORD`   | password for monit status, if needed | String  | none

### Example config
```toml
listen_address = "0.0.0.0:9388"
metrics_path = "/metrics"
ignore_ssl = false
monit_scrape_uri = "https://localhost:2812/_status?format=xml&level=full"
monit_user = "monit"
monit_password = "example-secret"
```

## License

Distributed under the MIT License.

See `LICENSE` file for more information.

## Contact

Project: [monit_exporter](https://github.com/monofox/monit_exporter)

## Credits

Acknowledgements:
* [commercetools](https://github.com/commercetools/monit_exporter)
* [delucks](https://github.com/delucks/monit_exporter)
* [chaordic](https://github.com/chaordic/monit_exporter)
* [liv-io](https://github.com/liv-io/monit_exporter)

<!-- shields.io -->
[contributors-shield]: https://img.shields.io/github/contributors/monofox/monit_exporter.svg?style=flat
[contributors-url]: https://github.com/monofox/monit_exporter/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/monofox/monit_exporter.svg?style=flat
[forks-url]: https://github.com/monofox/monit_exporter/network/members
[stars-shield]: https://img.shields.io/github/stars/monofox/monit_exporter.svg?style=flat
[stars-url]: https://github.com/monofox/monit_exporter/stargazers
[issues-shield]: https://img.shields.io/github/issues/monofox/monit_exporter.svg?style=flat
[issues-url]: https://github.com/monofox/monit_exporter/issues
[license-shield]: https://img.shields.io/github/license/monofox/monit_exporter.svg?style=flat
[license-url]: https://github.com/monofox/monit_exporter/blob/master/LICENSE
