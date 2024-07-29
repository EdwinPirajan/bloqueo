package config

import (
	"encoding/json"
	"os"
)

const CurrentVersion = "1.0.5"

type Config struct {
	UpdateCheckURL string `json:"update_check_url"`
}

func LoadConfig() (*Config, error) {
	file, err := os.Open("config.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
