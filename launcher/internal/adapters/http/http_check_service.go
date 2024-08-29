package http

import (
	"encoding/json"
	"io"
	"net/http"

	"launcher/internal/core/domain"
)

type HTTPCheckService struct {
	url string
}

func NewHTTPCheckService(url string) *HTTPCheckService {
	return &HTTPCheckService{url: url}
}

func (hcs *HTTPCheckService) CheckForUpdates() (*domain.CheckResponse, error) {
	resp, err := http.Get(hcs.url)
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
