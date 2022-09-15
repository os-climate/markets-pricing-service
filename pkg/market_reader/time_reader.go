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

package market_reader

import (
	"fmt"
	"time"

	"os-climate.org/market-pricing/pkg/market_data_source"
)

// TimeReader is am implementaiton of the IMarketReader. This implementation time-based reader of the market data.
// The TimeReader will request the market data from the IMarketDataSource object every "n" seconds where n is defined
// as a configurable item.
type TimerReader struct {
	marketProvider   market_data_source.IMarketDataSource
	commsChannel     chan string
	quitChannel      chan int
	lastGetTimestamp string
	timeDelay        int
}

// Initialise is an implementaiton of the base class and is used to set up the working variables for the reader.
// Used to initialise the inter-process communication channels.
// This may not be required if the esign calls for all GetFxPricing to be used as a go routine.
// In which case the channel initialisers move into the base class.
func (r *TimerReader) Initialise(c chan string, quit chan int) {
	r.commsChannel = c
	r.quitChannel = quit
	r.timeDelay = 120 // TODO: Replace this with a value read from the config file.
}

// SetMarketProvider initialises the specific Market Provider that the market data will be rettirved from.
func (r *TimerReader) SetMarketProvider(mds market_data_source.IMarketDataSource) {
	r.marketProvider = mds
}

// GetFxPricing implements the base class function. It uses a Go routine that retrieves pricing on
// scheduled intervals and puts the result on a channel for the main thread to pick up.
func (r *TimerReader) GetFxPricing(currencies []string, baseCurrency string, updatedAfter string) {
	fmt.Println("GetFxPricing() request for TimerReader")

	r.lastGetTimestamp = updatedAfter

	if r.commsChannel == nil || r.quitChannel == nil {
		fmt.Println("ERROR: TimeReader::GetFxPricing(): Channels not initialised.")
	} else if r.marketProvider == nil {
		fmt.Println("ERROR: TimeReader::GetFxPricing(): MarketProvider not initialised.")
	} else {
		ticker := time.NewTicker(time.Duration(r.timeDelay) * time.Second)
		for _ = range ticker.C {
			// Timer has fired. Iterate through each currency and get the FX pricing.
			for _, currency := range currencies {
				resp := r.marketProvider.GetFxPricing(currency, baseCurrency, r.lastGetTimestamp)

				// Iterate over the list of returned market prices and send each to the channel for processing in the main thread..
				for _, v := range resp {
					priceData := v.Fx_key + "," + v.Provider_resp // Comma-separated header

					select {
					case r.commsChannel <- priceData: // Send the pricing info to the main loop via the pricing channel.
						continue
					case <-r.quitChannel: // Check if a quit signal has been received. If so, tell the main loop that all thread-termination steps are done..
						fmt.Printf("Received QUIT signal.\n")
						r.commsChannel <- "done"
						return
					}
				}

				// TODO: Update the lastGetTimestamp with now and persist it.
			}
		}
		fmt.Printf("ERROR: quoteGetter() exiting the thread incorrectly")
	}

	return
}
