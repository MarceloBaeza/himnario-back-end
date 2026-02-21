package hymns

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mbh/himnario-back-end-go/internal/core"
	"github.com/mbh/himnario-back-end-go/internal/core/domain"
	"github.com/mbh/himnario-back-end-go/internal/infra/config/security"
	"github.com/mbh/himnario-back-end-go/internal/infra/config/validation"
	"github.com/mbh/himnario-back-end-go/internal/infra/primary/controllers/hymns/request"
	"github.com/mbh/himnario-back-end-go/internal/infra/primary/controllers/hymns/validations"
)

const JwtHeader = "Authorization"

type HymnController struct {
	hymnCase core.HymnUseCaseHandler
}

func NewHymnController(hymnCase core.HymnUseCaseHandler) *HymnController {
	return &HymnController{hymnCase: hymnCase}
}

func (rc *HymnController) RunController(r *gin.Engine) {
	user := r.Group("/hymn/")
	user.Use(rc.validationsHeaders())
	user.POST("create/", rc.Create)
	user.GET("all/", rc.GetAll)
	user.GET(":id/", rc.GetHymnByID)
}

func (rc *HymnController) Create(ctx *gin.Context) {
	var dto *request.NewHymn
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

	jwtToken, err := rc.getJwtToken(ctx.Request.Header)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, rc.GenerateResponseError(http.StatusUnauthorized,
			"Unauthorized", &validation.Result{
				Success: false,
				Errors:  map[string]string{"error": "authorization header missing or malformed"},
			}))
		return
	}
	userFromToken, err := security.ValidateToken(jwtToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, rc.GenerateResponseError(http.StatusUnauthorized,
			"Unauthorized", &validation.Result{
				Success: false,
				Errors:  map[string]string{"error": "invalid or expired token"},
			}))
		return
	}
	err = rc.ValidateUser(userFromToken, dto.User, "creation")
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, rc.GenerateResponseError(http.StatusUnauthorized,
			"Unauthorized", &validation.Result{
				Success: false,
				Errors:  map[string]string{"error": "unauthorized"},
			}))
		return
	}

	domainObj := MapperNewHymn(dto)
	err = rc.hymnCase.Create(domainObj)
	if err != nil {
		log.Printf("hymn create error: %v", err)
		ctx.JSON(http.StatusInternalServerError, rc.GenerateResponseError(http.StatusInternalServerError,
			"Hymn creation failed", &validation.Result{
				Success: false,
				Errors:  map[string]string{"error": "internal server error"},
			}))
		return
	}
	ctx.JSON(http.StatusOK, rc.GenerateReponseOk(nil, "Hymn created successfully", http.StatusOK))
}

func (rc *HymnController) GetAll(ctx *gin.Context) {
	hymns, err := rc.hymnCase.GetAllHymns()
	if err != nil {
		log.Printf("hymn get all error: %v", err)
		ctx.JSON(http.StatusInternalServerError, rc.GenerateResponseError(http.StatusInternalServerError,
			"Failed to retrieve hymns", &validation.Result{
				Success: false,
				Errors:  map[string]string{"error": "internal server error"},
			}))
		return
	}
	ctx.JSON(http.StatusOK, rc.GenerateReponseOk(hymns, "Get all hymns successfully",
		http.StatusOK))
}

func (rc *HymnController) GetHymnByID(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, rc.GenerateResponseError(http.StatusBadRequest,
			"Invalid hymn ID", &validation.Result{
				Success: false,
				Errors:  map[string]string{"error": err.Error()},
			}))
		return
	}

	hymn, err := rc.hymnCase.GetHymnByID(id)
	if err != nil {
		log.Printf("hymn get by id error: %v", err)
		ctx.JSON(http.StatusInternalServerError, rc.GenerateResponseError(http.StatusInternalServerError,
			"Failed to get hymn", &validation.Result{
				Success: false,
				Errors:  map[string]string{"error": "internal server error"},
			}))
		return
	}
	if hymn == nil {
		ctx.JSON(http.StatusNotFound, rc.GenerateResponseError(http.StatusNotFound,
			"Hymn not found", &validation.Result{
				Success: false,
				Errors:  map[string]string{"error": "hymn with given ID does not exist"},
			}))
		return
	}
	ctx.JSON(http.StatusOK, rc.GenerateReponseOk(hymn, "Get hymn successfully", http.StatusOK))
}

func (rc *HymnController) GenerateReponseOk(data any, message string, statusCode int) domain.ResponseWrapper {
	return domain.ResponseWrapper{
		ResponseOk: &domain.ResponseOk{
			StatusCode: statusCode,
			Message:    message,
			Data:       data,
		},
	}
}

func (rc *HymnController) ValidateUser(userFromToken *domain.User, userFromBody request.User,
	action string,
) error {
	if userFromToken.Email != userFromBody.Email {
		return errors.New("user email does not match token email")
	}
	if userFromToken.Name != userFromBody.Name {
		return errors.New("user name does not match token name")
	}
	switch action {
	case "creation":
		return domain.ValidateRolCreateEditHymns(userFromToken.Role)
	default:
		return errors.New("invalid action")
	}
}

func (rc *HymnController) validationsHeaders() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		validationResult := validation.RunValidations(ctx,
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

func (rc *HymnController) GenerateResponseError(sts int, msg string, result *validation.Result) domain.ResponseWrapper {
	return domain.ResponseWrapper{
		ResponseError: &domain.ResponseError{
			StatusCode: sts,
			Error:      msg,
			Data:       result.Errors,
		},
	}
}

func (rc *HymnController) getJwtToken(hdrs map[string][]string) (string, error) {
	authorizationToken, ok := hdrs[JwtHeader]
	if !ok || len(authorizationToken) == 0 {
		return "", errors.New("authorization header not found")
	}
	token := strings.Split(authorizationToken[0], " ")

	if len(token) != 2 {
		return "", errors.New("jwt token not sent")
	}
	return token[1], nil
}
