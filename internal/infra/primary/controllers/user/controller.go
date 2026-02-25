package user

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mbh/himnario-back-end-go/internal/core"
	"github.com/mbh/himnario-back-end-go/internal/core/domain"
	"github.com/mbh/himnario-back-end-go/internal/infra/config/validation"
	"github.com/mbh/himnario-back-end-go/internal/infra/primary/controllers/user/request"
	"github.com/mbh/himnario-back-end-go/internal/infra/primary/controllers/user/validations"
)

type UserController struct {
	useCase core.UserUseCaseHandler
}

func NewUserController(useCase core.UserUseCaseHandler) *UserController {
	return &UserController{useCase: useCase}
}

func (rc *UserController) RunController(r *gin.Engine) {
	user := r.Group("/user/")
	user.Use(rc.validationsHeaders())
	user.POST("authentication", rc.Authentication)
	user.POST("create", rc.Create)
}

func (rc *UserController) Create(ctx *gin.Context) {
	var dto *request.UserRegistry
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, rc.GenerateResponseError(
			http.StatusBadRequest,
			"Invalid request body",
			&validation.Result{
				Success: false,
				Errors:  map[string]string{"error": err.Error()},
			}))
		return
	}
	validationResult := validations.NewBodyCreation().Validate(dto)
	if !validationResult.Success {
		ctx.JSON(http.StatusBadRequest, rc.GenerateResponseError(http.StatusBadRequest, "Invalid request body", validationResult))
		return
	}
	domainObj := MapperRegistry(dto)
	err := rc.useCase.Create(domainObj)
	if err != nil {
		if errors.Is(err, domain.ErrEmailExists) {
			ctx.JSON(http.StatusConflict, rc.GenerateResponseError(http.StatusConflict,
				"User creation failed", &validation.Result{
					Success: false,
					Errors:  map[string]string{"email": "email already exists"},
				}))
			return
		}
		log.Printf("user create error: %v", err)
		ctx.JSON(http.StatusInternalServerError, rc.GenerateResponseError(http.StatusInternalServerError,
			"User creation failed", &validation.Result{
				Success: false,
				Errors:  map[string]string{"error": "internal server error"},
			}))
		return
	}
	ctx.JSON(http.StatusOK, rc.GenerateReponseOk(nil, "User created successfully", http.StatusOK))
}

func (rc *UserController) Authentication(ctx *gin.Context) {
	var dto *request.UserLogin
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, rc.GenerateResponseError(http.StatusBadRequest, "Invalid request body",
			&validation.Result{
				Success: false,
				Errors:  map[string]string{"error": err.Error()},
			}))
		return
	}
	validationResult := validations.NewBodyAuthentication().Validate(dto)
	if !validationResult.Success {
		ctx.JSON(http.StatusBadRequest, rc.GenerateResponseError(http.StatusBadRequest, "Invalid request body", validationResult))
		return
	}
	domainObj := MapperLogin(dto)
	result, err := rc.useCase.Login(domainObj)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			ctx.JSON(http.StatusUnauthorized, rc.GenerateResponseError(http.StatusUnauthorized,
				"Authentication failed", &validation.Result{
					Success: false,
					Errors:  map[string]string{"authentication": "invalid credentials"},
				}))
			return
		}
		log.Printf("user authentication error: %v", err)
		ctx.JSON(http.StatusInternalServerError, rc.GenerateResponseError(http.StatusInternalServerError,
			"Authentication failed", &validation.Result{
				Success: false,
				Errors:  map[string]string{"error": "internal server error"},
			}))
		return
	}
	ctx.JSON(http.StatusOK, rc.GenerateReponseOk(result, "Authentication successful", http.StatusOK))
}

func (rc *UserController) GenerateReponseOk(data any, message string, statusCode int) domain.ResponseWrapper {
	return domain.ResponseWrapper{
		ResponseOk: &domain.ResponseOk{
			StatusCode: statusCode,
			Message:    message,
			Data:       data,
		},
	}
}

func (rc *UserController) validationsHeaders() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		validationResult := validation.RunValidations(ctx.Request.Header,
			validations.GetHeadersValidations())
		if !validationResult.Success {
			ctx.AbortWithStatusJSON(http.StatusBadRequest,
				rc.GenerateResponseError(http.StatusBadRequest, "Invalid headers",
					validationResult))
			return
		}
		ctx.Next()
	}
}

func (rc *UserController) GenerateResponseError(sts int, msg string, result *validation.Result) domain.ResponseWrapper {
	return domain.ResponseWrapper{
		ResponseError: &domain.ResponseError{
			StatusCode: sts,
			Error:      msg,
			Data:       result.Errors,
		},
	}
}
