package database

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/mbh/himnario-back-end-go/internal/infra/config/property"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	once   sync.Once
	dbPool *DBPool
)

type DBPool struct {
	Pool *pgxpool.Pool
}

func NewPool(ctx context.Context, properties property.DatabaseSettings) *DBPool {
	once.Do(func() {
		log.Printf("connect to %s", properties.SafeAddr())
		if properties.Host == "" {
			log.Fatalf("database host is required")
		}

		poolCfg, err := pgxpool.ParseConfig(properties.AppDSN())
		if err != nil {
			log.Fatalf("parse pool config: %v", err)
		}

		poolCfg.MaxConns = int32(properties.MaxOpenConnections)
		poolCfg.MinConns = int32(properties.MinOpenConnections)
		poolCfg.MaxConnLifetime = time.Duration(properties.MaxConnLifetime) * time.Second
		poolCfg.MaxConnIdleTime = time.Duration(properties.MaxConnIdleTime) * time.Second
		poolCfg.HealthCheckPeriod = time.Duration(properties.HealthCheckPeriod) * time.Second
		cap := properties.StatementCacheCap
		if cap <= 0 {
			cap = 256
		}
		poolCfg.ConnConfig.StatementCacheCapacity = cap

		pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
		if err != nil {
			log.Fatalf("create pool: %v", err)
		}

		// Ping
		cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := pool.Ping(cctx); err != nil {
			pool.Close()
			log.Fatalf("db ping: %v", err)
		}
		dbPool = &DBPool{Pool: pool}
	})
	return dbPool
}
