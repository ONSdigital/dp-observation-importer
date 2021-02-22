Feature: Batching messages from Kafka

    Scenario: Consuming an observation
        Given the following observation was extracted from ...
            """
            {InstanceID: "7", Row: "5,,sex,male,age,30"}
            """
        And for instance ID "7" the dataset api has headers "sex,age"
        When this observation is consumed
        # And the batching timeout limit has passed
        Then the following data is inserted into the graph
            """
            1,2,Helloworld
            """