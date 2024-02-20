package main

import "time"

var retryIntervals = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}
