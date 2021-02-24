Feature: Batching messages from Kafka

    Scenario: Consuming one observation whose instance has only headers
        Given dataset instance "7" has headers "sex,age"
        And dataset instance "7" has no dimensions
        When these observations are consumed:
            | InstanceID | Row                |
            | 7          | 5,,sex,male,age,30 |
        Then these observations should be inserted into the database for batch "0":
            | InstanceID | Observation        |
            | 7          | 5,,sex,male,age,30 |


    Scenario: Consuming one observation whose instance has headers and one dimension
        Given dataset instance "7" has headers "sex,ddd"
        And dataset instance "7" has dimensions:
            | DimensionName | NodeID | Option     |
            | 10            | 111    | someoption |
        When these observations are consumed:
            | InstanceID | Row                |
            | 7          | 5,,sex,male,age,30 |
        Then these observations should be inserted into the database for batch "0":
            | InstanceID | Observation          |
            | 7          | 5,,sex,male,age,30 |
        And these dimensions should be inserted into the database for batch "0":
            | NodeID | Dimension       | 
            | 111    | 7_10_someoption |
        And a message stating "1" observation(s) inserted for instance ID "7" is sent