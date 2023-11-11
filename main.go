package main

import (
	"fmt"
	"time"
)

type OrderBookData struct {
	ExchangeID string
	Price      float64
}

type OrderBookEvent struct {
	MaxBid OrderBookData
	MinAsk OrderBookData
}

type Balances map[string]float64

type Fees struct {
	Base  float64
	Quote float64
}

// Designed taking into account the fact that funds do not move
// between exchanges but are initially located there.

// The POC doesn't have session reload mechanism
// If you need to implement it, don't forget to reset the balances
// At the beginning of the session you should have only USD balance
func main() {
	startTime := time.Now()

	// These variables can be increased for a bit as a risk management mechanism
	// It depend on the current market state, adjust it dynamically to maximum effect
	sessionTotalProfitUsd := 0.0
	sessionTotalProfitPct := 0.0

	prevMinAskPrice := 0.0
	prevMaxBidPrice := 0.0

	// Initially you should have only USD balance on every exchange
	usdBalances := Balances{
		"exchange1": 100,
		"exchange2": 100,
		"exchange3": 100,
	}
	// At the end of session move it all to USD to concrete the profits
	cryptoBalances := Balances{
		"exchange1": 0,
		"exchange2": 0,
		"exchange3": 0,
	}

	totalUsdBalance := 0.0
	for _, v := range usdBalances {
		totalUsdBalance += v
	}

	startingBuyPriceAvg := FakeCryptoRate
	totalCryptoBalance := (totalUsdBalance / 2) / startingBuyPriceAvg
	cryptoPerOperation := (totalCryptoBalance / float64(len(exchanges))) * 0.99

	// It all fake logs just to show how it should work
	fmt.Print("\nFetching USD balance ...\n")
	fmt.Printf("Total USD balance is $%.2f\n\n", totalUsdBalance)
	time.Sleep(1 * time.Second)

	fmt.Printf("Fetching the %s average price ...\n", CryptoSymbol)
	fmt.Printf("%s average price is $%.2f\n\n", CryptoSymbol, startingBuyPriceAvg)
	time.Sleep(1 * time.Second)

	// These can be either limit orders or market orders.
	// Limit orders are more profitable but they take time to be executed!
	fmt.Printf("Sending buy orders to all exchanges to get some %s ...\n", CryptoSymbol)
	fmt.Printf("Buy orders successfully filled!")
	time.Sleep(1 * time.Second)

	for _, exchange := range exchanges {
		usdBalances[exchange] = usdBalances[exchange] / 2
		cryptoBalances[exchange] = usdBalances[exchange] / FakeCryptoRate
	}

	fmt.Printf("\n\nCURRENT BALANCES\n\n")

	for _, exchange := range exchanges {
		fmt.Printf(
			"%s: %.3f %s / %.2f USDT\n",
			exchange,
			cryptoBalances[exchange],
			CryptoSymbol,
			usdBalances[exchange],
		)
	}

	fmt.Println()
	time.Sleep(1 * time.Second)

	orderBookData := make(chan OrderBookEvent)

	go watchOrderBook(orderBookData)

	fmt.Printf("Watching order book to find arbitrage opportunity!\n\n")
	time.Sleep(1 * time.Second)

	i := 0
	for data := range orderBookData {
		fmt.Printf("MinAsk [%-8s]: %-16.10f MaxBid [%-8s]: %-16.10f\n", data.MinAsk.ExchangeID, data.MinAsk.Price, data.MaxBid.ExchangeID, data.MaxBid.Price)

		baseFeeBid := fees[data.MaxBid.ExchangeID].Base
		quoteFeeBid := fees[data.MaxBid.ExchangeID].Quote
		baseFeeAsk := fees[data.MinAsk.ExchangeID].Base
		quoteFeeAsk := fees[data.MinAsk.ExchangeID].Quote

		bidBaseMultiplier := 1 + baseFeeBid
		bidQuoteMultiplier := 1 - quoteFeeBid
		askBaseMultiplier := 1 + baseFeeAsk
		askQuoteMultiplier := 1 - quoteFeeAsk

		usdReceived := cryptoPerOperation / bidBaseMultiplier * data.MaxBid.Price * bidQuoteMultiplier
		usdSpent := (cryptoPerOperation / askQuoteMultiplier) * data.MinAsk.Price * askBaseMultiplier

		copiedUsd := copyMap(usdBalances)
		copiedUsd[data.MaxBid.ExchangeID] += usdReceived
		copiedUsd[data.MinAsk.ExchangeID] -= usdSpent

		copiedCrypto := copyMap(cryptoBalances)
		copiedCrypto[data.MinAsk.ExchangeID] += cryptoPerOperation
		copiedCrypto[data.MaxBid.ExchangeID] -= cryptoPerOperation

		initialUsdCombined := usdBalances[data.MinAsk.ExchangeID] + usdBalances[data.MaxBid.ExchangeID]
		changeUsd := copiedUsd[data.MinAsk.ExchangeID] + copiedUsd[data.MaxBid.ExchangeID] - initialUsdCombined

		totalInitialUsd := 0.0
		for _, exchange := range exchanges {
			totalInitialUsd += usdBalances[exchange]
		}
		changeUsdPct := (changeUsd / totalInitialUsd) * 100

		isDifferentExchange := data.MinAsk.ExchangeID != data.MaxBid.ExchangeID
		isDifferentAskPrice := prevMinAskPrice != data.MinAsk.Price
		isDifferentBidPrice := prevMaxBidPrice != data.MaxBid.Price
		isProfitUsdCriteriaMet := changeUsd > MinimumUsdGain
		isProfitUsdPctCriteriaMet := changeUsdPct > MinimumProfitPercentage

		isSufficientUsdBalance := usdBalances[data.MinAsk.ExchangeID] >= usdSpent
		isSufficientCryptoBalance := cryptoBalances[data.MaxBid.ExchangeID] >= cryptoPerOperation

		if isDifferentExchange && isProfitUsdCriteriaMet && isProfitUsdPctCriteriaMet && isDifferentAskPrice && isDifferentBidPrice && isSufficientUsdBalance && isSufficientCryptoBalance {
			i += 1

			sessionTotalProfitUsd += changeUsd
			sessionTotalProfitPct += changeUsdPct

			cryptoMaxBidFeeAmount := cryptoPerOperation * fees[data.MaxBid.ExchangeID].Base
			cryptoMinAskFeeAmount := cryptoPerOperation * fees[data.MinAsk.ExchangeID].Quote

			maxBidUsdFees := (cryptoPerOperation * data.MaxBid.Price) * fees[data.MaxBid.ExchangeID].Quote
			minAskUsdFees := (cryptoPerOperation * data.MinAsk.Price) * fees[data.MinAsk.ExchangeID].Base

			totalCryptoFees := cryptoMaxBidFeeAmount + cryptoMinAskFeeAmount
			totalUsdFees := maxBidUsdFees + minAskUsdFees

			var exBalancesCurrent string
			var exBalancesPossible string

			for _, exchange := range exchanges {
				exBalancesCurrent += fmt.Sprintf(
					"\n• %s: %.3f %s / %.2f USDT",
					exchange,
					cryptoBalances[exchange],
					CryptoSymbol,
					usdBalances[exchange],
				)
				exBalancesPossible += fmt.Sprintf(
					"\n• %s: %.3f %s / %.2f USDT",
					exchange,
					copiedCrypto[exchange],
					CryptoSymbol,
					copiedUsd[exchange],
				)
			}

			fmt.Printf("\n🚀 Opportunity #%d Uncovered! 🚀 (Buy: %s @ %f ➡️ Sell: %s @ %f)\n\n💰 Estimated Profit: %.4f%% (%.4f USD)\n💸 Total Fees: %.4f USD and %.4f %s\n📈 Overall Session Earnings: %.4f%% (%.4f USD)\n⏰ Time Elapsed: %s\n\n🔍 CURRENT BALANCES 🔍\n%s\n\n🔮 EXPECTED BALANCES AFTER TRADE 🔮\n%s\n\n",
				i,
				data.MinAsk.ExchangeID,
				data.MinAsk.Price,
				data.MaxBid.ExchangeID,
				data.MaxBid.Price,
				changeUsdPct,
				changeUsd,
				totalUsdFees,
				totalCryptoFees,
				CryptoSymbol,
				sessionTotalProfitPct,
				sessionTotalProfitUsd,
				time.Since(startTime).String(),
				exBalancesCurrent,
				exBalancesPossible,
			)

			// These can be either limit orders or market orders.
			// Limit orders are more profitable but they take time to be executed!
			fmt.Printf("Sending buy order\n")
			fmt.Printf("Sending sell order\n")
			time.Sleep(1 * time.Second)
			fmt.Printf("Sell order successfully filled!\n")
			time.Sleep(1 * time.Second)
			fmt.Printf("Buy order successfully filled!\n\n")
			time.Sleep(1 * time.Second)

			// If you are dealing with spot remember to time limit them

			// After real execution you can fetch actual spent, recieved amounts
			// And recalculate the balances more precisely

			// Update balances and continue
			usdBalances[data.MinAsk.ExchangeID] -= usdSpent
			cryptoBalances[data.MinAsk.ExchangeID] += cryptoPerOperation
			cryptoBalances[data.MaxBid.ExchangeID] -= cryptoPerOperation
			usdBalances[data.MaxBid.ExchangeID] += usdReceived

			time.Sleep(2 * time.Second)
		}
	}
}
