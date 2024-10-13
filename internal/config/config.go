/*
Copyright © 2024 The Dues Authors
*/
package config

import (
	"encoding/json"
	"os"
)

// Reads configuration file and populates User Config
func ReadConfigFile(path string, config *UserConfig) error {
	jsonFile, err := os.Open(path)

	if err != nil {
		return err
	}
	defer jsonFile.Close()

	err = json.NewDecoder(jsonFile).Decode(config)
	if err != nil {
		return err
	}
	return nil
}
