# Markets Pricing Service

The Markets Pricing Service provides an omni-channel interface to mulitple market-data providers.

The service uses a pluggable architecture via the Bridge Pattern so that market data sources and and market data publisers can be swapped in and out very easily. E.g. here is a console publisher, and a Kafka publisher. They implement a common interface and are swapped in and out with one line of code change.

With this architecture there is no reason why there couldn't be a list of each and instantiate them at runtime to allow multiple sources and publishers.

TODO: Document the extensibility architecture.