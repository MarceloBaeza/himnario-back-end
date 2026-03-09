package service

import (
	"sync"

	"github.com/mbh/himnario-back-end-go/internal/core/domain"
	"github.com/mbh/himnario-back-end-go/internal/core/port"
)

var (
	onceHymnsService sync.Once
	hymnsService     *HymnsService
)

type HymnsService struct {
	hymnsPersistence port.PersistenceHymns
}

func NewHymnsService(hymnsPersistence port.PersistenceHymns) *HymnsService {
	onceHymnsService.Do(func() {
		hymnsService = &HymnsService{
			hymnsPersistence: hymnsPersistence,
		}
	})
	return hymnsService
}

func (hs *HymnsService) Create(newHymn *domain.Hymn) error {
	return hs.hymnsPersistence.AddHymn(newHymn)
}

func (hs *HymnsService) GetAllHymns() ([]*domain.Hymn, error) {
	return hs.hymnsPersistence.GetAllHymns()
}

func (hs *HymnsService) GetHymnByID(id int) (*domain.Hymn, error) {
	return hs.hymnsPersistence.GetHymnByID(id)
}

func (hs *HymnsService) EditHymnByID(hymn *domain.Hymn) (*domain.Hymn, error) {
	err := hs.hymnsPersistence.EditHymn(hymn)
	if err != nil {
		return nil, err
	}
	return hymn, nil
}
