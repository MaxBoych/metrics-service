package main

import "flag"

var flagRunAddr string
var flagReportInterval int
var flagPollInterval int

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.IntVar(&flagReportInterval, "r", 10, "frequency of sending metrics on the server")
	flag.IntVar(&flagPollInterval, "p", 2, "frequency of polling metrics from the 'runtime' package")
	flag.Parse()
}
