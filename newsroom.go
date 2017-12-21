package main

import "fmt"
import "time"
import "github.com/mmcdole/gofeed"

type Newsroom struct {
	Conf           *Configuration
	PostgresClient *PostgresClient
}

// Get the contents of an rss feed
// TODO: Handle deduplication
func (nr *Newsroom) GetFeed(feedInfo FeedInfo) {
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(feedInfo.Url)
	fmt.Println(feed.Title)
	for _, item := range feed.Items {
		nr.PostgresClient.InsertFeedItem(feed.Title, item.Title, item.Content, item.Description, item.Link)
	}
}

// Begin running
func (nr *Newsroom) Start() {
	idx := 0
	// Sleep for 60 seconds between runs
	// pauseDuration = time.Duration(int64(time.Second) * 60 * 30)
	// TODO: Testing, remove this line
	pauseDuration := time.Duration(int64(time.Second) * 10)
	numFeeds := len(nr.Conf.Feeds)
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
