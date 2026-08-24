package middleware

import (
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{7,63}$`)

func RequestID(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader("X-Request-ID"))
		if requestID == "" {
			panic("missing request id")
		}
		if !requestIDPattern.MatchString(requestID) {
			panic("invalid request id")
		}
		if prefix != "" {
			requestID = prefix + "-" + requestID
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}
