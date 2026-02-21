package user

import (
	"time"

	"github.com/mbh/himnario-back-end-go/internal/core/domain"
	"github.com/mbh/himnario-back-end-go/internal/infra/primary/controllers/user/request"
)

func MapperLogin(rq *request.UserLogin) *domain.UserLogin {
	return &domain.UserLogin{
		Email:    rq.Email,
		Password: rq.Password,
	}
}

func MapperRegistry(rq *request.UserRegistry) *domain.UserRegistry {
	return &domain.UserRegistry{
		Email:     rq.Email,
		Password:  rq.Password,
		Name:      rq.Name,
		Role:      domain.UserRole(rq.Role),
		CreatedAt: time.Now(),
	}
}
