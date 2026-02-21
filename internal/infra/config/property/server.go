package property

import "sync"

var (
	serverPropertyOnce     sync.Once
	serverPropertyInstance *ServerProperty
)

func GetServerProperty() *ServerProperty {
	serverPropertyOnce.Do(func() {
		serverPropertyInstance = &ServerProperty{}
	})
	return serverPropertyInstance
}

type ServerProperty struct {
	Server `yaml:"server"`
}
type Server struct {
	Port           uint   `yaml:"port"`
	MaxBodySize    uint   `yaml:"max-body-size"`
	MaxRequestTime uint   `yaml:"max-request-time"`
	LimitByIp      uint   `yaml:"limit-by-ip"`
	RunMode        string `yaml:"run-mode"`
	AllowPaths     string `yaml:"allow-paths"`
}
