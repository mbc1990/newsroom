package main

import "encoding/json"
import "fmt"
import "os"
import "net/http"
import "github.com/prometheus/client_golang/prometheus"

// Metadata about feeds that will be scraped
type FeedInfo struct {
	Url string
}

type Configuration struct {
	PGHost                        string
	PGPort                        int
	PGUser                        string
	PGPassword                    string
	PGDbname                      string
	PrometheusPort                string
	Feeds                         []FeedInfo
	FeedCollectionIntervalSeconds int
	NumScraperWorkers             int
	ScrapedTextDir                string
}

func main() {
	fmt.Println("Starting Newsroom")
	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Println("Usage: ./main <absolute path to configuration file>")
		return
	}
	file, _ := os.Open(args[0])
	decoder := json.NewDecoder(file)
	var conf = Configuration{}
	err := decoder.Decode(&conf)
	if err != nil {
		fmt.Println("error:", err)
	}

	// Instrument Prometheus
	prometheus.MustRegister(feedItemsGauge)
	prometheus.MustRegister(cryptoGauge)
	prometheus.MustRegister(bitcoinPriceGauge)
	http.Handle("/metrics", prometheus.Handler())
	go http.ListenAndServe(conf.PrometheusPort, nil)

	// Start service
	nr := NewNewsroom(&conf)
	nr.Start()
}
