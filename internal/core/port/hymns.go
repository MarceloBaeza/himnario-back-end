package port

import "github.com/mbh/himnario-back-end-go/internal/core/domain"

type PersistenceHymns interface {
	AddHymn(hymn *domain.Hymn) error
	GetHymnByID(id int) (*domain.Hymn, error)
	GetAllHymns() ([]*domain.Hymn, error)
}
