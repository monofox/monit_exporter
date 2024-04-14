package main

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/net/html/charset"
)

const (
	namespace               = "monit" // Prefix for Prometheus metrics.
	SERVICE_TYPE_FILESYSTEM = 0
	SERVICE_TYPE_DIRECTORY  = 1
	SERVICE_TYPE_FILE       = 2
	SERVICE_TYPE_PROCESS    = 3
	SERVICE_TYPE_HOST       = 4
	SERVICE_TYPE_SYSTEM     = 5
	SERVICE_TYPE_FIFO       = 6
	SERVICE_TYPE_PROGRAM    = 7
	SERVICE_TYPE_NET        = 8
)

var serviceTypes = map[int]string{
	SERVICE_TYPE_FILESYSTEM: "filesystem",
	SERVICE_TYPE_DIRECTORY:  "directory",
	SERVICE_TYPE_FILE:       "file",
	SERVICE_TYPE_PROCESS:    "process",
	SERVICE_TYPE_HOST:       "host",
	SERVICE_TYPE_SYSTEM:     "system",
	SERVICE_TYPE_FIFO:       "fifo",
	SERVICE_TYPE_PROGRAM:    "program",
	SERVICE_TYPE_NET:        "network",
}

type monitXML struct {
	MonitServices []monitService `xml:"service"`
}

// Simplified structure of monit check.
type monitService struct {
	Type         int                `xml:"type,attr"`
	Name         string             `xml:"name"`
	Status       int                `xml:"status"`
	Monitored    string             `xml:"monitor"`
	Memory       monitServiceMem    `xml:"memory"`
	CPU          monitServiceCPU    `xml:"cpu"`
	DiskWrite    monitServiceDisk   `xml:"write"`
	DiskRead     monitServiceDisk   `xml:"read"`
	ServiceTimes monitServiceTime   `xml:"servicetime"`
	Ports        []monitServicePort `xml:"port"`
	UnixSockets  []monitServicePort `xml:"unix"`
	Link         monitServiceLink   `xml:"link"`
}

type monitServiceMem struct {
	Percent       float64 `xml:"percent,attr"`
	PercentTotal  float64 `xml:"percenttotal"`
	Kilobyte      int     `xml:"kilobyte"`
	KilobyteTotal int     `xml:"kilobytetotal"`
}

type monitServiceCPU struct {
	Percent      float64 `xml:"percent,attr"`
	PercentTotal float64 `xml:"percenttotal"`
}

type monitServiceDisk struct {
	Bytes monitBytes `xml:"bytes"`
}

type monitServiceTime struct {
	Read  float64 `xml:"read"`
	Write float64 `xml:"write"`
	Wait  float64 `xml:"wait"`
	Run   float64 `xml:"run"`
}

type monitServicePort struct {
	Hostname     string  `xml:"hostname"`
	Path         string  `xml:"path"`
	Portnumber   string  `xml:"portnumber"`
	Protocol     string  `xml:"protocol"`
	Type         string  `xml:"type"`
	Responsetime float64 `xml:"responsetime"`
}

type monitServiceLink struct {
	State    int                       `xml:"state"`
	Speed    int                       `xml:"speed"`
	Duplex   int                       `xml:"duplex"`
	Download monitServiceLinkDirection `xml:"download"`
	Upload   monitServiceLinkDirection `xml:"upload"`
}

type monitServiceLinkDirection struct {
	Packets monitNetworkCount `xml:"packets"`
	Bytes   monitNetworkCount `xml:"bytes"`
	Errors  monitNetworkCount `xml:"errors"`
}

type monitBytes struct {
	Count int `xml:"count"`
	Total int `xml:"total"`
}
type monitNetworkCount struct {
	Now   int `xml:"now"`
	Total int `xml:"total"`
}

// Exporter collects monit stats from the given URI and exports them using
// the prometheus metrics package.
type Exporter struct {
	config *Config
	mutex  sync.RWMutex
	client *http.Client

	up                 prometheus.Gauge
	checkStatus        *prometheus.GaugeVec
	checkMem           *prometheus.GaugeVec
	checkCPU           *prometheus.GaugeVec
	checkDiskWrite     *prometheus.GaugeVec
	checkDiskRead      *prometheus.GaugeVec
	checkPortRespTimes *prometheus.GaugeVec
	checkLinkState     *prometheus.GaugeVec
	checkLinkStats     *prometheus.GaugeVec
}

type Config struct {
	listen_address   string
	metrics_path     string
	ignore_ssl       bool
	monit_scrape_uri string
	monit_user       string
	monit_password   string
}

// FetchMonitStatus gather metrics from Monit API
func FetchMonitStatus(e *Exporter) ([]byte, error) {
	req, err := http.NewRequest("GET", e.config.monit_scrape_uri, nil)
	if err != nil {
		log.Errorf("Unable to create request: %v", err)
	}

	req.SetBasicAuth(e.config.monit_user, e.config.monit_password)
	resp, err := e.client.Do(req)
	if err != nil {
		log.Error("Unable to fetch monit status")
		return nil, err
	}
	switch resp.StatusCode {
	case 200:
	case 401:
		return nil, errors.New("authentication with monit failed")
	default:
		return nil, fmt.Errorf("monit returned %s", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Fatal("Unable to read monit status")
		return nil, err
	}
	return data, nil
}

// ParseMonitStatus parse XML data and return it to struct
func ParseMonitStatus(data []byte) (monitXML, error) {
	var statusChunk monitXML
	reader := bytes.NewReader(data)
	decoder := xml.NewDecoder(reader)

	// Parsing status results to structure
	decoder.CharsetReader = charset.NewReaderLabel
	err := decoder.Decode(&statusChunk)
	return statusChunk, err
}

// ParseConfig parse exporter binary options from command line
func ParseConfig() *Config {
	flag.String("conf", "./config.toml", "Configuration file for exporter")
	flag.Parse()

	v := viper.New()

	// Provide all configurations as environment variable as well.
	v.SetEnvPrefix(strings.ToUpper(namespace))
	v.AutomaticEnv()

	v.SetDefault("listen_address", "0.0.0.0:9388")
	v.SetDefault("metrics_path", "/metrics")
	v.SetDefault("ignore_ssl", false)
	v.SetDefault("monit_scrape_uri", "http://localhost:2812/_status?format=xml&level=full")
	v.SetDefault("monit_user", "")
	v.SetDefault("monit_password", "")
	v.SetConfigFile(flag.Lookup("conf").Value.String())
	v.SetConfigType("toml")

	err := v.ReadInConfig() // Find and read the config file
	if err != nil {         // Handle errors reading the config file
		log.Printf("Error reading config file: %s. Using defaults.", err)
	}

	return &Config{
		listen_address:   v.GetString("listen_address"),
		metrics_path:     v.GetString("metrics_path"),
		ignore_ssl:       v.GetBool("ignore_ssl"),
		monit_scrape_uri: v.GetString("monit_scrape_uri"),
		monit_user:       v.GetString("monit_user"),
		monit_password:   v.GetString("monit_password"),
	}
}

// Returns an initialized Exporter.
func NewExporter(c *Config) (*Exporter, error) {

	return &Exporter{
		config: c,
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: c.ignore_ssl},
			},
		},
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Monit status availability",
		}),
		checkStatus: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "service_check",
			Help:      "Monit service check info",
		},
			[]string{"check_name", "type", "monitored"},
		),
		checkMem: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "service_mem_bytes",
			Help:      "Monit service mem info",
		},
			[]string{"check_name", "type"},
		),
		checkCPU: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "service_cpu_perc",
			Help:      "Monit service CPU info",
		},
			[]string{"check_name", "type"},
		),
		checkDiskWrite: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "service_write_bytes",
			Help:      "Monit service Disk Writes Bytes",
		},
			[]string{"check_name", "type"},
		),
		checkDiskRead: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "service_read_bytes",
			Help:      "Monit service Disk Read Bytes",
		},
			[]string{"check_name", "type"},
		),
		checkPortRespTimes: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "service_port_response_times",
			Help:      "Monit service port and unix socket checks response times",
		},
			[]string{"check_name", "hostname", "path", "port", "protocol", "type", "uri"},
		),
		checkLinkState: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "service_network_link_state",
			Help:      "Monit service link states",
		},
			[]string{"check_name"},
		),
		checkLinkStats: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "service_network_link_statistics",
			Help:      "Monit service link statistics",
		},
			[]string{"check_name", "direction", "unit", "type"},
		),
	}, nil
}

// Describe describes all the metrics ever exported by the monit exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.up.Describe(ch)
	e.checkStatus.Describe(ch)
	e.checkCPU.Describe(ch)
	e.checkMem.Describe(ch)
	e.checkDiskWrite.Describe(ch)
	e.checkDiskRead.Describe(ch)
	e.checkPortRespTimes.Describe(ch)
	e.checkLinkState.Describe(ch)
	e.checkLinkStats.Describe(ch)
}

func (e *Exporter) scrape() error {
	data, err := FetchMonitStatus(e)
	if err != nil {
		// set "monit_exporter_up" gauge to 0, remove previous metrics from e.checkStatus vector
		e.up.Set(0)
		e.checkStatus.Reset()
		log.Errorf("Error getting monit status: %v", err)
		return err
	} else {
		parsedData, err := ParseMonitStatus(data)
		if err != nil {
			e.up.Set(0)
			e.checkStatus.Reset()
			log.Errorf("Error parsing data from monit: %v\n%s", err, data)
		} else {
			e.up.Set(1)
			// Constructing metrics
			for _, service := range parsedData.MonitServices {
				e.checkStatus.With(prometheus.Labels{"check_name": service.Name, "type": serviceTypes[service.Type], "monitored": service.Monitored}).Set(float64(service.Status))
				e.checkStatus.With(
					prometheus.Labels{
						"check_name": service.Name,
						"type":       serviceTypes[service.Type],
						"monitored":  service.Monitored,
					}).Set(float64(service.Status))

				// Memory + CPU only for specifiy status types (cf. monit/xml.c)
				if service.Type == SERVICE_TYPE_PROCESS || service.Type == SERVICE_TYPE_SYSTEM {
					e.checkMem.With(
						prometheus.Labels{
							"check_name": service.Name,
							"type":       "kilobyte",
						}).Set(float64(service.Memory.Kilobyte * 1024))
					e.checkMem.With(
						prometheus.Labels{
							"check_name": service.Name,
							"type":       "kilobyte_total",
						}).Set(float64(service.Memory.KilobyteTotal * 1024))
					e.checkCPU.With(
						prometheus.Labels{
							"check_name": service.Name,
							"type":       "percentage",
						}).Set(float64(service.CPU.Percent))
					e.checkCPU.With(
						prometheus.Labels{
							"check_name": service.Name,
							"type":       "percentage_total",
						}).Set(float64(service.CPU.PercentTotal))
				}
				if service.Type == SERVICE_TYPE_PROCESS || service.Type == SERVICE_TYPE_FILESYSTEM {
					e.checkDiskWrite.With(
						prometheus.Labels{
							"check_name": service.Name,
							"type":       "write_count",
						}).Set(float64(service.DiskWrite.Bytes.Count))
					e.checkDiskWrite.With(
						prometheus.Labels{
							"check_name": service.Name,
							"type":       "write_count_total",
						}).Set(float64(service.DiskWrite.Bytes.Total))
					e.checkDiskRead.With(
						prometheus.Labels{
							"check_name": service.Name,
							"type":       "read_count",
						}).Set(float64(service.DiskRead.Bytes.Count))
					e.checkDiskRead.With(
						prometheus.Labels{
							"check_name": service.Name,
							"type":       "read_count_total",
						}).Set(float64(service.DiskRead.Bytes.Total))
				}

				// Link (only relevant for network checks)
				if service.Type == SERVICE_TYPE_NET {
					e.checkLinkState.With(
						prometheus.Labels{
							"check_name": service.Name,
						}).Set(float64(service.Link.State))
					e.addNetLinkElement(&service, "download", &service.Link.Download)
					e.addNetLinkElement(&service, "upload", &service.Link.Upload)
				}

				// Port checks
				for _, port := range service.Ports {
					var uri = fmt.Sprintf("%s://%s:%s", strings.ToLower(port.Type), port.Hostname, port.Portnumber)
					e.checkPortRespTimes.With(
						prometheus.Labels{
							"check_name": service.Name,
							"type":       port.Type,
							"hostname":   port.Hostname,
							"path":       "",
							"port":       port.Portnumber,
							"protocol":   port.Protocol,
							"uri":        uri,
						}).Set(float64(port.Responsetime))
				}

				// Unix socket checks
				for _, port := range service.UnixSockets {
					var uri = fmt.Sprintf("unix://%s", port.Path)
					e.checkPortRespTimes.With(
						prometheus.Labels{
							"check_name": service.Name,
							"type":       "UNIX",
							"hostname":   "",
							"path":       port.Path,
							"port":       "",
							"protocol":   port.Protocol,
							"uri":        uri,
						}).Set(float64(port.Responsetime))
				}
			}
		}
		return err
	}
}

func (e *Exporter) addNetLinkElement(service *monitService, direction string, lnk *monitServiceLinkDirection) {
	e.addNetLinkUnitElement(service, direction, "packets", &lnk.Packets)
	e.addNetLinkUnitElement(service, direction, "bytes", &lnk.Bytes)
	e.addNetLinkUnitElement(service, direction, "errors", &lnk.Errors)
}

func (e *Exporter) addNetLinkUnitElement(service *monitService, direction string, unit string, lnk *monitNetworkCount) {
	e.checkLinkStats.With(
		prometheus.Labels{
			"check_name": service.Name,
			"direction":  direction,
			"unit":       unit,
			"type":       "now",
		}).Set(float64(lnk.Now))
	e.checkLinkStats.With(
		prometheus.Labels{
			"check_name": service.Name,
			"direction":  direction,
			"unit":       unit,
			"type":       "total",
		}).Set(float64(lnk.Total))
}

// Collect fetches the stats from configured monit location and delivers them
// as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // Protect metrics from concurrent collects.
	defer e.mutex.Unlock()
	e.checkStatus.Reset()
	if err := e.scrape(); err == nil {
		e.up.Collect(ch)
		e.checkStatus.Collect(ch)
		e.checkMem.Collect(ch)
		e.checkCPU.Collect(ch)
		e.checkDiskWrite.Collect(ch)
		e.checkDiskRead.Collect(ch)
		e.checkPortRespTimes.Collect(ch)
		e.checkLinkState.Collect(ch)
		e.checkLinkStats.Collect(ch)
	}
}

func main() {

	config := ParseConfig()
	exporter, err := NewExporter(config)

	if err != nil {
		log.Fatal(err)
	}
	prometheus.MustRegister(exporter)

	log.Printf("Starting monit_exporter: %s", config.listen_address)
	http.Handle(config.metrics_path, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
			<head><title>Monit Exporter</title></head>
			<body>
			<h1>Monit Exporter</h1>
			<p><a href="` + config.metrics_path + `">Metrics</a></p>
			</body>
			</html>`))

		if err != nil {
			log.Fatal(err)
		}
	})

	log.Fatal(http.ListenAndServe(config.listen_address, nil))
}
