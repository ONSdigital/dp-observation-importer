Feature: Batching messages from Kafka

    Scenario: Consuming one observation whose instance has only headers
        Given dataset instance "7" has headers:
            """
            {
                "Headers": [
                    "sex",
                    "age"
                ]
            }
            """
        And dataset instance "7" has dimensions:
            """
            {
                "items": []
            }
            """
        When this observation is consumed:
            """
            {InstanceID: "7", Row: "5,,sex,male,age,30"}
            """
        Then the following data is inserted into the graph for instance ID "7":
            """
            5,,sex,male,age,30
            """


    Scenario: Consuming one observation whose instance has headers and one dimension
        Given dataset instance "7" has headers:
            """
            {
                "Headers": [
                    "sex",
                    "age"
                ]
            }
            """
        And dataset instance "7" has dimensions:
            """
            {
                "items": [
                    {
                        "dimension": "10",
                        "node_id": "111",
                        "option": "someoption"
                    }
                ]
            }
            """
        When this observation is consumed:
            """
            {InstanceID: "7", Row: "5,,sex,male,age,30"}
            """
        Then the following data is inserted into the graph for instance ID "7":
            """
            5,,sex,male,age,30
            """