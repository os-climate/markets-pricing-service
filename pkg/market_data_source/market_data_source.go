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

// Package market_data_source provides method to retrieve market dt from numberous sources.
// Each data source implements the IMarketDataSource interface.
package market_data_source

// FxPriceDetails is the standard structure that market data should be returned in.
type FxPriceDetails struct {
	Fx_key        string
	Provider_resp string
}

// IMarketDataSource defines the interface that all data sources should implement.
type IMarketDataSource interface {
	Initialise()
	GetFxPricing(currency string, baseCurrency string, updatedAfter string) []FxPriceDetails
}
