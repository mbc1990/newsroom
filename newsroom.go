package main

import "fmt"
import "time"
import "strconv"
import "github.com/mmcdole/gofeed"

type Newsroom struct {
	Conf           *Configuration
	PostgresClient *PostgresClient
}

// Get the contents of an rss feed
func (nr *Newsroom) GetFeed(feedInfo FeedInfo) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedInfo.Url)
	if err != nil {
		fmt.Println("Bad url: " + feedInfo.Url)
		badUrlsCounter.Inc()
		return
	}
	for _, item := range feed.Items {
		nr.PostgresClient.InsertFeedItem(feed.Title, item.Title, item.Content, item.Description, item.Link)
	}
}

func (nr *Newsroom) DBMetrics() {
	for {

		items := nr.PostgresClient.GetNumFeedItems()
		feedItemsGauge.Set(float64(items))
		pauseDuration := time.Duration(int(time.Second) * nr.Conf.FeedCollectionIntervalSeconds)
		time.Sleep(pauseDuration)
	}
}

// Begin running
func (nr *Newsroom) Start() {
	idx := 0
	pauseDuration := time.Duration(int(time.Second) * nr.Conf.FeedCollectionIntervalSeconds)
	numFeeds := len(nr.Conf.Feeds)
	go nr.DBMetrics()
	fmt.Println("Collecting news from " + strconv.Itoa(numFeeds) + " sources.")
	for {
		// If we've gone through everything, reset index and sleep
		if idx == numFeeds {
			idx = 0
			time.Sleep(pauseDuration)
		}
		go nr.GetFeed(nr.Conf.Feeds[idx])
		idx++
	}
}

func NewNewsroom(conf *Configuration) *Newsroom {
	n := new(Newsroom)
	n.Conf = conf
	n.PostgresClient = NewPostgresClient(n.Conf.PGHost, n.Conf.PGPort,
		n.Conf.PGUser, n.Conf.PGPassword, n.Conf.PGDbname)
	return n
}
