dp-observation-importer
================

* Consumes a Kafka message for each observation to be imported.
* Query the dp-import-api for the dimension option database id's
* Insert the observation into the DB

### Getting started

### Configuration


| Environment variable       | Default                  | Description
| ---------------------------| -----------------------  | ----------------------------------------------------
| BIND_ADDR                  | ":21700"                 | The port to bind to
| KAFKA_ADDRESS              | "http://localhost:9092"  | The address of the Kafka instance
| OBSERVATION_CONSUMER_GROUP | "observation-extracted"  | The Kafka consumer group to consume observation extracted events from
| OBSERVATION_CONSUMER_TOPIC | "observation-extracted"  | The Kafka topic to consume observation extracted events from
| DATABASE_ADDRESS           | ""                       | The address of the database
| IMPORT_API_URL             | "http://localhost:21800" | The URL of the import API

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2017, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
