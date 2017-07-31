package observationtest

import "github.com/johnnadratowski/golang-neo4j-bolt-driver/errors"

type DimensionStore struct {
	ReturnError bool
}


func (ds *DimensionStore) GetOrder(instanceID string) ([]string, error) {
  if ds.ReturnError {
	  return []string{}, errors.New("Failed to fetch csv header")
  }
	return []string{"V4", "data_markings", "sex_codelist", "sex"}, nil
}

func (ds *DimensionStore) GetNodeIDs(instanceID string) (map[string]string, error) {
	nodeMap:= make(map[string]string)
	if ds.ReturnError {
		return nodeMap, errors.New("Failed to fetch node ids")
	}
	nodeMap["3_age_30"] = "123"
	nodeMap["3_sex_female"] = "321"
	return nodeMap, nil

}
