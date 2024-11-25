package handlers

import (
	"github.com/SZabrodskii/music-library/song-service/services"
	"github.com/SZabrodskii/music-library/utils/cache"
	"github.com/SZabrodskii/music-library/utils/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Router struct {
	engine *gin.Engine
}

func NewRouter(cache *cache.Cache, client *services.Client, logger *zap.Logger, songHandler *SongHandler) *Router {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(func(ctx *gin.Context) {

		ctx.Next()
		logger.Info("New request", zap.String("method", ctx.Request.Method), zap.String("path", ctx.Request.URL.Path), zap.Any("query", ctx.Request.URL.Query()))
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
