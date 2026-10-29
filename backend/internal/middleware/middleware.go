package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"mixm/backend/internal/apperror"
	"mixm/backend/internal/auth"
	"mixm/backend/internal/config"
	"mixm/backend/internal/handler"
	"mixm/backend/internal/model"
	"mixm/backend/internal/repository"
)

// CORS 设置跨域响应头并处理预检请求。
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

// Logger 记录请求耗时、访问路径与响应状态。
func Logger(c *gin.Context) {
	start := time.Now()
	c.Next()
	slog.Info(
		"HTTP request",
		"method",
		c.Request.Method,
		"path",
		c.Request.URL.Path,
		"status",
		c.Writer.Status(),
		"duration",
		time.Since(start),
	)
}

// Recovery 捕获请求处理中的异常并返回统一错误响应。
func Recovery(c *gin.Context) {
	defer func() {
		if recovered := recover(); recovered != nil {
			slog.Error("handler panic", "path", c.Request.URL.Path)
			handler.Failure(c, &apperror.Error{Status: http.StatusInternalServerError, Detail: "服务器错误"})
		}
	}()

	c.Next()
}

// Auth 校验访问令牌并加载当前有效用户。
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

		c.Set(
			"actor",
			model.Actor{
				ID:          user.ID,
				Username:    user.Username,
				IsSuperuser: user.IsSuperuser != nil && *user.IsSuperuser,
			},
		)
		c.Next()
	}
}
