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
)

// App configuration details loaded from config file at boot.
var globalConfig struct {
	currencies    []string
	baseCurrency  string
	updatedAfter  string
	reader        string
	dataPublisher string
	dataSource    string
}

// Map that contains all of the possible publisher. A configuration determines which wil lbe instantiated.
var publisherMap = map[string]market_data_publisher.IMarketDataPublisher{
	"console-publisher": &market_data_publisher.ConsolePublisher{},
	"kafka-publisher":   &market_data_publisher.KafkaPublisher{}}

// Map that contains all of the possible data sources. A configuration determines which wil lbe instantiated.
var readerMap = map[string]market_reader.IMarketReader{
	"time-reader": &market_reader.TimerReader{}}

// Map that contains all of the possible data sources. A configuration determines which wil lbe instantiated.
var providerMap = map[string]market_data_source.IMarketDataSource{
	"simulator": &market_data_source.MarketSimulator{},
	"ecb":       &market_data_source.ECBMarketDataProvider{}}

func init() {
	fmt.Println("Initialising...")

	// Load the configuration data from the configuration file
	config := utils.ReadConfig("./config/app-config.properties")

	// TODO: Add checks to ensure the config values exist.

	globalConfig.currencies = strings.Split(config["currencies"], ",") // Split the comma separatedf list
	globalConfig.baseCurrency = config["base-currency"]
	globalConfig.updatedAfter = config["updated-after"]
	globalConfig.reader = config["reader"]
	globalConfig.dataPublisher = config["market-data-publisher"] // Which publisher will the service use?
	globalConfig.dataSource = config["market-data-source"]       // Which market data source will the service use?

	fmt.Printf("Loaded config: %s\n", globalConfig)
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
	for run {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			m := <-c // Test the channel to see if the price getter has retrieved a quote
			if m != "" {
				SendToPublisher(publisher, m)
			}
		}
	}
	quit <- 0 // Send a quit signal

	// Wait for clean termination response from the thread.
	for q := <-c; q != "done"; {
		continue
	}
	fmt.Printf("Received clean termination signal from all threads.\n")
	fmt.Printf("Exiting")
}

// Send the key/value to the instantiated Market Data Publisher
func SendToPublisher(publisher market_data_publisher.IMarketDataPublisher, priceData string) {
	arr := strings.SplitN(priceData, ",", 2)

	// TODO: Iterate through the list of publishers

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
