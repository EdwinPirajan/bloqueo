package repository

import (
	"encoding/json"
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
	resp, err := http.Get(r.checkURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var versionInfo domain.VersionInfo
	if err := json.NewDecoder(resp.Body).Decode(&versionInfo); err != nil {
		return nil, err
	}

	return &versionInfo, nil
}
