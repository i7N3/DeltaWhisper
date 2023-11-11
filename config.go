package main

// There can be ofc more than 3 exchanges
var exchanges = [3]string{"exchange1", "exchange2", "exchange3"}

var fees = map[string]Fees{
	"exchange1": {Base: 0.001, Quote: 0.003},
	"exchange2": {Base: 0.001, Quote: 0.003},
	"exchange3": {Base: 0.001, Quote: 0.003},
}

const (
	MinimumUsdGain          = 0.0
	MinimumProfitPercentage = 0.0
)
