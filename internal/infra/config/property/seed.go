package property

import "sync"

var (
	seedPropertyOnce     sync.Once
	seedPropertyInstance *SeedProperty
)

func GetSeedProperty() *SeedProperty {
	seedPropertyOnce.Do(func() {
		seedPropertyInstance = &SeedProperty{}
	})
	return seedPropertyInstance
}

type SeedProperty struct {
	Seed SeedConfig `yaml:"seed"`
}

type SeedConfig struct {
	AdminEmail    string `yaml:"admin-email"`
	AdminName     string `yaml:"admin-name"`
	AdminPassword string `yaml:"admin-password"`
}
