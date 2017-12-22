package main

import "fmt"
import "time"
import "strings"
import "strconv"
import "github.com/mmcdole/gofeed"

type Newsroom struct {
	Conf            *Configuration
	PostgresClient  *PostgresClient
	Transformations *[]Transformation
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
	for _, item := range feed.Items {
		nr.PostgresClient.InsertFeedItem(feed.Title, item.Title, item.Content, item.Description, item.Link)
	}
}

// Called before a transformation, creates documents for each entry in the requested timespan
func (nr *Newsroom) GetDocuments(timespan Timespan) *[]Document {
	ret := make([]Document, 0)
	items := nr.PostgresClient.GetFeedItems(timespan)
	for _, item := range *items {
		doc := new(Document)
		doc.Id = item.Id
		doc.RawText = item.Headline
		doc.Tokens = RemoveStopWords(Tokenize(RemovePunctuation(strings.ToLower(item.Headline))))
		ret = append(ret, *doc)
	}
	return &ret
}

// Entry point for a run of transformations
func (nr *Newsroom) RunTransformations() {
	for _, t := range *nr.Transformations {
		ts := t.GetTimespan()
		docs := nr.GetDocuments(ts)
		name := t.GetName()
		fmt.Println(strconv.Itoa(len(*docs)) + " documents being processed for transformation " + name)
		t.Transform(docs)
	}
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
	transformations := make([]Transformation, 0)
	// Trending in headlines
	tih := new(TrendingInHeadlines)
	transformations = append(transformations, tih)
	// (Initialize other transformations here)
	n.Transformations = &transformations
	return n
}
