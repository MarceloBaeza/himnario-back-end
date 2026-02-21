package property

import "sync"

var (
	validationsPropertyOnce     sync.Once
	validationsPropertyInstance *ValidationsProperty
)

func GetValidationsProperty() *ValidationsProperty {
	validationsPropertyOnce.Do(func() {
		validationsPropertyInstance = &ValidationsProperty{}
	})
	return validationsPropertyInstance
}

type ValidationsProperty struct {
	Validations Validations `yaml:"validations"`
}
type Validations struct {
	Headers Headers `yaml:"headers"`
}
type Headers struct {
	XClient []string `yaml:"x-client"`
}
