package core

import "github.com/mbh/himnario-back-end-go/internal/core/domain"

type UserUseCaseHandler interface {
	Create(newUser *domain.UserRegistry) error
	Login(user *domain.UserLogin) (*domain.User, error)
}
type HymnUseCaseHandler interface {
	Create(newHymn *domain.Hymn) error
	GetAllHymns() ([]*domain.Hymn, error)
	GetHymnByID(id int) (*domain.Hymn, error)
	EditHymnByID(hymn *domain.Hymn) (*domain.Hymn, error)
}
