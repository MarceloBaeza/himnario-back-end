package validations

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/mbh/himnario-back-end-go/internal/infra/config/validation"
)

var (
	validatables []validation.Validatable[*gin.Context]
	once         sync.Once
)

func GetHeadersValidations() []validation.Validatable[*gin.Context] {
	once.Do(func() {
		validatables = []validation.Validatable[*gin.Context]{
			NewHeadersValidator(),
		}
	})
	return validatables
}
