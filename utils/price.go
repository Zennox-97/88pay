/* price.go -- helper fuction and file for getting USD value of live SOL donations */

package utils

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

var (
	solUSDPrice     float64
	solUSDPriceTime time.Time
	priceMu         sync.RWMutex
	priceTTL        = 60 * time.Second // refresh every minute
)

// GetSolanaUSDPrice returns current SOL price in USD with simple caching.
func GetSolanaUSDPrice() float64 {
	priceMu.RLock()
	if time.Since(solUSDPriceTime) < priceTTL && solUSDPrice > 0 {
		defer priceMu.RUnlock()
		return solUSDPrice
	}
	priceMu.RUnlock()

	priceMu.Lock()
	defer priceMu.Unlock()

	// Double-check after acquiring write lock
	if time.Since(solUSDPriceTime) < priceTTL && solUSDPrice > 0 {
		return solUSDPrice
	}

	resp, err := http.Get("https://api.coingecko.com/api/v3/simple/price?ids=solana&vs_currencies=usd")
	if err != nil {
		return solUSDPrice // return last known price on error
	}
	defer resp.Body.Close()

	var data map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return solUSDPrice
	}

	if price, ok := data["solana"]["usd"]; ok {
		solUSDPrice = price
		solUSDPriceTime = time.Now()
	}
	return solUSDPrice
}
