package handlers

import (
	"context"
	"errors"
	"github.com/SZabrodskii/music-library/song-service/services"
	"github.com/SZabrodskii/music-library/utils/models"
	"github.com/SZabrodskii/music-library/utils/providers"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"net/http"
	"os"
)

type SongHandler struct {
	cache   *providers.CacheProvider
	service *services.SongService
	logger  *zap.Logger
}

func NewSongHandler(cache *providers.CacheProvider, service *services.SongService, logger *zap.Logger) *SongHandler {
	return &SongHandler{
		cache:   cache,
		service: service,
		logger:  logger,
	}
}

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
func (h *SongHandler) GetSongs(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "10")
	filters := c.QueryArray("filters")

	request := &services.GetSongsRequest{
		Page:     page,
		PageSize: pageSize,
		Filters:  filters,
	}

	songs, err := h.service.GetSongs(request)
	if err != nil {
		h.logger.Error("Failed to get songs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Set("responseData", songs)
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
func (h *SongHandler) GetSongText(c *gin.Context) {
	songId := c.Param("songId")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "10")

	request := &services.GetSongTextRequest{
		SongId:   songId,
		Page:     page,
		PageSize: pageSize,
	}

	verses, err := h.service.GetSongText(request)
	if err != nil {
		h.logger.Error("Failed to get song text", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Set("responseData", verses)
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
func (h *SongHandler) DeleteSong(c *gin.Context) {
	songId := c.Param("songId")
	request := &services.DeleteSongRequest{
		SongId: songId,
	}

	if err := h.service.PublishToQueue("delete_song_queue", request); err != nil {
		h.logger.Error("Failed to publish delete song task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.cache.DeleteFromCache("song_" + songId)
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
func (h *SongHandler) UpdateSong(c *gin.Context) {
	songId := c.Param("songId")
	var song models.Song
	if err := c.ShouldBindJSON(&song); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	request := &services.UpdateSongRequest{
		SongID: songId,
		Song:   &song,
	}

	if err := h.service.PublishToQueue("update_song_queue", request); err != nil {
		h.logger.Error("Failed to publish update song task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.cache.DeleteFromCache("song_" + songId)
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
func (h *SongHandler) AddSong(c *gin.Context) {
	var song models.Song
	if err := c.ShouldBindJSON(&song); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	request := &services.AddSongRequest{
		Song: &song,
	}

	if err := h.service.AddSongToQueue(request); err != nil {
		h.logger.Error("Failed to add song to queue", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.cache.ClearCache()
	c.Status(http.StatusNoContent)
}

func RegisterHandlers(
	logger *zap.Logger,
	cache *providers.CacheProvider,
	songService *services.SongService,
	lifecycle fx.Lifecycle,
) *gin.Engine {
	handler := NewSongHandler(cache, songService, logger)
	r := gin.Default()

	r.GET("/songs", handler.GetSongs)
	r.GET("/songs/:songId/text", handler.GetSongText)
	r.DELETE("/songs/:songId", handler.DeleteSong)
	r.PATCH("/songs/:songId", handler.UpdateSong)
	r.POST("/songs", handler.AddSong)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			songService.RegisterConsumers()
			songService.ConsumerManager.StartConsumers()
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
		if err := r.Run(":" + port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	return r
}
