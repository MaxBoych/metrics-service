package values

import "time"

var RetryIntervals = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}
