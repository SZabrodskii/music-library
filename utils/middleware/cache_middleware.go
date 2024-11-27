package middleware

import (
	"bytes"
	"github.com/SZabrodskii/music-library/utils/providers"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
)

type writer struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *writer) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func CacheMiddleware(cache *providers.CacheProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}
		cacheKey := generateCacheKey(c)
		if val, ok := cache.GetFromCache(cacheKey); ok {
			c.JSON(http.StatusOK, val)
			c.Abort()
			return
		}

		w := &writer{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = w
		c.Next()

		if c.Writer.Status() == http.StatusOK {
			cache.SetToCache(cacheKey, w.body.Bytes(), 10*time.Minute)
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
