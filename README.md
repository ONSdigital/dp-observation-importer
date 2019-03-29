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

### Kafka scripts

Scripts for updating and debugging Kafka can be found [here](https://github.com/ONSdigital/dp-data-tools)(dp-data-tools)

### Configuration

| Environment variable       | Default                          | Description
| ---------------------------|--------------------------------- |-----------------------------------------------------
| BIND_ADDR                  | :21700                               | The port to bind to
| KAFKA_ADDR                 | http://localhost:9092                | The address of the Kafka instance
| OBSERVATION_CONSUMER_GROUP | dp-observation-importer              | The Kafka consumer group to consume observation extracted events from
| OBSERVATION_CONSUMER_TOPIC | observation-extracted                | The Kafka topic to consume observation extracted events from
| DATASET_API_URL            | http://localhost:21800               | The URL of the dataset API
| BATCH_SIZE                 | 1000                                 | The number of messages to process in each batch if the time out has not been reached
| BATCH_WAIT_TIME            | 200ms                                | The duration to wait before processing a partially full batch of messages (time.Duration)
| ERROR_PRODUCER_TOPIC       | report-events                        | The Kafka topic to send the error messages to
| RESULT_PRODUCER_TOPIC      | import-observations-inserted         | The Kafka topic to send the observations inserted messages to
| CACHE_TTL                  | 60m                                  | The amount of time to wait before clearing the cache (time.Duration)
| GRACEFUL_SHUTDOWN_TIMEOUT  | 10s                                  | The shutdown timeout (time.Duration)
| SERVICE_AUTH_TOKEN         | AA78C45F-DD64-4631-BED9-FEAE29200620 | The service authorization token
| ZEBEDEE_URL                | http://localhost:8082                | The host name for Zebedee

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2017, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
