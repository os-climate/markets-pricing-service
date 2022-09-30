# Copyright 2022 Bryon Baker

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

clean:
	rm bin/market-pricing-svc

build:
	go build -o bin/market-pricing-svc cmd/main.go

# TOTO: Externalise the release version so it is not hard coded here and in the deployment config.
# At the moment these need to be kept in sync manually.
package: clean build
	podman build . -t quay.io/brbaker/market-pricing:v0.4.4

run:
	go run cmd/main.go

test: clean build
	echo "Tsk tsk! make test is not implemented yet."

all: clean build