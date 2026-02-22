package seeder

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/mbh/himnario-back-end-go/internal/infra/config/database"
	"github.com/mbh/himnario-back-end-go/internal/infra/config/property"
	"github.com/mbh/himnario-back-end-go/internal/infra/config/security"
)

func SeedAdminUser(pool *database.DBPool, cfg property.SeedConfig) {
	if cfg.AdminPassword == "" {
		log.Println("seeder: ADMIN_INITIAL_PASSWORD not set, skipping admin seed")
		return
	}

	email := strings.ToLower(strings.TrimSpace(cfg.AdminEmail))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	const checkQ = `SELECT id FROM sp_get_user_auth_by_email($1);`
	var id int
	err := pool.Pool.QueryRow(ctx, checkQ, email).Scan(&id)
	if err == nil {
		log.Printf("seeder: admin user %s already exists, skipping\n", email)
		return
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		log.Printf("seeder: error checking admin user: %v\n", err)
		return
	}

	hash, err := security.GenerateHashPassword(cfg.AdminPassword)
	if err != nil {
		log.Fatalf("seeder: error hashing admin password: %v", err)
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	const createQ = `CALL sp_create_user($1, $2, $3, $4);`
	_, err = pool.Pool.Exec(ctx2, createQ, email, hash, cfg.AdminName, "admin")
	if err != nil {
		log.Fatalf("seeder: error creating admin user: %v", err)
	}
	log.Printf("seeder: admin user %s created successfully\n", email)
}
