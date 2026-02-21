package validations

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/mbh/himnario-back-end-go/internal/infra/config/validation"
	"github.com/mbh/himnario-back-end-go/internal/infra/primary/controllers/user/request"
)

type BodyAuthentication struct{}

func NewBodyAuthentication() *BodyAuthentication {
	return &BodyAuthentication{}
}

func (v *BodyAuthentication) Validate(dto *request.UserLogin) *validation.Result {
	err := validation.GetValidator().Struct(dto)
	validationResult := validation.Result{Success: true}
	if err != nil {
		errorsMessage := map[string]string{}
		var validationErrors validator.ValidationErrors

		errors.As(err, &validationErrors)

		for _, validationError := range validationErrors {
			field := validationError.Field()
			errorsMessage[field] = validationError.Error()
		}

		validationResult.Errors = errorsMessage
		validationResult.Success = false
	}

	return &validationResult
}
