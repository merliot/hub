//go:build !tinygo

package device

import (
	"encoding/json"
	"reflect"
)

func (c Config) getPrettyJSON() []byte {
	json, err := json.MarshalIndent(&c, "", "\t")
	if err != nil {
		return []byte(err.Error())
	}
	return json
}

func (c Config) getParamsJSON() []byte {
	t := reflect.TypeOf(c.State)
	schema := structToSchema(t)
	json, err := json.MarshalIndent(schema, "", "\t")
	if err != nil {
		return []byte(err.Error())
	}
	return json
}
