package main

import "os"
import "io/ioutil"
import "encoding/json"
import "strings"
import "strconv"

// Represents an individual Article
type Article struct {
	Id               int // UUID for article
	Headline         string
	Body             string         // Unmanipulated text
	Tokens           *[]string      // Tokenized, in order text
	BagOfWords       map[string]int // Tokenized, stopwords removed, word/count vector
	TokenizerVersion int            // Used to invalidate saved tokens made with an old version
}

// Initialize article for manipulation
func (a *Article) Initialize(item FeedItem, savedTextDir string) {
	a.TokenizerVersion = 1
	a.Id = item.Id
	a.Headline = item.Headline
	a.PopulateBody(savedTextDir)
	a.PopulateTokens(savedTextDir)
	a.PopulateBoW(savedTextDir)
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
func (a *Article) PopulateTokens(savedTextDir string) {
	// If we already tokenized this article, read it from disk
	path := savedTextDir + strconv.Itoa(a.Id) + ".tokens." + strconv.Itoa(a.TokenizerVersion) + ".json"
	if _, err := os.Stat(path); err == nil {
		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}
		var tokens []string
		json.Unmarshal(bytes, &tokens)
		a.Tokens = &tokens
		return
	}

	// Tokens aren't cached, so tokenize the article and write the tokens to disk
	toTokenize := a.Headline
	if a.HasBody() {
		toTokenize = toTokenize + " " + a.Body
	}
	a.Tokens = RemoveStopWords(Tokenize(RemovePunctuation(strings.ToLower(toTokenize))))
	bytes, err := json.Marshal(a.Tokens)
	if err != nil {
		panic(err)
	}
	file, err := os.Create(path)
	defer file.Close()
	if err != nil {
		panic(err)
	}
	file.WriteString(string(bytes))
}

// Populates BagOfWords
func (a *Article) PopulateBoW(savedTextDir string) {
	// If the BoW is saved, pull it from disk
	path := savedTextDir + strconv.Itoa(a.Id) + ".bow." + strconv.Itoa(a.TokenizerVersion) + ".json"
	if _, err := os.Stat(path); err == nil {
		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}
		bow := make(map[string]int)
		json.Unmarshal(bytes, &bow)
		a.BagOfWords = bow
		return
	}

	// Otherwise, create it and save it
	a.BagOfWords = make(map[string]int)
	for _, token := range *a.Tokens {
		_, ok := a.BagOfWords[token]
		if !ok {
			a.BagOfWords[token] = 0
		}
		a.BagOfWords[token] += 1
	}

	bytes, err := json.Marshal(a.BagOfWords)
	if err != nil {
		panic(err)
	}
	file, err := os.Create(path)
	defer file.Close()
	if err != nil {
		panic(err)
	}
	file.WriteString(string(bytes))
}
