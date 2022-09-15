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

package market_data_publisher

import (
	"fmt"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaPublisher struct {
	initialised bool
}

var topic = "ECB-FX" // TODO: Make this a config file option
var kafkaProducer *kafka.Producer
var config kafka.ConfigMap

// Maps the environment vairable to the kafka property
var envMap = map[string]string{
	"KAFKA_BOOTSTRAP_SERVERS": "bootstrap.servers",
	"KAFKA_SECURITY_PROTOCOL": "security.protocol",
	"KAFKA_SASL_MECHANISMS":   "sasl.mechanisms",
	"KAFKA_SASL_USERNAME":     "sasl.username",
	"KAFKA_SASL_PASSWORD":     "sasl.password",
	"KAFKA_ACKS":              "acks",
}

// the config from environment variables and build the ConfigMap
func LoadConfigFromEnvironment() kafka.ConfigMap {
	config = make(kafka.ConfigMap)

	fmt.Printf("Loading Kafka environment variables:\n")
	for key, value := range envMap {
		env := os.Getenv(key)
		fmt.Printf("%s:%s\n", value, env)
		if env == "" {
			fmt.Printf("ERROR: Missing environment variable %s.\n", key)
			os.Exit(-1)
		}
		config[value] = env
	}

	return config
}

// Load the configuration file and establish the connection to the broker.
func (p *KafkaPublisher) Initialise() {
	configFile := "./config/kafka.properties"
	fmt.Printf("Reading config file from: %s\n", configFile)
	conf := ReadConfig(configFile)
	// conf := LoadConfigFromEnvironment()

	var err error
	kafkaProducer, err = kafka.NewProducer(&conf)

	if err != nil {
		fmt.Printf("Failed to create producer: %s", err)
		os.Exit(1)
	}

	// Go-routine to handle message delivery reports and
	// possibly other event types (errors, stats, etc)
	go func() {
		for e := range kafkaProducer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Failed to deliver message: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Produced event to topic %s: key = %-10s value = %s\n",
						*ev.TopicPartition.Topic, string(ev.Key), string(ev.Value))
				}
			}
		}
	}()

	p.initialised = true
}

func (p *KafkaPublisher) PublishPricingData(key string, data string) {
	fmt.Printf("KafkaPublisher::PublishPricingData()\n")

	if !p.initialised {
		fmt.Printf("ERROR: KafkaPublisher in not initialised.")
		os.Exit(-1)
	}

	// fmt.Printf("Key: %s\nData: %s\n", key, data)

	kafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          []byte(data),
	}, nil)

	// Wait for all messages to be delivered
	kafkaProducer.Flush(15 * 1000)
}

// Close the Kafka handle
func (p *KafkaPublisher) Cleanup() {
	kafkaProducer.Close()
}
