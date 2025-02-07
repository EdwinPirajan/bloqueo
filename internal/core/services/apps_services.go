package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type ConfigResponse struct {
	ProcessesToMonitor []string `json:"processes"`
	UrlsToBlock        []string `json:"urls"`
}

func FetchConfiguration(cliente string) (ConfigResponse, error) {
	var config ConfigResponse
	url := fmt.Sprintf("http://localhost:8080/apps/%s", cliente)

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

	fmt.Println("Configuración obtenida correctamente", config)

	return config, nil
}
