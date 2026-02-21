package port

import "github.com/mbh/himnario-back-end-go/internal/core/domain"

type PersistenceUser interface {
	Registry(user *domain.UserRegistry) error
	Login(email string) (*domain.User, *string, error)
}
