package main

import "fmt"
import "sort"
import "strconv"

type Timespan struct {
	Start int
	End   int
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

	for j := 0; j < 10; j++ {
		fmt.Println(keys[j] + " " + strconv.Itoa(counts[keys[j]]))
	}
}

func (tih *TrendingInHeadlines) GetTimespan() Timespan {
	// TODO: Dummy values until timespan data problem is fixed
	ts := Timespan{0, 0}
	return ts
}
