package main

import "github.com/prometheus/client_golang/prometheus"

var feedItemsGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "num_feed_items",
	Help: "Number of items collected",
})

var cryptoGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "crypto_mentions",
	Help: "Number of crypto keyword mentions in headlines in the last 6 hours",
})
