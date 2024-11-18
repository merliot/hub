//go:build !tinygo

package device

import (
	"encoding/json"
	"io/ioutil"
)

func fileWriteJSON(name string, payload any) error {
	data, err := json.MarshalIndent(payload, "", "\t")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(name, data, 0644)
}

func fileReadJSON(name string, payload any) error {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &payload)
	if err != nil {
		return err
	}
	return nil
}
