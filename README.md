# Markets Pricing Service

## Overview

The Markets Pricing Service provides an omni-channel interface to mulitple market-data providers.

The service uses a pluggable architecture via the Bridge Pattern so that market data sources and and market data publisers can be swapped in and out very easily. E.g. here is a console publisher, and a Kafka publisher. They implement a common interface and are swapped in and out with one line of code change.

With this architecture there is no reason why there couldn't be a list of each and instantiate them at runtime to allow multiple sources and publishers.

TODO: Document the extensibility architecture.

Separately from the questions of the service implementation, users may want to access the data that it federates to the Data Commons.  Please review the [ecb-fx-query notebook](https://github.com/os-climate/data-platform-demo/blob/master/notebooks/ecb-fx-query.ipynb) to see how to access the data. 

## Current status

The service uses a cronjob to pull the previous days FX details.
There is one particular area for improvement:
1. Add a retry in case the ECB service times out (I have seen this a few times)
