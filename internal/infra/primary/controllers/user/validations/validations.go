package validations

import (
	"net/http"
	"sync"

	"github.com/mbh/himnario-back-end-go/internal/infra/config/validation"
)

var (
	validatables []validation.Validatable[http.Header]
	once         sync.Once
)

func GetHeadersValidations() []validation.Validatable[http.Header] {
	once.Do(func() {
		validatables = []validation.Validatable[http.Header]{
			NewHeadersValidator(),
		}
	})
	return validatables
}
