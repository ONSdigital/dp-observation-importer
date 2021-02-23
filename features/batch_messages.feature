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
        And dataset instance "7" has no dimensions
        When these observations are consumed:
            | InstanceID | Row                |
            | 7          | 5,,sex,male,age,30 |
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
            | DimensionName | NodeID | Option     |
            | 10            | 111    | someoption |
        When these observations are consumed:
            | InstanceID | Row                |
            | 7          | 5,,sex,male,age,30 |
        Then observation "5,,sex,male,age,30" is inserted into the graph for instance ID "7"
        And dimension key "7_10_someoption" is mapped to "111"
        And a message stating "1" observation(s) inserted for instance ID "7" is sent