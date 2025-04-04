//go:build !tinygo

package device

import (
	"encoding/json"
	"os"
)

func fileWriteJSON(name string, payload any) error {
	data, err := json.MarshalIndent(payload, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(name, data, 0644)
}

func fileReadJSON(name string, payload any) error {
	data, err := os.ReadFile(name)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &payload)
	if err != nil {
		return err
	}
	return nil
}
