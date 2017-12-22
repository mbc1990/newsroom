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

	// TODO: Store results somewhere such that the api can query them
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

func (tih *TrendingInHeadlines) GetTimespan() Timespan {
	// Last hour
	now := time.Now()
	then := now.Add(-1 * time.Hour)
	ts := Timespan{then, now}
	return ts
}
