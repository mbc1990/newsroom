package main

import "github.com/prometheus/client_golang/prometheus"

var feedItemsCounter = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "feed_items",
	Help: "Number of items collected",
})

var badUrlsCounter = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "bad_feed_url",
	Help: "Error when fetching feed",
})
