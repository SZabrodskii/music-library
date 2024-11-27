package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
)

func TraceParentMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Header.Get("traceparent") == "" {
			c.Request.Header.Set("traceparent", GenerateTraceparent())
		}
		c.Next()
	}
}

func generateTraceID() string {
	randomNumber := rand.Uint64()

	return fmt.Sprintf("%016x", randomNumber)
}

func generateSpanID() string {
	randomNumber := rand.Uint64()
	return fmt.Sprintf("%016x", randomNumber)
}

func generateTraceFlags() string {
	defer func() {
		requests++
	}()

	if requests%100 == 0 {
		return "01"
	}

	return "00"
}

var requests = 1

func GenerateTraceparent() string {
	version := "00"
	traceID := generateTraceID()
	spanID := generateSpanID()
	traceFlags := generateTraceFlags()

	return fmt.Sprintf("%s-%s-%s-%s", version, traceID, spanID, traceFlags)
}
