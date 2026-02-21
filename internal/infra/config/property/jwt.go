package property

import "sync"

var (
	jwtPropertyOnce     sync.Once
	jwtPropertyInstance *JwtProperty
)

func GetJwtProperty() *JwtProperty {
	jwtPropertyOnce.Do(func() {
		jwtPropertyInstance = &JwtProperty{}
	})
	return jwtPropertyInstance
}

type JwtProperty struct {
	Jwt `yaml:"jwt"`
}
type Jwt struct {
	SecretKey      string `yaml:"secret-key"`
	ExpirationTime uint   `yaml:"expiration-time"`
}
