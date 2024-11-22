package handlers

import (
	"github.com/SZabrodskii/music-library/song-service/services"
	"github.com/SZabrodskii/music-library/utils/cache"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func NewRouter(cache *cache.Cache, service *services.SongService, logger *zap.Logger) *gin.Engine {
	router := gin.Default()

	router.GET("/api/v1/songs", func(c *gin.Context) {
		GetSongs(c, cache, service, logger)
	})
	router.GET("/api/v1/songs/:songId/text", func(c *gin.Context) {
		GetSongText(c, cache, service, logger)
	})
	router.DELETE("/api/v1/songs/:songId", func(c *gin.Context) {
		DeleteSong(c, cache, service, logger)
	})
	router.PATCH("/api/v1/songs/:songId", func(c *gin.Context) {
		UpdateSong(c, cache, service, logger)
	})
	router.POST("/api/v1/songs", func(c *gin.Context) {
		AddSong(c, cache, service, logger)
	})

	return router
}
