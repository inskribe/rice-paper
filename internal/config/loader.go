package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

var AppConfig Config

func LoadApplicationConfig() error {
	data, err := os.ReadFile("application_config.yml")
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(data, &AppConfig); err != nil {
		return err
	}

	return nil
}
