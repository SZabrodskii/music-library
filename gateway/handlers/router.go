package handlers

import (
	"github.com/SZabrodskii/music-library/utils/middleware"
	"github.com/SZabrodskii/music-library/utils/providers"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

type Router struct {
	engine *gin.Engine
}

func NewRouter(logger *zap.Logger, cache *providers.CacheProvider, songHandler *SongHandler) *Router {
	router := gin.New()
	router.Use(middleware.TraceParentMiddleware())
	router.Use(gin.Recovery())
	router.Use(func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()
		duration := time.Since(start)
		logger.Info("Request completed",
			zap.String("traceparent", ctx.Request.Header.Get("traceparent")),
			zap.String("method", ctx.Request.Method),
			zap.String("path", ctx.Request.URL.Path),
			zap.Any("query", ctx.Request.URL.Query()),
			zap.Duration("duration", duration),
		)
	})
	router.Use(middleware.CacheMiddleware(cache))

	router.GET("/api/v1/songs", songHandler.GetSongs)
	router.GET("/api/v1/songs/:songId/text", songHandler.GetSongText)
	router.DELETE("/api/v1/songs/:songId", songHandler.DeleteSong)
	router.PATCH("/api/v1/songs/:songId", songHandler.UpdateSong)
	router.POST("/api/v1/songs", songHandler.AddSong)

	return &Router{engine: router}

}

func (r *Router) Start() {
	r.engine.Run(":8080")
}
