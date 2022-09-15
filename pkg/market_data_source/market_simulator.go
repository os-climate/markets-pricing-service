// Copyright 2022 Bryon Baker

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package market_data_source

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

// MarketSimulator is an implementation of the IMarketDataSource. This implementation is used
// for implementing demonstrations of market rates without the need for a live connection to a market data source.
type MarketSimulator struct {
}

// The  structure that dummy market data should be returned in.
type mockFxProviderResponse struct {
	Currency     string  `json:"currency"`
	BaseCurrency string  `json:"base_currency"`
	Ask          float32 `json:"ask"`
	Bid          float32 `json:"bid"`
	Date         string  `json:"date"`
	HighAsk      float32 `json:"high_ask"`
	HighBid      float32 `json:"high_bid"`
	LowAsk       float32 `json:"low_ask"`
	LowBid       float32 `json:"low_bid"`
	Midpoint     float32 `json:"midpoint"`
}

var dummyPrice mockFxProviderResponse

var defaultFX = map[string]float32{"AUD": 0.69373, "CAD": 0.77616, "EUR": 1.02166, "JPY": 0.00733, "NZD": 0.62524, "NOK": 0.10117, "GBP": 1.20256, "SEK": 0.09804, "CHF": 1.03716}

func (r *MarketSimulator) Initialise() {
	// Set up the standard dataset for the simulation
	dummyPrice.BaseCurrency = "undefined"
	dummyPrice.Currency = "undefined"
	dummyPrice.Ask = 0.72894
	dummyPrice.Bid = 0.72890
	dummyPrice.HighAsk = 0
	dummyPrice.HighBid = 0
	dummyPrice.LowAsk = 0
	dummyPrice.LowBid = 0
	dummyPrice.Midpoint = 0
}

func (r *MarketSimulator) GetFxPricing(currency string, baseCurrency string, updatedAfter string) []FxPriceDetails {
	fmt.Println("GetFxPricing() requested for MarketSimulator")

	var mockResponses []FxPriceDetails

	t := time.Now()

	dummyPrice.Date = t.UTC().Format(time.UnixDate)
	dummyPrice.BaseCurrency = baseCurrency

	var mockPrice FxPriceDetails

	dummyPrice.Currency = currency

	r.simulateMarketFluctuations(dummyPrice.Currency)

	mockPrice.Fx_key = dummyPrice.BaseCurrency + "_" + dummyPrice.Currency

	// Convert to json
	jsonPrice, _ := json.Marshal(dummyPrice)
	mockPrice.Provider_resp = string(jsonPrice)

	mockResponses = append(mockResponses, mockPrice)

	return mockResponses
}

func (r *MarketSimulator) simulateMarketFluctuations(currency string) {

	ask, ok := defaultFX[currency]
	if !ok {
		ask = 0.75
	}

	dummyPrice.Ask = ask + (rand.Float32()-0.5)/100
	dummyPrice.Bid = dummyPrice.Ask - 0.00002

	if dummyPrice.Bid > dummyPrice.HighBid || dummyPrice.HighBid == 0 {
		dummyPrice.HighBid = dummyPrice.Bid
	}
	if dummyPrice.Ask > dummyPrice.HighAsk || dummyPrice.HighAsk == 0 {
		dummyPrice.HighAsk = dummyPrice.Ask
	}
	if dummyPrice.Ask < dummyPrice.LowAsk || dummyPrice.LowAsk == 0 {
		dummyPrice.LowAsk = dummyPrice.Ask
	}
	if dummyPrice.Bid < dummyPrice.LowBid || dummyPrice.LowBid == 0 {
		dummyPrice.LowBid = dummyPrice.Bid
	}
	dummyPrice.Midpoint = dummyPrice.Bid + (dummyPrice.Ask - dummyPrice.Bid)
}
