package main

import "github.com/prometheus/client_golang/prometheus"

var feedItemsGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "num_feed_items",
	Help: "Number of items collected",
})

var scrapedItemsGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "num_items_scraped",
	Help: "Number of items that have had their webpage scraped",
})

var scrapeQueueGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "scrape_queue_size",
	Help: "Number of items in the queue",
})

var cryptoGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "crypto_mentions",
	Help: "Number of crypto keyword mentions in headlines in the last 6 hours",
})

var bitcoinPriceGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "bitcoin_price",
	Help: "Price of bitcoin in USD on coinbase",
})

var scrapeFailureCounter = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "scrape_failure",
	Help: "When the goquery call returns non-nil error",
})
