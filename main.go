package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/k3a/html2text"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	addr     = flag.String("listen-address", ":9019", "The address to listen on for HTTP requests.")
	location = flag.String("location", "Munich", "The location of promcon")
	year     = flag.String("year", "2019", "The year of promcon")
	url      = flag.String("url", "https://promcon.io", "The URL to watch for registration")
)

type promconCollector struct {
	registrationOpen *prometheus.GaugeVec
}

func newPromconCollector() *promconCollector {
	return &promconCollector{
		registrationOpen: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "promcon_registration_open",
				Help: "PromCon registration open.",
			},
			[]string{"year", "location"},
		),
	}
}

func (c promconCollector) Describe(ch chan<- *prometheus.Desc) {
	c.registrationOpen.Describe(ch)
}

func (c promconCollector) Collect(ch chan<- prometheus.Metric) {
	content, err := getPromconContent()
	if err != nil {
		log.Println(err)
	} else {
		plain := html2text.HTML2Text(string(content))

		m := c.registrationOpen.WithLabelValues(*year, *location)
		if strings.Contains(plain, fmt.Sprintf("Registration for PromCon %s will open soon", *year)) {
			log.Printf("registration for promcon %s in %s is not open yet", *year, *location)
			m.Set(0)
		} else {
			log.Printf("registration for promcon %s in %s is open", *year, *location)
			m.Set(1)
		}
	}

	c.registrationOpen.Collect(ch)
}

func getPromconContent() ([]byte, error) {
	client := &http.Client{}
	url := fmt.Sprintf("%s/%s-%s", *url, *year, strings.ToLower(*location))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []byte{}, err
	}
	req.Header.Set("User-Agent", "promcon_registration_exporter/0.1")

	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return []byte{}, fmt.Errorf("unexpected error code: %d %s", resp.StatusCode, url)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	return body, nil
}

func main() {
	c := newPromconCollector()
	prometheus.MustRegister(c)

	flag.Parse()
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}
