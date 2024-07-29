package ports

import "github.com/EdwinPirajan/bloqueo.git/internal/core/domain"

// VersionChecker es la interfaz para verificar versiones.
type VersionChecker interface {
	CheckForUpdates() (*domain.VersionInfo, error)
}
