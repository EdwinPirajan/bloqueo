package ports

import "launcher/internal/core/domain"

type FileStorage interface {
	LoadConfig() (*domain.Config, error)
}
