package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type ConfigResponse struct {
	ProcessesToMonitor []string `json:"processes_to_monitor"`
	UrlsToBlock        []string `json:"urls_to_block"`
}

func FetchConfiguration(cliente string) (ConfigResponse, error) {
	var config ConfigResponse
	url := fmt.Sprintf("http://10.96.16.68:8080/apps?client=%s", cliente)

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return config, fmt.Errorf("error realizando la petición GET: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return config, fmt.Errorf("error en la respuesta del servidor: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return config, fmt.Errorf("error leyendo la respuesta del servidor: %v", err)
	}

	err = json.Unmarshal(body, &config)
	if err != nil {
		return config, fmt.Errorf("error deserializando la respuesta JSON: %v", err)
	}

	return config, nil
}
