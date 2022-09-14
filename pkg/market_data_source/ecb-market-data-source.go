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
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/itchyny/gojq"
)

// ECBMarketDataProvider is an implementation of the MarketDataProvider interface.
// It uses ECB as the data rovider for retireving FX market data.
type ECBMarketDataProvider struct {
}

// The  structure that dummy market data should be returned in.
type ecbFxProviderResponse struct {
	Key             string  `json:"key"`
	Freq            string  `json:"freq"`
	Currency        string  `json:"currency"`
	CurrencyDenom   string  `json:"currency-denom"`
	Sender          string  `json:"sender"`
	ExrType         string  `json:"exr-type"`
	ExrSuffix       string  `json:"exr-suffix"`
	StartTimePeriod string  `json:"starttime-period"`
	EndTimePeriod   string  `json:"end-time-period"`
	ObsValue        float64 `json:"obs-value"`
	ObsStatus       string  `json:"obs-status"`
	TimeFormat      string  `json:"time-format"`
}

// The following constants are defined by the interface at: https://sdw-wsrest.ecb.europa.eu/help/. These are used to build the request message.
// Example request to retrieve all daily AUD and NOK FX agaonst the EUR after 1st July 2022
// https://sdw-wsrest.ecb.europa.eu/service/data/EXR/D.AUD+NOK.EUR.SP00.A?updatedAfter=2022-07-01T00%3A00%3A00%2B01%3A00
const wsEntryPoint = "https://sdw-wsrest.ecb.europa.eu/service" // The web service entry point:
const resource = "data"                                         // The resource for data queries is data
const flowRef = "EXR"                                           // A reference to the dataflow describing the data that needs to be returned.
const frequency = "D"                                           // Options are D (daily), M (monthly), A (annually). REVISIT: Maybe make this configurable...
const exRateType = "SP00"                                       // SP00 is FX
const seriesVariation = "A"                                     // Average

// Initialise is used as a kibd of "constructor" to set up any internal properties.
// It should be called as soon as the ECBMarketDataProvider is instantiated.
func (r *ECBMarketDataProvider) Initialise() {
}

// GetFxPricing is an implementaiotn of the MarketDataProvider interface. It retrieves the FX rates from the
// UCB provider and returns them as am array
func (r *ECBMarketDataProvider) GetFxPricing(currency string, baseCurrency string, updatedAfter string) []FxPriceDetails {
	var result []FxPriceDetails

	log.Printf("ECBMarketDataProvider::GetFxPricing(%s)", currency)

	// Build the http GET command.
	httpRequest := r.constructRequest(currency, baseCurrency, updatedAfter)
	log.Printf("HTTP Request: %s\n", httpRequest)

	// Retrieve the rates from the provider
	ecbJsonResp := requestData(httpRequest)
	// ECB returns a simple string if the query returns no results
	if ecbJsonResp == "No results found." {
		log.Println("WARNING: ECB returned no results. Check the date in the query (", updatedAfter, ") is correct: ")
	} else {
		// In case the date returns multiple FX rates, find the number of rates returned.
		var key, newFxMsg string
		numRates := getNumReturnedRates(ecbJsonResp)
		if numRates == -1 {
			log.Fatal("FATAL: There was an error determining the number of rates in the result set. Check the message format has not changed.")
		}

		// Build a response message for each result.
		for i := 0; i < numRates; i++ {
			key, newFxMsg = parseResponse(ecbJsonResp, i)
			// log.Printf("Reformatted result: %s\n", newFxMsg)

			// Append the message to the response
			var fxResult FxPriceDetails
			fxResult.Fx_key = key
			fxResult.Provider_resp = newFxMsg
			result = append(result, fxResult)
		}
	}

	return result
}

// getNumReturnedRates inspects the response and returns the number of resuts in the response data.
// If the number of rates and the number of dates does not match the function returns -1.
func getNumReturnedRates(ecbJsonResp string) int {
	var numRates int = -1

	// A list of all the JQueries that are used.
	queries := map[string]string{
		"count-fxrates": ".dataSets[].series.\"0:0:0:0:0\".observations | length",
		"count-dates":   ".structure.dimensions.observation[0].values | length"}

	// Run all thew JQueries to extract the data
	var input map[string]interface{}
	json.Unmarshal([]byte(ecbJsonResp), &input)

	var jsonVal interface{}

	jsonVal = queryPath(&input, queries["count-fxrates"])
	numFxRates := jsonVal.(int)

	jsonVal = queryPath(&input, queries["count-dates"])
	numDates := jsonVal.(int)

	// The number of results should always match - but just check to be sure.
	if numFxRates == numDates {
		numRates = numFxRates
	} else {
		log.Println("WARNING: the number of FX rates does not match the number of dates: ", numFxRates, " : ", numDates, ". JSON response was:", ecbJsonResp)
		numRates = -1
	}
	return numRates
}

// parseResponse extracts the FX details from the repsonse and stores it in a usable format
// that is base don the CSV formst you can download form ECB. Input params are:
// ecbJsonResp string: The json message returned from ECB
// index int: The index of the rate in the json message to return
func parseResponse(ecbJsonResp string, index int) (string, string) {
	var resp ecbFxProviderResponse

	// A list of all the JQueries that are used.
	queries := map[string]string{
		"sender":      ".header.sender.id",
		"time-period": ".dataSets[0].validFrom",

		"frequency":      ".structure.dimensions.series[] | select(.id==\"FREQ\") | .values[0].id",
		"currency":       ".structure.dimensions.series[] | select(.id==\"CURRENCY\") | .values[0].id",
		"currency-denom": ".structure.dimensions.series[] | select(.id==\"CURRENCY_DENOM\") | .values[0].id",
		"exr-type":       ".structure.dimensions.series[] | select(.id==\"EXR_TYPE\") | .values[0].id",
		"exr-suffix":     ".structure.dimensions.series[] | select(.id==\"EXR_SUFFIX\") | .values[0].id",

		"time-format": ".structure.attributes.series[] | select(.id==\"TIME_FORMAT\") | .values[0].name",
		"obs-status":  ".structure.attributes.observation[] | select(.id==\"OBS_STATUS\") | .values[0].id",

		// These queries are indexed. There is a part 1 & part2 that need concatenating
		"fx-rate-query1":    ".dataSets[0].series.\"0:0:0:0:0\".observations.",
		"fx-rate-query2":    "[0]",
		"start-date-query1": ".structure.dimensions.observation[0].values",
		"start-date-query2": ".start",
		"end-date-query1":   ".structure.dimensions.observation[0].values",
		"end-date-query2":   ".end"}

	// fmt.Println(ecbJsonResp)

	// Run all thew JQueries to extract the data
	var input map[string]interface{}
	json.Unmarshal([]byte(ecbJsonResp), &input)

	var jsonVal interface{}

	jsonVal = queryPath(&input, queries["sender"])
	resp.Sender = jsonVal.(string)

	jsonVal = queryPath(&input, queries["frequency"])
	resp.Freq = jsonVal.(string)

	jsonVal = queryPath(&input, queries["currency"])
	resp.Currency = jsonVal.(string)

	jsonVal = queryPath(&input, queries["currency-denom"])
	resp.CurrencyDenom = jsonVal.(string)

	jsonVal = queryPath(&input, queries["exr-type"])
	resp.ExrType = jsonVal.(string)

	jsonVal = queryPath(&input, queries["exr-suffix"])
	resp.ExrSuffix = jsonVal.(string)

	// Extract the FX Rate for the supplied index number
	queryString := queries["fx-rate-query1"] + "\"" + strconv.Itoa(index) + "\"" + queries["fx-rate-query2"]
	jsonVal = queryPath(&input, queryString)
	resp.ObsValue = jsonVal.(float64)

	// Extract the FX Rate for the supplied index number
	queryString = queries["start-date-query1"] + "[" + strconv.Itoa(index) + "]" + queries["start-date-query2"]
	jsonVal = queryPath(&input, queryString)
	resp.StartTimePeriod = jsonVal.(string)

	// Extract the FX Rate for the supplied index number
	queryString = queries["end-date-query1"] + "[" + strconv.Itoa(index) + "]" + queries["end-date-query2"]
	jsonVal = queryPath(&input, queryString)
	resp.EndTimePeriod = jsonVal.(string)

	jsonVal = queryPath(&input, queries["time-format"])
	resp.TimeFormat = jsonVal.(string)

	jsonVal = queryPath(&input, queries["obs-status"])
	resp.ObsStatus = jsonVal.(string)

	// // Construct the key
	resp.Key = "EXR" + "." + resp.Freq + "." + resp.Currency + "." + resp.CurrencyDenom + "." + resp.ExrType + "." + resp.ExrSuffix

	// Format into a new JSON message
	convertedJsonMsg, err := json.Marshal(resp)
	if err != nil {
		log.Fatal(err)
	}

	return resp.Key, string(convertedJsonMsg)
}

// queryPath ueries a single path in a json message. It returns an interface because the caller
// understands the context and will need to cast it to the appropriate type.
// This can be used to search for a specific value at a path, or to return a subtree that
// can be parsed further. E.g. Return a float of an array.
func queryPath(input *map[string]interface{}, queryString string) interface{} {
	var resp interface{}

	query, err := gojq.Parse(queryString)
	if err != nil {
		log.Fatal(err)
	}

	iter := query.Run(*input) // or query.RunWithContext

	// While there are more items to fetch - fetch them.
	for value, more := iter.Next(); more; value, more = iter.Next() {
		// "more" and "value" are dual-purpose result codes. In the gojq code the bool result is the
		// inverse of "done". So result is false when done and true when not done. If true, the
		// Interface{} result can be cast to a value or an Error type. So if true you should check
		// if it was an error before checking for the result.
		if err, more := value.(error); more {
			log.Fatalln(err)
		} else if value == nil {
			var e error = fmt.Errorf("query returns no result: <%s>", queryString)
			log.Fatal(e)
		} else {
			resp = value
		}
	}

	return resp
}

// requestData sendfs the fx rate request to the data provider and returns the response as a string.
func requestData(request string) string {

	// TODO: Add a Context so it will time out.
	response, err := http.Get(request)
	if err != nil {
		log.Fatal(err)
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	return string(responseData)
}

// constructRequest formats the http request message for the market-data provider.
func (r *ECBMarketDataProvider) constructRequest(currency string, baseCurrency string, updatedAfter string) string {
	request := wsEntryPoint + "/" + resource + "/" + flowRef + "/" + frequency + "." + currency + "." + baseCurrency + "." + exRateType + "." + seriesVariation

	if updatedAfter != "" {
		// TODO: Parse the time and replace : and + with %3A and %2B respectively.

		request += "?updatedAfter=" + updatedAfter
	}

	request += "&format=jsondata"

	return request
}
