package middleware

import (
	"github.com/SZabrodskii/music-library/utils/providers"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func CacheMiddleware(cache *providers.CacheProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		cacheKey := generateCacheKey(c)
		if val, ok := cache.GetFromCache(cacheKey); ok {
			c.JSON(http.StatusOK, val)
			c.Abort()
			return
		}

		c.Next()

		if c.Writer.Status() == http.StatusOK {
			responseData := c.MustGet("responseData")
			cache.SetToCache(cacheKey, responseData, 5*time.Minute)
		}
	}
}

func generateCacheKey(c *gin.Context) string {
	return c.Request.Method + ":" + c.Request.URL.Path + ":" + c.Request.URL.RawQuery
}
