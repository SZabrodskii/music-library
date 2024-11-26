package middleware

import (
	"github.com/SZabrodskii/music-library/utils/providers"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
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
	query := c.Request.URL.Query()
	page := query.Get("page")
	pageSize := query.Get("pageSize")
	filters := query["filters"]
	filterString := strings.Join(filters, "_")
	return "songs_" + page + "_" + pageSize + "_" + filterString
}
