dp-observation-importer
================

* Consumes a Kafka message for each observation to be imported.
* Query the dp-import-api for the dimension option database id's
* Insert the observation into the DB

### Getting started

* Run ```brew install kafka```
* Run ```brew install neo4j```
* Configure neo4j, edit /usr/local/Cellar/neo4j/3.2.0/libexec/conf/neo4j.conf
  * Set ```dbms.security.auth_enabled=false```
* Run ```brew services restart neo4j```

### Configuration

| Environment variable       | Default                          | Description
| ---------------------------|--------------------------------- |-----------------------------------------------------
| BIND_ADDR                  | ":21700"                         | The port to bind to
| KAFKA_ADDRESS              | "http://localhost:9092"          | The address of the Kafka instance
| OBSERVATION_CONSUMER_GROUP | "dp-observation-importer"        | The Kafka consumer group to consume observation extracted events from
| OBSERVATION_CONSUMER_TOPIC | "observation-extracted"          | The Kafka topic to consume observation extracted events from
| DATABASE_ADDRESS           | "bolt://localhost:7687"          | The address of the database
| DATASET_API_URL            | "http://localhost:21800"         | The URL of the dataset API
| BATCH_SIZE                 | 1000                             | The number of messages to process in each batch if the time out has not been reached
| BATCH_WAIT_TIME            | 200ms                            | The duration to wait before processing a partially full batch of messages
| ERROR_PRODUCER_TOPIC       | "report-events"                  | The Kafka topic to send the error messages to
| RESULT_PRODUCER_TOPIC      | "import-observations-inserted"   | The Kafka topic to send the observations inserted messages to
| CACHE_TTL                  | 60m                              | The amount of time to wait before clearing the cache (In minutes)
| GRACEFUL_SHUTDOWN_TIMEOUT  | "10s"                            | The shutdown timeout in seconds

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2017, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
