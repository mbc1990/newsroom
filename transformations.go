package main

import "fmt"
import "sort"
import "time"
import "strconv"

type Timespan struct {
	Start time.Time
	End   time.Time
}

type Transformation interface {

	// Performs the actual transformation logic and puts the output somewhere
	// Typically, this would perform some kind of alert, or replace cached data somewhere
	Transform(docs *[]Document)

	// Should return a timespan representing the range across with the transformation will be applied
	GetTimespan() Timespan

	// Transformation's human readable name for logging
	GetName() string
}

// Most popular words in headlines
type TrendingInHeadlines struct{}

func (tih *TrendingInHeadlines) Transform(docs *[]Document) {
	counts := make(map[string]int)
	for _, doc := range *docs {
		for _, tok := range *doc.Tokens {
			_, ok := counts[tok]
			if !ok {
				counts[tok] = 0
			}
			counts[tok] += 1
		}
	}

	keys := make([]string, len(counts))
	i := 0
	for k := range counts {
		keys[i] = k
		i++
	}
	sort.Slice(keys, func(i, j int) bool {
		return counts[keys[i]] > counts[keys[j]]
	})

	// Handle fewer than 10 terms being available
	end := 10
	if len(keys) < end {
		end = len(keys)
	}

	for j := 0; j < end; j++ {
		fmt.Println(keys[j] + " " + strconv.Itoa(counts[keys[j]]))
	}
}

func (tih *TrendingInHeadlines) GetName() string {
	return "Trending in headlines"
}

func (tih *TrendingInHeadlines) GetTimespan() Timespan {
	// Last hour
	now := time.Now()
	then := now.Add(-1 * time.Hour)
	ts := Timespan{then, now}
	return ts
}

// Headlines mentioning cryptocurrency
type CryptoMentions struct{}

func (cm *CryptoMentions) Transform(docs *[]Document) {
	cryptoTerms := []string{
		"bitcoin",
		"btc",
		"crypto",
		"cryptocurrency",
		"cryptocurrencies",
		"blockchain",
	}
	cryptoMentions := 0
	for _, doc := range *docs {
		for _, t := range cryptoTerms {
			if Contains(&*doc.Tokens, t) {
				cryptoMentions += 1
				break
			}
		}
	}
	pctContaining := float64(cryptoMentions) / float64(len(*docs)) * 100.0
	fmt.Println(FloatToStr(pctContaining) + " percent containing crypto terms")
	cryptoGauge.Set(float64(pctContaining))
}

func (cm *CryptoMentions) GetName() string {
	return "Headlines mentioning crypto in the last 6 hours"
}

func (cm *CryptoMentions) GetTimespan() Timespan {
	// Last hour
	now := time.Now()
	then := now.Add(-1 * time.Hour)
	ts := Timespan{then, now}
	return ts
}
