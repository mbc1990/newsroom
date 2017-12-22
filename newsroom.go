package main

import "fmt"
import "time"
import "strings"
import "strconv"
import "github.com/mmcdole/gofeed"

type Newsroom struct {
	Conf           *Configuration
	PostgresClient *PostgresClient
}

// Represents an individual document
type Document struct {
	Id         int             // UUID for document
	RawText    string          // Unmanipulated text
	Tokens     *[]string       // Tokenized, in order text
	BagOfWords *map[string]int // Tokenized, stopwords removed, word/count vector
}

// Get the contents of an rss feed
func (nr *Newsroom) GetFeed(feedInfo FeedInfo) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedInfo.Url)
	if err != nil {
		fmt.Println("Bad url: " + feedInfo.Url)
		return
	}
	// TODO: This should insert a timestamp from the feed, not the current time
	for _, item := range feed.Items {
		nr.PostgresClient.InsertFeedItem(feed.Title, item.Title, item.Content, item.Description, item.Link)
	}
}

// Called before a transformation, creates documents for each entry in the requested timespan
func (nr *Newsroom) GetDocuments(timespan Timespan) *[]Document {
	ret := make([]Document, 0)
	items := nr.PostgresClient.GetFeedItems()
	for _, item := range *items {
		doc := new(Document)
		doc.Id = item.Id
		doc.RawText = item.Headline
		doc.Tokens = RemoveStopWords(Tokenize(RemovePunctuation(strings.ToLower(item.Headline))))
		ret = append(ret, *doc)
	}
	return &ret
}

// Entry point for a run of transformers
func (nr *Newsroom) RunTransformations() {
	// TODO: Transformations run on *all* documents right now.
	// TODO: For this to change, the timestamp needs to come from the RSS feed
	ts := Timespan{0, 0}
	docs := nr.GetDocuments(ts)

	// TODO: This operation should be generalized over N things that implement the Transformation interface
	tih := new(TrendingInHeadlines)
	tih.Transform(docs)
}

// Periodically log database metrics for prometheus
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
			// Rerun the transformers now that we have updated data
			go nr.RunTransformations()
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
