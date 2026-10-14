package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"mixm-edu/backend/internal/apperror"
	"mixm-edu/backend/internal/auth"
	"mixm-edu/backend/internal/config"
	"mixm-edu/backend/internal/handler"
	"mixm-edu/backend/internal/model"
	"mixm-edu/backend/internal/repository"
)

func CORS(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}
	c.Next()
}
func Logger(c *gin.Context) {
	start := time.Now()
	c.Next()
	slog.Info("HTTP request", "method", c.Request.Method, "path", c.Request.URL.Path, "status", c.Writer.Status(), "duration", time.Since(start))
}
func Recovery(c *gin.Context) {
	defer func() {
		if recovered := recover(); recovered != nil {
			slog.Error("handler panic", "path", c.Request.URL.Path)
			handler.Failure(c, &apperror.Error{Status: http.StatusInternalServerError, Detail: "服务器错误"})
		}
	}()
	c.Next()
}
func Auth(config config.Config, db *repository.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := auth.Parse(config.Secret, strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "), "access")
		if err != nil {
			handler.Failure(c, &apperror.Error{Status: http.StatusUnauthorized, Detail: "登录已过期"})
			return
		}
		user, err := repository.ActiveUser(c.Request.Context(), db, claims.Subject)
		if err != nil {
			if !errors.Is(err, repository.ErrNotFound) {
				slog.Error("authentication lookup failed", "error", err)
			}
			handler.Failure(c, &apperror.Error{Status: http.StatusUnauthorized, Detail: "用户不可用"})
			return
		}
		c.Set("actor", model.Actor{ID: user.ID, Username: user.Username, IsSuperuser: user.IsSuperuser != nil && *user.IsSuperuser})
		c.Next()
	}
}
