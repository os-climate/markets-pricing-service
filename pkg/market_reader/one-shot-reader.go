package market_reader

import (
	"fmt"
	"os"

	"os-climate.org/market-pricing/pkg/market_data_source"
)

// TimeReader is am implementaiton of the IMarketReader. This implementation time-based reader of the market data.
// The TimeReader will request the market data from the IMarketDataSource object every "n" seconds where n is defined
// as a configurable item.
type OneShotReader struct {
	marketProvider   market_data_source.IMarketDataSource
	commsChannel     chan string
	quitChannel      chan int
	lastGetTimestamp string
}

// Initialise is an implementaiton of the base class and is used to set up the working variables for the reader.
// Used to initialise the inter-process communication channels.
// This may not be required if the esign calls for all GetFxPricing to be used as a go routine.
// In which case the channel initialisers move into the base class.
func (r *OneShotReader) Initialise(c chan string, quit chan int) {
	r.commsChannel = c
	r.quitChannel = quit

	fmt.Println(os.Args)
}

// SetMarketProvider assigns the MarketDataProvider so this implementation can request the data to be retrieved.
func (r *OneShotReader) SetMarketProvider(mds market_data_source.IMarketDataSource) {
	r.marketProvider = mds
}

// GetFxPricing initiates the retrieval of the market data from the provider. It defined the go channel for providing the results,
// a separate channel for controlling shutdown, a list of currencies to retrieve the FX details for, the base Currency for the FX,
// and a date stamp to filter the FX data on.
func (r *OneShotReader) GetFxPricing(currencies []string, baseCurrency string, updatedAfter string) {
	fmt.Println("OneShotReader::GetFxPricing()")

	r.lastGetTimestamp = updatedAfter

	r.getPricingFromMarketProvider(currencies, baseCurrency)

	r.commsChannel <- "done"
}

// getPricingFromMarketProvider calls the specific IMarketDataProvbider and processes the responses.
func (r *OneShotReader) getPricingFromMarketProvider(currencies []string, baseCurrency string) {
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

// isDryRun check is there is an os arg of "--dry-run". If there is then it returns tru. If not then it returns false.
func isDryRun() bool {
	result := false

	for _, v := range os.Args {
		if v == "--dry-run" {
			result = true
		}
	}
	return result
}
