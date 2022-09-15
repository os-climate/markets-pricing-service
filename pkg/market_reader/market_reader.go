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
	"os-climate.org/market-pricing/pkg/market_data_source"
)

// IMarketReader defines an interface for reading market data from some market data provider. The Market Reader
// is used to implement the method of triggering the read from the market data source. For example, a MarketReader
// could be a scheduled read every 5 seconds, run once, a file, or trigger on an API POST to some HTTP endpoint.
// The MarketReader is intended to operate on a separate thread so it can read market data asynchronously and post the result
// To a channel for processing in some other thread.
type IMarketReader interface {
	// SetMarketProvider assigns the MarketDataProvider so this implementation can request the data to be retrieved.
	SetMarketProvider(market_data_source.IMarketDataSource)

	// GetFxPricing initiates the retrieval of the market data from the provider. It defined the go channel for providing the results,
	// a separate channel for controlling shutdown, a list of currencies to retrieve the FX details for, the base Currency for the FX,
	// and a date stamp to filter the FX data on.
	GetFxPricing(c chan string, quit chan int, currencies []string, baseCurrency string, updatedAfter string)
}
