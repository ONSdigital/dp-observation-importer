# dp-observation-importer

* Consumes a Kafka message for each observation to be imported.
* Query the dp-import-api for the dimension option database id's
* Insert the observation into the DB

## Getting started

* Run ```brew install kafka```
* Run ```brew install neo4j```
* Configure neo4j, edit /usr/local/Cellar/neo4j/3.2.0/libexec/conf/neo4j.conf
* Set ```dbms.security.auth_enabled=false```
* Run ```brew services restart neo4j```

## Kafka scripts

Scripts for updating and debugging Kafka can be found [here](https://github.com/ONSdigital/dp-data-tools)(dp-data-tools)

## Configuration

| Environment variable          | Default                              | Description
| ------------------------------|------------------------------------- |-----------------------------------------------------
| BIND_ADDR                     | :21700                               | The port to bind to
| DATASET_API_URL               | `http://localhost:21800`             | The URL of the dataset API
| SERVICE_AUTH_TOKEN            | AA78C45F-DD64-4631-BED9-FEAE29200620 | The service authorization token
| KAFKA_ADDR                    | `http://localhost:9092`              | The addresses of the Kafka instances (comma-separated)
| KAFKA_VERSION                 | `1.0.2`                              | The kafka version that this service expects to connect to
| KAFKA_OFFSET_OLDEST           | true                                 | sets the kafka offset to be oldest if true
| KAFKA_SEC_PROTO               | _unset_                              | if set to `TLS`, kafka connections will use TLS ([ref-1])
| KAFKA_SEC_CLIENT_KEY          | _unset_                              | PEM for the client key ([ref-1])
| KAFKA_SEC_CLIENT_CERT         | _unset_                              | PEM for the client certificate ([ref-1])
| KAFKA_SEC_CA_CERTS            | _unset_                              | CA cert chain for the server cert ([ref-1])
| KAFKA_SEC_SKIP_VERIFY         | false                                | ignores server certificate issues if `true` ([ref-1])
| OBSERVATION_CONSUMER_GROUP    | dp-observation-importer              | The Kafka consumer group to consume observation extracted events from
| OBSERVATION_CONSUMER_TOPIC    | observation-extracted                | The Kafka topic to consume observation extracted events from
| ERROR_PRODUCER_TOPIC          | report-events                        | The Kafka topic to send the error messages to
| RESULT_PRODUCER_TOPIC         | import-observations-inserted         | The Kafka topic to send the observations inserted messages to
| BATCH_SIZE                    | 100                                  | The number of messages to process in each batch if the time out has not been reached
| BATCH_WAIT_TIME               | 200ms                                | The duration to wait before processing a partially full batch of messages (time.Duration)
| KAFKA_MAX_BYTES               | 0                                    | The max message size for kafka producer
| CACHE_TTL                     | 60m                                  | The amount of time to wait before clearing the cache (time.Duration)
| GRACEFUL_SHUTDOWN_TIMEOUT     | 10s                                  | The shutdown timeout (time.Duration)
| HEALTHCHECK_INTERVAL          | 30s                                  | The period of time between health checks
| HEALTHCHECK_CRITICAL_TIMEOUT  | 90s                                  | The period of time after which failing checks will result in critical global check status
| GRAPH_DRIVER_TYPE             | neo4j                                | String identifier for the implementation to be used (e.g. 'neo4j', 'neptune' or 'mock')
| ENABLE_GET_GRAPH_DIMENSION_ID | true                                 | Use store ID's for Neptune search

[ref-1]:  https://github.com/ONSdigital/dp-kafka/tree/main/examples#tls 'kafka TLS examples documentation'

## Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

## License

Copyright © 2016-2021, [Office for National Statistics](https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
