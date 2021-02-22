Feature: Batching messages from Kafka

    Scenario: Consuming an observation
        Given for instance ID "7" the dataset api has headers
            """
            {
                "Headers": [
                    "sex",
                    "age"
                ]
            }
            """
        When this observation is consumed:
            """
            {InstanceID: "7", Row: "5,,sex,male,age,30"}
            """
        # And the batching timeout limit has passed
        Then the following data is inserted into the graph
            """
            5,,sex,male,age,30
            """