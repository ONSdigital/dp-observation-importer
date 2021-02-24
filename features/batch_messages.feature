Feature: Batching messages from Kafka

    Scenario: Consuming one observation whose instance has only headrs
        Given dataset instance "7" has headers "V4_1,H1,H2,Time,,Code,,Age"
        And dataset instance "7" has no dimensions
        When these observations are consumed:
            | InstanceID | Row                              |
            | 7          | 128,,Month,Aug-16,K02000001,0,29 |
        Then these observations should be inserted into the database for batch "0":
        """
            [
                {
                    "Row": "128,,Month,Aug-16,K02000001,0,29",
                    "RowIndex": 0,
                    "InstanceID": "7",
                    "DimensionOptions": [
                        {"DimensionName":"Time", "Name":"Aug-16"},
                        {"DimensionName":"Code", "Name":"K02000001"},
                        {"DimensionName":"Age", "Name":"29"}
                    ]
                }
            ]
        """
        And a message stating "1" observation(s) inserted for instance ID "7" is sent

    Scenario: Consuming one observation whose instance has several dimensions
        Given dataset instance "7" has headers "V1,Code,,Age"
        And dataset instance "7" has dimensions:
            | DimensionName | NodeID | Option |
            | age           | 111    | 29     |
            | sex           | 111    | male   |
        When these observations are consumed:
            | InstanceID | Row                  |
            | 7          | 5,AK101,29,male,30,5 |
        Then these observations should be inserted into the database for batch "0":
        """
            [
                {
                    "Row": "5,AK101,29,male,30,5",
                    "RowIndex": 0,
                    "InstanceID": "7",
                    "DimensionOptions": [
                        {"DimensionName":"Code", "Name":"AK101"},
                        {"DimensionName":"Age", "Name":"29"}
                    ]
                }
            ]
        """
        And these dimensions should be inserted into the database for batch "0":
            | NodeID | Dimension  |
            | 111    | 7_age_29   |
            | 111    | 7_sex_male |
        And a message stating "1" observation(s) inserted for instance ID "7" is sent