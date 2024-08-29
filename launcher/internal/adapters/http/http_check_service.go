package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"launcher/internal/core/domain"
)

type HTTPCheckService struct {
	baseURL string
}

func NewHTTPCheckService(baseURL string) *HTTPCheckService {
	return &HTTPCheckService{baseURL: baseURL}
}

// CheckForUpdates implementa la interfaz CheckService
func (hcs *HTTPCheckService) CheckForUpdates(client string) (*domain.CheckResponse, error) {
	// Construir la URL con el parámetro del cliente
	url := fmt.Sprintf("%s?client=%s", hcs.baseURL, client)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var checkResponse domain.CheckResponse
	if err := json.Unmarshal(body, &checkResponse); err != nil {
		return nil, err
	}

	return &checkResponse, nil
}
