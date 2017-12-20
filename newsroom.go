package main

import "fmt"
import "time"
import "github.com/mmcdole/gofeed"

type Newsroom struct {
	Conf *Configuration
}

// Get the contents of an rss feed
func (nr *Newsroom) GetFeed(feedUrl string) {
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(feedUrl)
	fmt.Println(feed.Title)

	// items := feed.Items
	/*
		// for each item
		  // Extract:
		    // Title
		    // Content
		    // Categories
		    // Description
		    // Link
		  // Store in postgres
		  // If scrape text
		    // Send link to scraper
	*/
}

// Begin running
func (nr *Newsroom) Start() {
	idx := 0
	// Sleep for 60 seconds between runs
	// pauseDuration = time.Duration(int64(time.Second) * 60 * 30)
	// TODO: Testing, remove this line
	pauseDuration := time.Duration(int64(time.Second) * 10)
	numFeeds := len(nr.Conf.FeedURLs)
	for {
		// If we've gone through everything, reset index and sleep
		if idx == numFeeds {
			idx = 0
			time.Sleep(pauseDuration)
		}
		go nr.GetFeed(nr.Conf.FeedURLs[idx])
		idx++
	}
}

func NewNewsroom(conf *Configuration) *Newsroom {
	n := new(Newsroom)
	n.Conf = conf
	return n
}
