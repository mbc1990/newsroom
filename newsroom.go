package main

import "fmt"
import "time"
import "strings"
import "io/ioutil"
import "encoding/json"
import "net/http"
import "strconv"
import "os"
import "log"
import "github.com/mmcdole/gofeed"
import "github.com/PuerkitoBio/goquery"

type Newsroom struct {
	Conf            *Configuration
	PostgresClient  *PostgresClient
	Transformations *[]Transformation
	ScraperJobQueue chan ScraperJob
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
		id := nr.PostgresClient.GetIdForItem(item.Title)

		// Enqueue a scraper job
		job := new(ScraperJob)
		job.ItemId = id
		job.Url = item.Link
		nr.ScraperJobQueue <- *job
		scrapeQueueGauge.Inc()
	}
}

func (nr *Newsroom) ScraperWorker() {
	for job := range nr.ScraperJobQueue {
		scrapeQueueGauge.Dec()
		doc, err := goquery.NewDocument(job.Url)
		if err != nil {
			// This happens when a URL is bad or a page is for some reason unparsable
			// So let's log it move on
			log.Print(err)
			scrapeFailureCounter.Inc()
			continue
		}
		texts := make([]string, 0)
		doc.Find("p").Each(func(index int, item *goquery.Selection) {
			text := item.Text()
			texts = append(texts, text)
		})

		// Join the scraped text with a space
		fullText := strings.Join(texts[:], " ")

		// Write the text to file
		file, err := os.Create(nr.Conf.ScrapedTextDir + strconv.Itoa(job.ItemId) + ".txt")
		defer file.Close()
		if err != nil {
			// This is bad, we should never be re-scraping
			panic(err)
		}
		file.WriteString(fullText)
		nr.PostgresClient.SetScraped(job.ItemId)
	}
}

// Called before a transformation, creates documents for each entry in the requested timespan
func (nr *Newsroom) GetArticles(timespan Timespan) *[]Article {
	ret := make([]Article, 0)
	items := nr.PostgresClient.GetFeedItems(timespan)
	for _, item := range *items {
		art := new(Article)
		art.Initialize(item, nr.Conf.ScrapedTextDir)
		ret = append(ret, *art)
	}
	return &ret
}

// These transformations are run with every feed collection interval
// They are most useful for executing some logic on a sliding window of content
func (nr *Newsroom) RunPeriodicTransformations() {
	for _, t := range *nr.Transformations {
		ts := t.GetTimespan()
		docs := nr.GetArticles(ts)
		name := t.GetName()
		fmt.Println(strconv.Itoa(len(*docs)) + " documents being processed for transformation: " + name)
		t.Transform(docs)
	}
}

// Periodically log database metrics for prometheus
func (nr *Newsroom) DBMetrics() {
	for {
		items := nr.PostgresClient.GetNumFeedItems()
		feedItemsGauge.Set(float64(items))

		scraped := nr.PostgresClient.GetScrapedItems()
		scrapedItemsGauge.Set(float64(scraped))

		pauseDuration := time.Duration(int(time.Second) * nr.Conf.FeedCollectionIntervalSeconds)
		time.Sleep(pauseDuration)
	}
}

type CoinbaseResponse struct {
	Bpi struct {
		USD struct {
			Rate string
		}
	}
}

// TODO: This should probably move to a separate service
func (nr *Newsroom) BitcoinPrice() {
	for {
		resp, err := http.Get("https://api.coindesk.com/v1/bpi/currentprice.json")
		if err != nil {
			panic(err)
		}
		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		cbResp := new(CoinbaseResponse)
		json.Unmarshal(body, &cbResp)
		noCommas := strings.Replace(cbResp.Bpi.USD.Rate, ",", "", -1)
		price, _ := strconv.ParseFloat(noCommas, 64)
		bitcoinPriceGauge.Set(price)
		pauseDuration := time.Duration(int(time.Second) * nr.Conf.FeedCollectionIntervalSeconds)
		time.Sleep(pauseDuration)
	}
}

// Used to enqueue an article to be scraped
type ScraperJob struct {
	ItemId int
	Url    string
}

func (nr *Newsroom) PopulateScrapeQueue() {
	unscrapedItems := nr.PostgresClient.GetUnscrapedJobs()
	for _, item := range unscrapedItems {
		nr.ScraperJobQueue <- item
		scrapeQueueGauge.Inc()
	}
}

// Begin running
func (nr *Newsroom) Start() {
	// Populate scraper worker pool
	for i := 0; i < nr.Conf.NumScraperWorkers; i++ {
		go nr.ScraperWorker()
	}

	// In case we restarted and lost the queue, repopulate it from the database
	nr.PopulateScrapeQueue()

	idx := 0
	pauseDuration := time.Duration(int(time.Second) * nr.Conf.FeedCollectionIntervalSeconds)
	numFeeds := len(nr.Conf.Feeds)
	go nr.BitcoinPrice()
	go nr.DBMetrics()
	fmt.Println("Collecting news from " + strconv.Itoa(numFeeds) + " sources.")
	for {
		// If we've gone through everything, reset index and sleep
		if idx == numFeeds {
			// Rerun the transformers now that we have updated data
			go nr.RunPeriodicTransformations()
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

	// Crypto watcher
	cm := new(CryptoMentions)
	transformations = append(transformations, cm)

	// (Initialize other transformations here)
	n.Transformations = &transformations

	// Queue for scraping articles
	n.ScraperJobQueue = make(chan ScraperJob)
	return n
}
