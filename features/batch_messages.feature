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
        Then observation "5,,sex,male,age,30" is inserted into the graph for instance ID "7"



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
        Then observation "5,,sex,male,age,30" is inserted into the graph for instance ID "7"
        And dimension key "7_10_someoption" is mapped to "111"
        And a message containing "7" is output