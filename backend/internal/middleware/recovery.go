package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func Recovery(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				panicText := recovered.(error).Error()
				log.Error("request panic recovered", "request_id", c.GetString("request_id"), "method", c.Request.Method, "path", c.Request.URL.Path, "panic", panicText, "stack", string(debug.Stack()))
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL_ERROR", "message": "服务暂时无法完成请求"}, "request_id": c.GetString("request_id")})
			}
		}()
		c.Next()
	}
}
