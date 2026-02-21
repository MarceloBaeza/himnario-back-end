package hymns

import (
	"github.com/mbh/himnario-back-end-go/internal/core/domain"
	"github.com/mbh/himnario-back-end-go/internal/infra/primary/controllers/hymns/request"
)

func MapperNewHymn(rq *request.NewHymn) *domain.Hymn {
	return &domain.Hymn{
		Title:     rq.Title,
		Content:   rq.Content,
		EmailUser: rq.User.Email,
	}
}
