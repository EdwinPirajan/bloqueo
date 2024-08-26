package config

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

const CurrentVersion = "1.0.3"

// Config representa la configuración que obtendremos de la API
type Config struct {
	UpdateCheckURL string `json:"url"`
}

func LoadConfig() (*Config, error) {
	apiURL := "http://10.96.16.68:8080/check"

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch config: " + resp.Status)
	}

	var cfg Config
	if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
		return nil, err
	}

	log.Println("Loaded config:", cfg)

	return &cfg, nil
}
