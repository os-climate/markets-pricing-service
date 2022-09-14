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
)

type ConsolePublisher struct {
}

func (p *ConsolePublisher) PublishPricingData(key string, data string) {
	fmt.Printf("ConsolePublisher::PublishPricingData()\n")

	fmt.Printf("Key: %s\nData: %s\n", key, data)
}

func (p *ConsolePublisher) Initialise() {
	// Do nothing
}

// Clean up any resources on exit
func (p *ConsolePublisher) Cleanup() {
	// Do nothing
}
