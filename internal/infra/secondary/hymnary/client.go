package hymnary

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mbh/himnario-back-end-go/internal/core/domain"
	"github.com/mbh/himnario-back-end-go/internal/infra/config/database"
	"github.com/mbh/himnario-back-end-go/internal/infra/config/property"
)

var (
	once   sync.Once
	client *Hymns
)

type Hymns struct {
	*database.DBPool
}

func NewHymns() *Hymns {
	once.Do(func() {
		ctx := context.Background()
		properties := property.GetDatabaseProperty().DatabaseSettings.Himnario
		client = &Hymns{
			DBPool: database.NewPool(ctx, properties),
		}
	})
	return client
}

func (c *Hymns) AddHymn(hymn *domain.Hymn) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	contentJson, err := json.Marshal(hymn.Content)
	if err != nil {
		return fmt.Errorf("error marshalling hymn content: %w", err)
	}
	// Llamar SP
	const q = `SELECT sp_create_hymn_by_author_email($1, $2::jsonb, $3);`
	_, err = c.Pool.Exec(ctx, q, hymn.Title, string(contentJson), hymn.EmailUser)
	if err != nil {
		return fmt.Errorf("error create hymn: %w", err)
	}
	log.Printf("Hymn created successfully: %s\n", hymn.Title)
	return nil
}

func (c *Hymns) GetHymnByID(id int) (*domain.Hymn, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `SELECT id, title, content, created_at FROM public.sp_get_hymn_by_id($1);`

	row := c.Pool.QueryRow(ctx, q, id)
	var item domain.Hymn
	err := row.Scan(
		&item.Id,
		&item.Title,
		&item.Content,
		&item.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr != nil {
			log.Printf("Hymn with ID %d, %s\n", id, pgErr.Message)
			if pgErr.Code == "P0001" {
				return nil, nil
			}
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("sp_get_hymn_by_id: %w", err)
	}
	return &item, nil
}

func (c *Hymns) GetAllHymns() ([]*domain.Hymn, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `SELECT id, title, created_at FROM public.sp_list_hymns();`

	rows, err := c.Pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("sp_list_hymns query: %w", err)
	}
	defer rows.Close()

	out := make([]*domain.Hymn, 0)

	for rows.Next() {
		var item domain.Hymn
		if err := rows.Scan(
			&item.Id,
			&item.Title,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("sp_list_hymns scan: %w", err)
		}
		out = append(out, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sp_list_hymns rows: %w", err)
	}
	return out, nil
}
