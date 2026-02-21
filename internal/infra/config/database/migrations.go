package database

import (
	"database/sql"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/mbh/himnario-back-end-go/internal/infra/config/property"

	_ "github.com/lib/pq"
)

func RunMigrations(props property.DatabaseSettings) {
	log.Printf("connect to %s\n", props.Url)
	sqlDB, err := sql.Open("postgres", props.Url)
	if err != nil {
		log.Fatalf("open sql db: %v", err)
	}
	defer sqlDB.Close()

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		log.Fatalf("postgres driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		props.MigrationsDirectory,
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatalf("migrate init: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migrate up: %v", err)
	}
	log.Printf("migrations ok %s, database %s\n", props.MigrationsDirectory, props.Url)
}
