package property

import (
	"fmt"
	"sync"
)

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
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	Name              string `yaml:"name"`
	User              string `yaml:"user"`
	Password          string `yaml:"password"`
	MigrationUser     string `yaml:"migration-user"`
	MigrationPassword string `yaml:"migration-password"`
	Driver            string `yaml:"driver"`
	SSLMode           string `yaml:"ssl-mode"`
	MaxOpenConnections  int    `yaml:"max-open-connections"`
	MinOpenConnections  int    `yaml:"min-open-connections"`
	MaxConnLifetime     int    `yaml:"max-conn-lifetime"`   // en segundos
	MaxConnIdleTime     int    `yaml:"max-conn-idle-time"`  // en segundos
	HealthCheckPeriod   int    `yaml:"health-check-period"` // en segundos
	StatementCacheCap   int    `yaml:"statement-cache-cap"`
	MigrationsDirectory string `yaml:"migrations-directory"`
}

func (d DatabaseSettings) AppDSN() string {
	return fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=%s", d.Driver, d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode)
}

func (d DatabaseSettings) MigrationDSN() string {
	return fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=%s", d.Driver, d.MigrationUser, d.MigrationPassword, d.Host, d.Port, d.Name, d.SSLMode)
}

func (d DatabaseSettings) SafeAddr() string {
	return fmt.Sprintf("%s:%d/%s", d.Host, d.Port, d.Name)
}
