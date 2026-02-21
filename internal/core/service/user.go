package service

import (
	"fmt"
	"strings"
	"sync"

	"github.com/mbh/himnario-back-end-go/internal/core/domain"
	"github.com/mbh/himnario-back-end-go/internal/core/port"
	"github.com/mbh/himnario-back-end-go/internal/infra/config/security"
)

var (
	onceUserService sync.Once
	userService     *UserService
)

type UserService struct {
	userPersistence port.PersistenceUser
}

func NewUserService(userPersistence port.PersistenceUser) *UserService {
	onceUserService.Do(func() {
		userService = &UserService{
			userPersistence: userPersistence,
		}
	})
	return userService
}

func (us *UserService) Create(newUser *domain.UserRegistry) error {
	hashPwd, err := security.GenerateHashPassword(newUser.Password)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}
	newUser.Password = hashPwd
	newUser.Email = strings.ToLower(strings.TrimSpace(newUser.Email))

	return us.userPersistence.Registry(newUser)
}

func (us *UserService) Login(user *domain.UserLogin) (*domain.User, error) {
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
	uResponse, pwddHasshed, err := us.userPersistence.Login(user.Email)
	if err != nil {
		return nil, err
	}
	err = security.CompareHashAndPassword(user.Password, *pwddHasshed)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	jwtToken, err := security.GenerateToken(uResponse)
	if err != nil {
		return nil, fmt.Errorf("error in authenticate: %w", err)
	}
	uResponse.Token = jwtToken
	return uResponse, nil
}
