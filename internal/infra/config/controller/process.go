package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	limits "github.com/gin-contrib/size"
	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
	"github.com/mbh/himnario-back-end-go/internal/infra/config/property"
	"github.com/mbh/libraries/golang/lightms"
)

type ControllerRunnable interface {
	RunController(r *gin.Engine)
}

type GinController struct {
	controllers []ControllerRunnable
}

func GetControllerInstance(c []ControllerRunnable) lightms.PrimaryProcess {
	return &GinController{controllers: c}
}

func (c *GinController) Start() {
	srvConf := property.GetServerProperty()
	_, server := c.buildServer(srvConf)

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("error trying to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("server exited")
}

func (c *GinController) buildServer(srvConf *property.ServerProperty) (*gin.Engine, *http.Server) {
	gin.SetMode(srvConf.RunMode)
	r := gin.New()

	for _, o := range c.controllers {
		o.RunController(r)
	}

	middlewares(r, srvConf.MaxRequestTime, srvConf.MaxBodySize, srvConf.LimitByIp,
		AllowedPaths(srvConf.AllowPaths))

	port := fmt.Sprintf(":%v", srvConf.Port)
	srv := &http.Server{
		Addr:         port,
		Handler:      r,
		ReadTimeout:  time.Duration(srvConf.MaxRequestTime) * time.Second,
		WriteTimeout: time.Duration(srvConf.MaxRequestTime) * time.Second,
	}

	return r, srv
}

func AllowedPaths(paths string) map[string]bool {
	listPath := strings.Split(paths, ",")
	aPaths := make(map[string]bool, len(listPath))
	for i := range listPath {
		aPaths[listPath[i]] = true
	}
	return aPaths
}

func matchPath(pattern, path string) bool {
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")
	if len(patternParts) != len(pathParts) {
		return false
	}
	for i, part := range patternParts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			continue
		}
		if part != pathParts[i] {
			return false
		}
	}
	return true
}

func middlewares(srv *gin.Engine, timeOut, maxSizeRequest uint, limitByIp uint, allowedPaths map[string]bool) {
	srv.Use(configurationLogging())
	srv.Use(timeoutMiddleware(timeOut)) // time out request
	srv.Use(AllowedPathsMiddleware(allowedPaths))
	srv.Use(limits.RequestSizeLimiter(int64(maxSizeRequest * 1024 * 1024))) // bytes max size request
	srv.Use(RateLimit(limitByIp))                                           // max requests per seconds
	srv.Use(gin.Recovery())
	srv.Use(HeadersValidations())
}

func AllowedPathsMiddleware(allowedPaths map[string]bool) gin.HandlerFunc {
	patterns := make([]string, 0, len(allowedPaths))
	for p := range allowedPaths {
		patterns = append(patterns, p)
	}
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		for _, pattern := range patterns {
			if matchPath(pattern, path) {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "not found",
		})
	}
}

func HeadersValidations() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		c.Header("Referrer-Policy", "strict-origin")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Permissions-Policy", "geolocation=(),midi=(),sync-xhr=(),microphone=(),camera=(),magnetometer=(),gyroscope=(),fullscreen=(self),payment=()")
		c.Next()
	}
}

func RateLimit(requestBySeconds uint) gin.HandlerFunc {
	store := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  time.Second,
		Limit: requestBySeconds,
	})
	return ratelimit.RateLimiter(store, &ratelimit.Options{
		ErrorHandler: errorHandler,
		KeyFunc:      keyFunc,
	})
}

func configurationLogging() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		data, _ := json.Marshal(param.Request.Body)
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
			data,
		)
	})
}

func keyFunc(c *gin.Context) string {
	return c.ClientIP()
}

func errorHandler(c *gin.Context, info ratelimit.Info) {
	c.String(http.StatusTooManyRequests, "Too many requests. Try again in "+time.Until(info.ResetTime).String())
}

func timeoutResponse(c *gin.Context) {
	c.String(http.StatusRequestTimeout, "timeout")
}

func timeoutMiddleware(s uint) gin.HandlerFunc {
	return timeout.New(
		timeout.WithTimeout(time.Duration(s)*time.Second),
		timeout.WithResponse(timeoutResponse),
	)
}
