package handlers

import (
	"context"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"music-library/song-service/models"
	"music-library/song-service/services"
	"music-library/utils/cache"
	"net/http"
	"os"
	"strings"
	"time"
)

// GetSongs godoc
// @Summary Get songs with filtering and pagination
// @Description Get songs with filtering and pagination
// @Tags songs
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Limit per page" default(10)
// @Param filters query []string false "Filters"
// @Success 200 {array} models.Song
// @Router /songs [get]
func GetSongs(c *gin.Context, cache *cache.Cache, service *services.SongService, logger *zap.Logger) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "10")
	filters := c.QueryArray("filters")

	cacheKey := "songs_" + page + "_" + pageSize + "_" + strings.Join(filters, "_")
	if val, ok := cache.GetFromCache(cacheKey); ok {
		c.JSON(http.StatusOK, val)
		return
	}

	songs, err := service.GetSongs(page, pageSize, filters)
	if err != nil {
		logger.Error("Failed to get songs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cache.SetToCache(cacheKey, songs, time.Minute*5)
	c.JSON(http.StatusOK, songs)
}

// GetSongText godoc
// @Summary Get song text with pagination by verses
// @Description Get song text with pagination by verses
// @Tags songs
// @Accept json
// @Produce json
// @Param songId path int true "Song ID"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Limit per page" default(10)
// @Success 200 {array} models.Verse
// @Router /songs/{songId}/text [get]
func GetSongText(c *gin.Context, cache *cache.Cache, service *services.SongService, logger *zap.Logger) {
	songId := c.Param("songId")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "10")

	cacheKey := "song_text_" + songId + "_" + page + "_" + pageSize
	if val, ok := cache.GetFromCache(cacheKey); ok {
		c.JSON(http.StatusOK, val)
		return
	}

	verses, err := service.GetSongText(songId, page, pageSize)
	if err != nil {
		logger.Error("Failed to get song text", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cache.SetToCache(cacheKey, verses, time.Minute*5)
	c.JSON(http.StatusOK, verses)
}

// DeleteSong godoc
// @Summary Delete a song
// @Description Delete a song by ID
// @Tags songs
// @Accept json
// @Produce json
// @Param songId path int true "Song ID"
// @Success 204
// @Router /songs/{songId} [delete]
func DeleteSong(c *gin.Context, cache *cache.Cache, service *services.SongService, logger *zap.Logger) {
	songId := c.Param("songId")

	if err := service.DeleteSong(songId); err != nil {
		logger.Error("Failed to delete song", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	cache.DeleteFromCache("song_" + songId)
	c.Status(http.StatusNoContent)
}

// UpdateSong godoc
// @Summary Update a song
// @Description Update a song by ID
// @Tags songs
// @Accept json
// @Produce json
// @Param songId path int true "Song ID"
// @Param song body models.Song true "Song data"
// @Success 200 {object} models.Song
// @Router /songs/{songId} [patch]
func UpdateSong(c *gin.Context, cache *cache.Cache, service *services.SongService, logger *zap.Logger) {
	songId := c.Param("songId")
	var song models.Song
	if err := c.ShouldBindJSON(&song); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.UpdateSong(songId, &song); err != nil {
		logger.Error("Failed to update song", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cache.DeleteFromCache("song_" + songId)
	c.JSON(http.StatusOK, song)
}

// AddSong godoc
// @Summary Add a new song
// @Description Add a new song
// @Tags songs
// @Accept json
// @Produce json
// @Param song body models.Song true "Song data"
// @Success 204
// @Router /songs [post]
func AddSong(c *gin.Context, cache *cache.Cache, service *services.SongService, logger *zap.Logger) {
	var song models.Song
	if err := c.ShouldBindJSON(&song); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.AddSongToQueue(&song); err != nil {
		logger.Error("Failed to add song to queue", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cache.ClearCache()
	c.Status(http.StatusNoContent)
}

func RegisterHandlers(
	logger *zap.Logger,
	cache *cache.Cache,
	songService *services.SongService,
	lifecycle fx.Lifecycle,
) {
	r := gin.Default()

	r.GET("/songs", func(c *gin.Context) {
		GetSongs(c, cache, songService, logger)
	})
	r.GET("/songs/:songId/text", func(c *gin.Context) {
		GetSongText(c, cache, songService, logger)
	})
	r.DELETE("/songs/:songId", func(c *gin.Context) {
		DeleteSong(c, cache, songService, logger)
	})
	r.PATCH("/songs/:songId", func(c *gin.Context) {
		UpdateSong(c, cache, songService, logger)
	})
	r.POST("/songs", func(c *gin.Context) {
		AddSong(c, cache, songService, logger)
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go songService.ConsumeSongQueue()
			return nil
		},
		OnStop: func(context.Context) error {
			return nil
		},
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	go func() {
		if err := r.Run(":" + port); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()
}
