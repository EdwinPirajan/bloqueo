package ports

import "launcher/internal/core/domain"

type CheckService interface {
	CheckForUpdates() (*domain.CheckResponse, error)
}
