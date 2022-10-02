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

package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"os-climate.org/market-pricing/pkg/market_data_publisher"
	"os-climate.org/market-pricing/pkg/market_data_source"
	"os-climate.org/market-pricing/pkg/market_reader"
	"os-climate.org/market-pricing/pkg/utils"

	"github.com/jessevdk/go-flags"
)

// App configuration details loaded from config file at boot.
var globalConfig struct {
	currencies    []string
	baseCurrency  string
	updatedAfter  string
	reader        string
	dataPublisher string
	dataSource    string
	dryRun        bool
}

// Map that contains all of the possible publisher. A configuration determines which wil lbe instantiated.
var publisherMap = map[string]market_data_publisher.IMarketDataPublisher{
	"console-publisher": &market_data_publisher.ConsolePublisher{},
	"kafka-publisher":   &market_data_publisher.KafkaPublisher{}}

// Map that contains all of the possible data sources. A configuration determines which wil lbe instantiated.
var readerMap = map[string]market_reader.IMarketReader{
	"time-reader": &market_reader.TimerReader{},
	"one-shot":    &market_reader.OneShotReader{}}

// Map that contains all of the possible data sources. A configuration determines which wil lbe instantiated.
var providerMap = map[string]market_data_source.IMarketDataSource{
	"simulator": &market_data_source.MarketSimulator{},
	"ecb":       &market_data_source.ECBMarketDataProvider{}}

func init() {
	fmt.Println("Initialising...")

	parseCommandLineArgs()

	// Load the configuration data from the configuration file
	// TODO: Add check to make sure the configuraiton item is valid
	config := utils.ReadConfig("./config/app-config.properties")
	globalConfig.dataSource = config["market-data-source"] // Which market data source will the service use?
	globalConfig.reader = config["reader"]
	globalConfig.dataPublisher = config["market-data-publisher"] // Which publisher will the service use?

	if globalConfig.dryRun {
		// Override the configuration file if the command line switch is --dry-run
		globalConfig.dataPublisher = "console-publisher"
	}

	fmt.Printf("Loaded config: %v\n", globalConfig)
}

func main() {
	// Create a function thaty will be called on program exit so you cam close file handles etc.
	defer func() {
		cleanup()
	}()

	// Set up a channel for handling Ctrl-C, etc
	sigchan := make(chan os.Signal, 1)
	c := make(chan string) // Channel for passing pricing information
	quit := make(chan int) // Channel for sending quit signals.
	defer close(sigchan)
	defer close(c)
	defer close(quit)

	provider, exists := providerMap[globalConfig.dataSource]
	if !exists {
		optionList := ""
		for k := range providerMap {
			optionList += k + " "
		}
		var err error = fmt.Errorf("specified market data source (%s) does not exist. Cannot instantiate the publisher. Options are: %s", globalConfig.dataSource, optionList)
		log.Fatal(err)
	}
	provider.Initialise()

	// Instantiate and initialise the Market Reader(s)
	// TODO: Add error handling
	reader, exists := readerMap[globalConfig.reader] // &market_reader.TimerReader{}
	if !exists {
		optionList := ""
		for k := range readerMap {
			optionList += k + " "
		}
		var err error = fmt.Errorf("specified reader (%s) does not exist. Cannot instantiate the market reader. Options are: %s", globalConfig.reader, optionList)
		log.Fatal(err)
	}

	reader.Initialise(c, quit)
	reader.SetMarketProvider(provider)

	// Instantiate and initialise the Market Publisher fro the global configuration data
	publisher, exists := publisherMap[globalConfig.dataPublisher]
	if !exists {
		optionList := ""
		for k := range publisherMap {
			optionList += k + " "
		}
		var err error = fmt.Errorf("specified market publisher (%s) does not exist. Cannot instantiate the publisher. Options are: %s", globalConfig.dataPublisher, optionList)
		log.Fatal(err)
	}
	publisher.Initialise()

	// Start the reader thread
	go reader.GetFxPricing(globalConfig.currencies, globalConfig.baseCurrency, globalConfig.updatedAfter)

	// Process messages
	run := true
loop:
	for run {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			m := <-c         // Test the channel to see if the price getter has retrieved a quote
			if m == "done" { // Check if the market reader is done.
				break loop
			} else if m != "" {
				SendToPublisher(publisher, m)
			}
		}
	}

	fmt.Printf("Exiting")
}

// Send the key/value to the instantiated Market Data Publisher
func SendToPublisher(publisher market_data_publisher.IMarketDataPublisher, priceData string) {
	arr := strings.SplitN(priceData, ",", 2)

	// Check the data is formatted properly
	if len(arr) == 2 {
		publisher.PublishPricingData(arr[0], arr[1])
	} else {
		fmt.Printf("ERROR: Badly formatted data in SendToPublisher. No comma separater: %s", priceData)
	}

}

// Called on program exit. Place any cleanup functions here
func cleanup() {

}

// isDryRun check is there is an os arg of "--dry-run". If there is then it returns tru. If not then it returns false.
func parseCommandLineArgs() {
	var opts struct {
		// Slice of bool will append 'true' each time the option
		// is encountered (can be set multiple times, like -vvv)
		DryRun bool `long:"dry-run" description:"Dry run - send output to console instead of the configured market-data publisher."`

		BaseCurrency string `long:"base-currency"  description:"The base currency that all currencies are rated against." required:"true"`

		Currencies string `long:"currencies"  description:"A comma separated list of currencies to retrieve." required:"true"`

		UpdatedAfter string `long:"updated-after"  description:"The earliest date to retrieve FX data from. Format YYYY-MM-DD" required:"true" default:"2022-01-01"`
	}

	_, err := flags.Parse(&opts)
	if err != nil {
		fmt.Println("Invalid command-line options. Use --help for details.")
		os.Exit(1)
	}

	fmt.Printf("Dry run: %v\n", opts.DryRun)
	fmt.Printf("Base Currency: %s\n", opts.BaseCurrency)
	fmt.Printf("Currencies: %s\n", opts.Currencies)
	fmt.Printf("Updated after: %s\n", opts.UpdatedAfter)

	globalConfig.dryRun = opts.DryRun
	globalConfig.baseCurrency = opts.BaseCurrency
	globalConfig.currencies = strings.Split(opts.Currencies, ",")
	globalConfig.updatedAfter = opts.UpdatedAfter

}
