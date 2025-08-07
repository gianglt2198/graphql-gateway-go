package utils

import (
	"encoding/json"

	"github.com/jinzhu/copier"
)

func StructToStruct[T, U any](t *T) (*U, error) {
	var u U
	b, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &u)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func ConvertTo[T any, M any](entity *T) (*M, error) {
	if entity == nil {
		return nil, nil
	}

	var modelObj M
	if err := copier.Copy(&modelObj, entity); err != nil {
		return nil, err
	}

	return &modelObj, nil
}
