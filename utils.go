package main

import (
	"math/rand"
	"time"
)

func copyMap(original map[string]float64) map[string]float64 {
	copied := make(map[string]float64)
	for key, value := range original {
		copied[key] = value
	}
	return copied
}

func generateRandomFloat64InRange(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

const CryptoSymbol = "SOL"
const FakeCryptoRate = 55.0 // For example 1 SOL = 55 USD
const EventIntervalMs = 100 // Emit event from EventInterval / 2 to EventInterval ms

func generateOrderBookData() OrderBookEvent {
	maxBidPrice := FakeCryptoRate + rand.Float64()
	minAskPrice := maxBidPrice + rand.Float64()

	// Simulate a potential opportunity
	// coef should be equal 1 when deal with real data
	coef := generateRandomFloat64InRange(1, 1.015)

	return OrderBookEvent{
		MaxBid: OrderBookData{
			ExchangeID: exchanges[rand.Intn(len(exchanges))],
			Price:      maxBidPrice * coef,
		},
		MinAsk: OrderBookData{
			ExchangeID: exchanges[rand.Intn(len(exchanges))],
			Price:      minAskPrice,
		},
	}
}

func watchOrderBook(channel chan OrderBookEvent) {
	defer close(channel)

	for {
		orderBookData := generateOrderBookData()
		halfInterval := EventIntervalMs / 2

		select {
		case channel <- orderBookData:
		case <-time.After(EventIntervalMs * time.Millisecond):
			// return
		}

		sleepTime := time.Duration(rand.Intn(halfInterval)+halfInterval) * time.Millisecond
		time.Sleep(sleepTime)
	}
}
