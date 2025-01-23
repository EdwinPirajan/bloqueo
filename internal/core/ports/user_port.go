package ports

import "github.com/EdwinPirajan/bloqueo.git/internal/core/domain"

type APIClient interface {
	SendUser(user domain.User) error
}
