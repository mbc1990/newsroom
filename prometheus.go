package main

import "github.com/prometheus/client_golang/prometheus"

var feedItemsGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "num_feed_items",
	Help: "Number of items collected",
})
