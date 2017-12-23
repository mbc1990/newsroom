package main

import "os"
import "io/ioutil"
import "strings"
import "strconv"

// Represents an individual Article
type Article struct {
	Id         int // UUID for article
	Headline   string
	Body       string         // Unmanipulated text
	Tokens     *[]string      // Tokenized, in order text
	BagOfWords map[string]int // Tokenized, stopwords removed, word/count vector
}

// Initialize article for manipulation
func (a *Article) Initialize(item FeedItem, savedTextDir string) {
	a.Id = item.Id
	a.Headline = item.Headline
	a.PopulateBody(savedTextDir)
	a.PopulateTokens()
	a.PopulateBoW()
}

// Not all articles will have a body due to failed or queued scraping jobs
func (a *Article) HasBody() bool {
	return len(a.Body) > 0
}

// Reads the body from disk (if it exists)
func (a *Article) PopulateBody(savedTextDir string) {
	path := savedTextDir + strconv.Itoa(a.Id) + ".txt"

	// No scraped text
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return
	}
	body, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	a.Body = string(body)
}

// Tokenizes headline + body (if it exists)
func (a *Article) PopulateTokens() {
	toTokenize := a.Headline
	if a.HasBody() {
		toTokenize = toTokenize + " " + a.Body
	}
	a.Tokens = RemoveStopWords(Tokenize(RemovePunctuation(strings.ToLower(toTokenize))))
}

// Populates BagOfWords
func (a *Article) PopulateBoW() {
	a.BagOfWords = make(map[string]int)
	for _, token := range *a.Tokens {
		_, ok := a.BagOfWords[token]
		if !ok {
			a.BagOfWords[token] = 0
		}
		a.BagOfWords[token] += 1
	}
}
