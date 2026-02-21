package validation

type Validatable[T any] interface {
	Validate(t T) *Result
}

type Result struct {
	Success bool
	Errors  map[string]string // Key is the field name, value is the error message
}

func (v *Result) AddErrors(errors map[string]string) {
	if v.Errors == nil {
		v.Errors = make(map[string]string)
	}
	for field, errorMessage := range errors {
		v.Errors[field] = errorMessage
	}
}

func CreateValid() *Result {
	return &Result{Success: true, Errors: nil}
}

func CreateInvalid(field, errorMessage string) *Result {
	return &Result{
		Success: false,
		Errors:  map[string]string{field: errorMessage},
	}
}

func RunValidations[T any](toValidate T, validations []Validatable[T]) *Result {
	validationResult := &Result{Success: true}
	for _, validator := range validations {
		result := validator.Validate(toValidate)
		if !result.Success {
			validationResult.Success = false
			validationResult.AddErrors(result.Errors)
		}
	}

	return validationResult
}
