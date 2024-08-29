package ports

import "launcher/internal/core/domain"

type CheckService interface {
	CheckForUpdates(client string) (*domain.CheckResponse, error)
}
