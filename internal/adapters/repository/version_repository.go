package repository

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/EdwinPirajan/bloqueo.git/internal/core/domain"
)

// VersionRepository es la implementación de VersionChecker.
type VersionRepository struct {
	checkURL string
}

// NewVersionRepository crea una nueva instancia de VersionRepository.
func NewVersionRepository(checkURL string) *VersionRepository {
	return &VersionRepository{checkURL: checkURL}
}

// CheckForUpdates obtiene la información de la versión desde el servidor.
func (r *VersionRepository) CheckForUpdates() (*domain.VersionInfo, error) {
	apiURL := "http://10.96.16.68:8080/check"

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to check for updates: Status code %d", resp.StatusCode)
	}

	var versionInfo domain.VersionInfo
	if err := json.NewDecoder(resp.Body).Decode(&versionInfo); err != nil {
		return nil, fmt.Errorf("error decoding response body: %v", err)
	}

	return &versionInfo, nil
}
