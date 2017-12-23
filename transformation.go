package main

import "time"

type Timespan struct {
	Start time.Time
	End   time.Time
}

type Transformation interface {

	// Performs the actual transformation logic and puts the output somewhere
	// Typically, this would perform some kind of alert, or replace cached data somewhere
	Transform(docs *[]Article)

	// Should return a timespan representing the range across with the transformation will be applied
	GetTimespan() Timespan

	// Transformation's human readable name for logging
	GetName() string
}
