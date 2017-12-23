package main

import "fmt"
import "time"

// Headlines mentioning cryptocurrency
type CryptoMentions struct{}

func (cm *CryptoMentions) Transform(docs *[]Article) {
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
