package users

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mbh/himnario-back-end-go/internal/core/domain"
	"github.com/mbh/himnario-back-end-go/internal/infra/config/database"
	"github.com/mbh/himnario-back-end-go/internal/infra/config/property"
	"github.com/mbh/himnario-back-end-go/internal/infra/secondary/users/responses"
)

var (
	once   sync.Once
	client *Client
)

type Client struct {
	*database.DBPool
}

func NewClient() *Client {
	once.Do(func() {
		ctx := context.Background()
		properties := property.GetDatabaseProperty().DatabaseSettings.Himnario
		client = &Client{
			DBPool: database.NewPool(ctx, properties),
		}
	})
	return client
}

func (c *Client) Registry(user *domain.UserRegistry) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Llamar SP
	const q = `CALL sp_create_user($1, $2, $3, $4);`

	_, err := c.Pool.Exec(ctx, q, user.Email, string(user.Password), user.Name, user.Role)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr != nil && pgErr.Code == "P0001" {
			return domain.ErrEmailExists
		}
		return fmt.Errorf("error creation user: %w", err)
	}
	return nil
}

func (c *Client) Login(email string) (*domain.User, *string, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `SELECT id, email, password_hash, role, name FROM sp_get_user_auth_by_email($1);`

	var row responses.AuthRow
	err := c.DBPool.Pool.QueryRow(ctx, q, email).Scan(&row.ID, &row.Email, &row.PasswordHash, &row.Role, &row.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, domain.ErrInvalidCredentials
		}
		return nil, nil, fmt.Errorf("login query: %w", err)
	}

	return &domain.User{
		Email: row.Email,
		Role:  domain.UserRole(row.Role),
		Name:  row.Name,
	}, &row.PasswordHash, nil
}
