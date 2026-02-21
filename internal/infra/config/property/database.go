package property

import "sync"

var (
	databasePropertyOnce     sync.Once
	databasePropertyInstance *DatabaseProperty
)

func GetDatabaseProperty() *DatabaseProperty {
	databasePropertyOnce.Do(func() {
		databasePropertyInstance = &DatabaseProperty{}
	})
	return databasePropertyInstance
}

type DatabaseProperty struct {
	DatabaseSettings Databases `yaml:"database"`
}
type Databases struct {
	Himnario DatabaseSettings `yaml:"himnario"`
}
type DatabaseSettings struct {
	Url                 string `yaml:"url"`
	MaxOpenConnections  int    `yaml:"max-open-connections"`
	MinOpenConnections  int    `yaml:"min-open-connections"`
	MaxConnLifetime     int    `yaml:"max-conn-lifetime"`   // en segundos
	MaxConnIdleTime     int    `yaml:"max-conn-idle-time"`  // en segundos
	HealthCheckPeriod   int    `yaml:"health-check-period"` // en segundos
	StatementCacheCap   int    `yaml:"statement-cache-cap"`
	MigrationsDirectory string `yaml:"migrations-directory"`
}
